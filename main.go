package main

import (
	"bufio"
	"image/color"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/creack/pty"
)



// TODO: Implement terminal layout for dynamic sizing
type termLayout struct {
	onResize func(fyne.Size)
}

func (l *termLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) > 0 {
		objects[0].Resize(size)
	}
	if l.onResize != nil {
		l.onResize(size)
	}
}

func (l *termLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(200, 200)
}

// Terminal struct and interfaces
type Terminal struct {
	Grid     [][]widget.TextGridCell
	MainGrid [][]widget.TextGridCell
	AltGrid  [][]widget.TextGridCell
	Cols     int
	Rows     int
	CursorX  int
	CursorY  int
	SavedX   int
	SavedY   int
	IsAlt    bool
	mu       sync.Mutex
}

func (t *Terminal) Scroll() {
	for y := 1; y < t.Rows; y++ {
		t.Grid[y-1] = t.Grid[y]
	}
	t.Grid[t.Rows-1] = make([]widget.TextGridCell, t.Cols)
	for x := range t.Cols {
		t.Grid[t.Rows-1][x] = widget.TextGridCell{Rune: ' '}
	}
}

func (t *Terminal) Resize(newCols, newRows int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if newCols == t.Cols && newRows == t.Rows {
		return
	}

	newMain := make([][]widget.TextGridCell, newRows)
	newAlt := make([][]widget.TextGridCell, newRows)
	for y := range newRows {
		newMain[y] = make([]widget.TextGridCell, newCols)
		newAlt[y] = make([]widget.TextGridCell, newCols)
		for x := range newCols {
			if y < t.Rows && x < t.Cols {
				if t.MainGrid != nil { newMain[y][x] = t.MainGrid[y][x] }
				if t.AltGrid != nil { newAlt[y][x] = t.AltGrid[y][x] }
			} else {
				newMain[y][x] = widget.TextGridCell{Rune: ' '}
				newAlt[y][x] = widget.TextGridCell{Rune: ' '}
			}
		}
	}

	t.MainGrid = newMain
	t.AltGrid = newAlt
	t.Cols = newCols
	t.Rows = newRows

	if t.IsAlt {
		t.Grid = t.AltGrid
	} else {
		t.Grid = t.MainGrid
	}

	if t.CursorX >= t.Cols {
		t.CursorX = t.Cols - 1
	}
	if t.CursorY >= t.Rows {
		t.CursorY = t.Rows - 1
	}
}

// CSI sequence helper functions
func getCSIParam(params string, defaultVal int) int {
	if params == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(params)
	if err != nil {
		return defaultVal
	}
	return val
}

func getCSICoords(params string) (int, int) {
	if params == "" {
		return 1, 1 // default to 1, 1
	}
	parts := strings.Split(params, ";")
	row := getCSIParam(parts[0], 1)
	col := 1
	if len(parts) > 1 {
		col = getCSIParam(parts[1], 1)
	}
	return row, col
}

