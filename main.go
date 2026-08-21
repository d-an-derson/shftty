package main

import (
	"bufio"
	"image/color"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	// fyne dependencies
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/creack/pty"
)


// Theme struct and interfaces
type fontTheme struct {
	fyne.Theme
	font fyne.Resource
}

func (t *fontTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return t.font
	}

	// Fall back to default Fyne theme for ui components
	return t.Theme.Font(style)
}

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

// =====================
// ------- Main --------
func main() {
	a := app.New()

	fontRes, err := fyne.LoadResourceFromPath("./fonts/FiraCodeNerdFont-Regular.ttf")
	if err != nil {
		fyne.LogError("Failed to load Nerd Font", err)
		os.Exit(1)
	}

	a.Settings().SetTheme(&fontTheme{
		Theme: theme.DefaultTheme(),
		font: fontRes,
	})

	w := a.NewWindow("shfte")

	defaultStyle := &widget.CustomTextGridStyle{
		FGColor: color.White,
		BGColor: color.Transparent,
	}

	ui := widget.NewTextGrid() // Create a new TextGrid

	// init terminal struct
	term := &Terminal{ 
		CursorX: 0, 
		CursorY: 0,
		defaultStyle: defaultStyle,
		currentStyle: defaultStyle,
		CursorHidden: false,
		MarginTop: 0,
		MarginBottom: 0,
		CursorStyle: &widget.CustomTextGridStyle{
			FGColor: color.RGBA{128,128,128,1},
			BGColor: color.White,
		},
	}

	term.Resize(120, 44)

	// Start pty
	os.Setenv("TERM", "xterm-256color")
	c := exec.Command("/bin/zsh")
	p, err := pty.StartWithSize(c, &pty.Winsize{
		Rows: uint16(44),
		Cols: uint16(120),
	})

	if err != nil {
		fyne.LogError("Failed to open pty", err)
		os.Exit(1)
	}

	defer c.Process.Kill()

	charSize := widget.NewTextGridFromString("X").MinSize()

	// debounce resize timer
	var resizeTimer *time.Timer
	var resizeMu sync.Mutex

	tl := &termLayout{
		onResize: func(s fyne.Size) {
			if charSize.Width <= 0 || charSize.Height <= 0 {
				return
			}

			newCols := int(s.Width / charSize.Width)
			newRows := int(s.Height / charSize.Height)

			// Debounce logic
			resizeMu.Lock()

			if resizeTimer != nil {
				resizeTimer.Stop()
			}

			resizeTimer = time.AfterFunc(50*time.Millisecond, func() {
				if newCols == term.Cols && newRows == term.Rows {
					return
				}

				term.Resize(newCols, newRows)

				_ = pty.Setsize(p, &pty.Winsize{
					Rows: uint16(newRows),
					Cols: uint16(newCols),
				})
			})

			resizeMu.Unlock()
		},
	}

	// Callback function that handles special keypresses
	onTypedKey := func(e *fyne.KeyEvent) {
		switch e.Name {
				case fyne.KeyEnter, fyne.KeyReturn:
					_, _ = p.Write([]byte{'\r'})
				case fyne.KeyBackspace:
					_, _ = p.Write([]byte{'\x7f'})
				case fyne.KeyTab:
					_, _ = p.Write([]byte{'\t'})
				case fyne.KeyEscape:
					_, _ = p.Write([]byte{'\x1b'})
					
				// Arrow keys
				case fyne.KeyUp:
					_, _ = p.Write([]byte("\x1b[A"))
				case fyne.KeyDown:
					_, _ = p.Write([]byte("\x1b[B"))
				case fyne.KeyRight:
					_, _ = p.Write([]byte("\x1b[C"))
				case fyne.KeyLeft:
					_, _ = p.Write([]byte("\x1b[D"))
					
				// Navigation keys
				case fyne.KeyHome:
					_, _ = p.Write([]byte("\x1b[H"))
				case fyne.KeyEnd:
					_, _ = p.Write([]byte("\x1b[F"))
				case fyne.KeyPageUp:
					_, _ = p.Write([]byte("\x1b[5~"))
				case fyne.KeyPageDown:
					_, _ = p.Write([]byte("\x1b[6~"))
				}
	}

	// Callback function that handles character keypresses
	onTypedRune := func(r rune) {
		_, _ = p.WriteString(string(r))
	}

	w.Canvas().SetOnTypedKey(onTypedKey)
	w.Canvas().SetOnTypedRune(onTypedRune)

	// Register Crtl + Key shortcuts and write them to pty
	// TODO: Implement Crtl + A-Z shortcuts
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyA,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+A is byte value 1
		_, _ = p.Write([]byte{'\x01'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyB, 
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+B is byte value 2
		_, _ = p.Write([]byte{'\x02'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyC, // Ctrl+C
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+C is byte value 3
		_, _ = p.Write([]byte{'\x03'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyD, // Crtl+D (EOF)
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+D is byte value 4
		_, _ = p.Write([]byte{'\x04'})
	})

	// buf reader for pty stream
	reader := bufio.NewReader(p)

	// Read & parse pty loop
	go func() {
		term.ReadPTY(reader, p)
	}()

	// Render loop
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

			// get cursor x,y and render cursor style 
			if term.CursorHidden == false && term.CursorY >= 0 && term.CursorY < len(gridRows) {
				if term.CursorX >= 0 && term.CursorX < term.Cols {
					// get rune at cursor x,y
					targetCell := gridRows[term.CursorY].Cells[term.CursorX]

					// default to space if cell is empty
					if targetCell.Rune == 0 {
						targetCell.Rune = ' '
					}

					targetCell.Style = term.CursorStyle
					gridRows[term.CursorY].Cells[term.CursorX] = targetCell
				}
			}

			term.mu.Unlock()

			fyne.Do(func() {
				// Apply the rows directly to the TextGrid and refresh it
				ui.Rows = gridRows
				ui.Refresh()
			})
		}
	}()

	// Create display container
	t := container.New(tl, ui)

	// Set window content & run app
	w.SetContent(t)
	w.Resize(fyne.NewSize(1200, 800))
	w.ShowAndRun()
}