// --- Main ---
func main() {
	a := app.New()
	w := a.NewWindow("shfte")

	ui := widget.NewTextGrid() // Create a new TextGrid
	term := &Terminal{ 
		CursorX: 0, 
		CursorY: 0,
	}

	term.Resize(80, 24)

	// Start pty
	os.Setenv("TERM", "xterm-256color")
	c := exec.Command("/bin/zsh")
	p, err := pty.Start(c)

	if err != nil {
		fyne.LogError("Failed to open pty", err)
		os.Exit(1)
	}

	defer c.Process.Kill()

	charSize := widget.NewTextGridFromString("X").MinSize()

	tl := &termLayout{
		onResize: func(s fyne.Size) {
			if charSize.Width <= 0 || charSize.Height <= 0 {
				return
			}

			newCols := int(s.Width / charSize.Width)
			newRows := int(s.Height / charSize.Height)

			if newCols < 1 { newCols = 1 }
			if newRows < 1 { newRows = 1 }

			term.Resize(newCols, newRows)

			_ = pty.Setsize(p, &pty.Winsize{
				Rows: uint16(newRows),
				Cols: uint16(newCols),
			})
		},
	}

	// Callback function that handles special keypresses
	onTypedKey := func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEnter || e.Name == fyne.KeyReturn {
			_, _ = p.Write([]byte{'\r'})
		} else if e.Name == fyne.KeyBackspace {
			_, _ = p.Write([]byte{'\x7f'})
		}
	}

	// Callback function that handles character keypresses
	onTypedRune := func(r rune) {
		_, _ = p.WriteString(string(r))
	}

	w.Canvas().SetOnTypedKey(onTypedKey)
	w.Canvas().SetOnTypedRune(onTypedRune)

	// Instead of runes, we now store TextGridCells which hold a rune AND a style
	reader := bufio.NewReader(p)

	// grid styles
	// TODO: figure out cursor styling

	// cursorStyle := &widget.CustomTextGridStyle{
	// 	FGColor: color.White,
	// 	BGColor: color.Opaque,
	// }

	// Define our default style and current active style
	defaultStyle := &widget.CustomTextGridStyle{
		FGColor: color.White,
		BGColor: color.Transparent,
	}
	currentStyle := defaultStyle

	// Goroutine that reads from pty and parses ANSI
	go func() {
		inEscape := false // Are we currently parsing an escape sequence?
		inCSI := false    // Are we in a Control Sequence Introducer (CSI)?
		csiParams := ""   // The numbers inside the sequence (e.g., "31;1")
		isPrivate := false // DEC Private mode


		for {
			r, _, err := reader.ReadRune()

			if err != nil {
				if err == io.EOF {
					a.Quit()
					return
				}
				os.Exit(0)
			}

			term.mu.Lock()

			// --- ANSI ESCAPE PARSER ---
			if r == '\x1b' {
				inEscape = true
				term.mu.Unlock()
				continue
			}
			if inEscape {
				if r == '[' {
					inCSI = true
					inEscape = false
					csiParams = ""
					term.mu.Unlock()
					continue
				}
				inEscape = false // Unsupported escape, abort
			}
			if inCSI {
				// catch DEC private mode
				if r == '?' {
					isPrivate = true
					term.mu.Unlock()
					continue
				}
				// If it's a number or semicolon, add it to our parameters
				if (r >= '0' && r <= '9') || r == ';' {
					csiParams += string(r)
					term.mu.Unlock()
					continue
				}

				// Any other letter terminates the CSI sequence (cursor moves, clears, etc.)
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {

					// Private Mode Commands
					if isPrivate {
						switch r {
						case 'h': // Set mode
							if getCSIParam(csiParams, 0) == 1049 {
								// save cursor position
								term.SavedX, term.SavedY = term.CursorX, term.CursorY
								// switch to alt grid
								term.IsAlt = true
								term.Grid = term.AltGrid
								// clear alt grid
								for y := 0; y < term.Rows; y++ {
									for x := 0; x < term.Cols; x++ {
										term.Grid[y][x] = widget.TextGridCell{Rune: ' '}
									}
								}
								term.CursorX, term.CursorY = 0, 0
							}
						case 'l': // Reset mode
							if getCSIParam(csiParams, 0) == 1049 {
								term.IsAlt = false
								term.Grid = term.MainGrid
								term.CursorX, term.CursorY = term.SavedX, term.SavedY
							}
						}

						inCSI = false
						isPrivate = false
						term.mu.Unlock()
						continue
					}

					// Standard commands
					switch r {
						case 'm':
							codes := strings.Split(csiParams, ";")
							for i := 0; i < len(codes); i++ {
								code, _ := strconv.Atoi(codes[i])
								switch code {
								case 0, 39, 49: // Reset
									currentStyle = defaultStyle
								case 31: // Red
									currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{255, 0, 0, 255}}
								case 32: // Green
									currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{0, 255, 0, 255}}
								case 33: // Yellow
									currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{255, 255, 0, 255}}
								case 34: // Blue
									currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{0, 100, 255, 255}}
								case 35: // Magenta
									currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{255, 0, 255, 255}}
								case 36: // Cyan
									currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{0, 255, 255, 255}}
								case 90: // Dark Gray
									currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{128, 128, 128, 255}}
								case 38: // 256-Color Foreground
									if i+2 < len(codes) && codes[i+1] == "5" {
										colorIndex, _ := strconv.Atoi(codes[i+2])

										if colorIndex == 8 {
											currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{128, 128, 128, 255}}
										}

										i += 2
									}
								}
							}
						case 'A':
							term.CursorY -= getCSIParam(csiParams, 1)
							if term.CursorY < 0 { term.CursorY = 0 }
						case 'B':
							term.CursorY += getCSIParam(csiParams, 1)
							if term.CursorY >= term.Rows { term.CursorY = term.Rows - 1 }
						case 'C':
							term.CursorX += getCSIParam(csiParams, 1)
							if term.CursorX >= term.Cols { term.CursorX = term.Cols - 1 }
						case 'D':
							term.CursorX -= getCSIParam(csiParams, 1)
							if term.CursorX < 0 { term.CursorX = 0 }
						// Move cursor to X,Y position
						case 'H', 'f':
							row, col := getCSICoords(csiParams)

							term.CursorY = row - 1
							term.CursorX = col - 1

							if term.CursorY < 0 { term.CursorY = 0 }
							if term.CursorY >= term.Rows { term.CursorY = term.Rows - 1 }
							if term.CursorX < 0 { term.CursorX = 0 }
							if term.CursorX >= term.Cols { term.CursorX = term.Cols - 1 }
						// Clear Display
						case 'J':
							n := getCSIParam(csiParams, 0)
							// TODO: refactor to switch?
							if n == 0 {
								for x := term.CursorX; x < term.Cols; x++ { 
									term.Grid[term.CursorY][x] = widget.TextGridCell{Rune: ' '} 
								}
								for y := term.CursorY + 1; y < term.Rows; y++ {
									for x := range term.Cols { term.Grid[y][x] = widget.TextGridCell{Rune: ' '} }
								}
							} else if n == 2 {
								for y := range term.Rows {
									for x := range term.Cols { term.Grid[y][x] = widget.TextGridCell{Rune: ' '} }
								}
							}
						case 'K':
							n := getCSIParam(csiParams, 0)
							switch n {
								case 0:
									for x := term.CursorX; x < term.Cols; x++ {
										term.Grid[term.CursorY][x] = widget.TextGridCell{Rune: ' '}
									}
								case 2:
									for x := 0; x < term.Cols; x++ { 
										term.Grid[term.CursorY][x] = widget.TextGridCell{Rune: ' '} 
									}
							}
					}
					inCSI = false
					term.mu.Unlock()
					continue
				}
			}
			// --- END ANSI PARSER ---

			// --- Grid Write Logic ---
			// TODO: refactor to switch
			if r == '\b' || r == '\x7f' {
				if term.CursorX > 0 {

					term.CursorX--
				} else if term.CursorY > 0 {
					term.CursorY--
					term.CursorX = term.Cols - 1
				}
			} else if r == '\r' {
				term.CursorX = 0
			} else if r == '\n' {
				term.CursorY++
				if term.CursorY >= term.Rows {
					term.Scroll()
					term.CursorY = term.Rows - 1
				}
			} else {
				if term.CursorX >= term.Cols {
					term.CursorX = 0
					term.CursorY++
					if term.CursorY >= term.Rows {
						term.Scroll()
						term.CursorY = term.Rows - 1
					}
				}

				term.Grid[term.CursorY][term.CursorX] = widget.TextGridCell{Rune: r, Style: currentStyle}
				//if term.CursorX + 1 < term.Cols { term.Grid[term.CursorY][term.CursorX + 1] = widget.TextGridCell{Rune: ' ', Style: cursorStyle} }
				term.CursorX++
			}

			term.mu.Unlock()
		}
	}()

	// Goroutine that renders to UI
	go func() {
		for {
			time.Sleep(16 * time.Millisecond)

			var gridRows []widget.TextGridRow

			term.mu.Lock()

			for y := 0; y < term.Rows; y++ {
				rowCells := make([]widget.TextGridCell, term.Cols)

				for x := 0; x < term.Cols; x++ {
					rowCells[x] = term.Grid[y][x]
				}

				gridRows = append(gridRows, widget.TextGridRow{Cells: rowCells})
			}

			// term.Grid[term.CursorY][term.CursorX] = widget.TextGridCell{Rune: ' ', Style: cursorStyle}
			term.mu.Unlock()

			fyne.Do(func() {
				// Apply the rows directly to the TextGrid and refresh it
				ui.Rows = gridRows
				ui.Refresh()
			})
		}
	}()

	t := container.New(tl, ui)

	w.SetContent(t)
	w.Resize(fyne.NewSize(1200, 800))
	w.ShowAndRun()
}
