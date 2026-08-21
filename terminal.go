package main

import (
	"fmt"
	"bufio"
	"image/color"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	// fyne dependencies
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

)


// Terminal struct and interfaces
type Terminal struct {
	Grid         [][]widget.TextGridCell
	MainGrid     [][]widget.TextGridCell
	AltGrid      [][]widget.TextGridCell
	Cols         int
	Rows         int
	MarginTop    int
	MarginBottom int
	CursorX      int
	CursorY      int
	SavedX       int
	SavedY       int
	CursorHidden bool
	IsAlt        bool
	mu           sync.Mutex
	inEscape     bool
	inCSI        bool
	isPrivate    bool
	inOSC        bool
	inCharset    bool
	defaultStyle *widget.CustomTextGridStyle
	currentStyle *widget.CustomTextGridStyle
  CursorStyle  *widget.CustomTextGridStyle
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
				if t.MainGrid != nil {
					newMain[y][x] = t.MainGrid[y][x]
				}
				if t.AltGrid != nil {
					newAlt[y][x] = t.AltGrid[y][x]
				}
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
	t.MarginBottom = t.Rows - 1

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

// Scroll terminal
func (t *Terminal) Scroll() {
	for y := 1; y < t.Rows; y++ {
		t.Grid[y-1] = t.Grid[y]
	}
	t.Grid[t.Rows-1] = make([]widget.TextGridCell, t.Cols)
	for x := range t.Cols {
		t.Grid[t.Rows-1][x] = widget.TextGridCell{Rune: ' '}
	}
}

// Move the terminal cursor
func (term *Terminal) MoveCursorUp(n int) {
	term.CursorY -= n
	if term.CursorY < 0 {
		term.CursorY = 0
	}
}

func (term *Terminal) MoveCursorDown(n int) {
	term.CursorY += n
	if term.CursorY >= term.Rows {
		term.CursorY = term.Rows - 1
	}
}

func (term *Terminal) MoveCursorLeft(n int) {
	term.CursorX -= n
	if term.CursorX < 0 {
		term.CursorX = 0
	}
}

func (term *Terminal) MoveCursorRight(n int) {
	term.CursorX += n
	if term.CursorX >= term.Cols {
		term.CursorX = term.Cols - 1
	}
}

func (term *Terminal) ClearScreen() {
	for y := 0; y < term.Rows; y++ {
		for x := 0; x < term.Cols; x++ {
			term.Grid[y][x] = widget.TextGridCell{Rune: ' ', Style: term.currentStyle}
		}
	}

}

// Read PTY stream and process commands
func (term *Terminal) ReadPTY(reader *bufio.Reader, p *os.File) {
	csiParams := "" // The numbers inside the sequence cmd "31;1"
	oscParams := "" // The numbers inside the osc sequence cmd

	for {
		r, _, err := reader.ReadRune()

		if err != nil {
			if err == io.EOF {
				os.Exit(0)
				return
			}
			os.Exit(0)
		}

		term.mu.Lock()

		// Escape Intercept
		if r == '\x1b' {

			// check if we have a buffered osc response
			if term.inOSC && (oscParams == "11;?" || strings.HasPrefix(oscParams, "11;?")) {
				// background color requested
				response := "\x1b]11;rgb:0000/0000/0000\a"
				_, err := p.WriteString(response)
				if err != nil {
					fyne.LogError("Failed to write OSC 11 response", err)
				}
			}

			term.inEscape = true
			term.inCSI = false
			term.inOSC = false
			term.isPrivate = false
			term.inCharset = false
			term.mu.Unlock()
			continue
		}

		// OSC Parser
		if term.inOSC {
			if r == '\a' || r == '\x07' {
				term.inOSC = false

				// background color requested
				if oscParams == "11;?" || strings.HasPrefix(oscParams, "11;?") {
					response := "\x1b]11;rgb:0000/0000/0000\x1b\\"
					_, err := p.WriteString(response)
					if err != nil {
						fyne.LogError("Failed to write OSC 11 response", err)
					}
				}
			} else {
				// loop until '\a' / '\x07' or '\x1b'
				oscParams += string(r)
			}
			term.mu.Unlock()
			continue
		}

		// Charset Parser
		if term.inCharset {
			// ignore next character ie: 'B' or '0'
			term.inCharset = false
			term.mu.Unlock()
			continue
		}

		// ANSI Router
		if term.inEscape {
			if r == '[' {
				term.inCSI = true
				term.inEscape = false
				csiParams = "" // clear csiParams
				term.mu.Unlock()
				continue
			}
			if r == ']' {
				term.inOSC = true
				term.inEscape = false
				oscParams = "" // clear oscParams
				term.mu.Unlock()
				continue
			}
			if r == '(' || r == ')' || r == '*' || r == '+' {
				term.inCharset = true
				term.inEscape = false
				term.mu.Unlock()
				continue
			}
			if r == 'c' {
				term.IsAlt = false
				term.Grid = term.MainGrid
				term.CursorX, term.CursorY = 0, 0
				term.MarginTop = 0
				term.MarginBottom = term.Rows - 1
				term.currentStyle = term.defaultStyle

				term.ClearScreen()
			}
			term.inEscape = false // Unsupported escape, abort
			term.mu.Unlock()
			continue
		}

		// CSI Parser
		if term.inCSI {
			// catch DEC private mode
			if r == '?' {
				term.isPrivate = true
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
				if term.isPrivate {
					switch r {
					case 'h': // Set mode
						param := getCSIParam(csiParams, 0)
						if param == 1049 {
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
						} else if param == 25 {
							term.CursorHidden = false
						}
					case 'c': // Device Attributes Request
						if strings.HasPrefix(csiParams, ">") {
							// mimic xterm
							// TODO: figure out what this means
							_, _ = p.WriteString("\x1b[>0;276;0c]")
						} else {
							// Primary DA: Mimic VT102
							_, _ = p.WriteString("\x1b[?6c")
						}
					case 'l': // Reset mode
						param := getCSIParam(csiParams, 0)
						if param == 1049 {
							term.IsAlt = false
							term.Grid = term.MainGrid
							term.CursorX, term.CursorY = term.SavedX, term.SavedY

							term.MarginTop = 0
							term.MarginBottom = term.Rows - 1

							term.currentStyle = term.defaultStyle
							term.CursorHidden = false
						} else if param == 25 {
							term.CursorHidden = true
						}
					}

					term.inCSI = false
					term.isPrivate = false
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
							term.currentStyle = term.defaultStyle
						case 31: // Red
							term.currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{255, 0, 0, 255}}
						case 32: // Green
							term.currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{0, 255, 0, 255}}
						case 33: // Yellow
							term.currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{255, 255, 0, 255}}
						case 34: // Blue
							term.currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{0, 100, 255, 255}}
						case 35: // Magenta
							term.currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{255, 0, 255, 255}}
						case 36: // Cyan
							term.currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{0, 255, 255, 255}}
						case 90: // Dark Gray
							term.currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{128, 128, 128, 255}}
						case 38: // 256-Color Foreground
							if i+2 < len(codes) && codes[i+1] == "5" {
								colorIndex, _ := strconv.Atoi(codes[i+2])

								if colorIndex == 8 {
									term.currentStyle = &widget.CustomTextGridStyle{FGColor: color.RGBA{128, 128, 128, 255}}
								}

								i += 2
							}
						}
					}
				case 'A':
					term.MoveCursorUp(getCSIParam(csiParams, 1))
				case 'B':
					term.MoveCursorDown(getCSIParam(csiParams, 1))
				case 'C':
					term.MoveCursorRight(getCSIParam(csiParams, 1))
				case 'D':
					term.MoveCursorLeft(getCSIParam(csiParams, 1))
				// Move cursor to X,Y position
				case 'H', 'f':
					row, col := getCSICoords(csiParams)

					term.CursorY = row - 1
					term.CursorX = col - 1

					if term.CursorY < 0 {
						term.CursorY = 0
					}
					if term.CursorY >= term.Rows {
						term.CursorY = term.Rows - 1
					}
					if term.CursorX < 0 {
						term.CursorX = 0
					}
					if term.CursorX >= term.Cols {
						term.CursorX = term.Cols - 1
					}
				// Clear Display
				case 'J':
					n := getCSIParam(csiParams, 0)
					// TODO: refactor to switch?
					if n == 0 {
						for x := term.CursorX; x < term.Cols; x++ {
							term.Grid[term.CursorY][x] = widget.TextGridCell{Rune: ' '}
						}
						for y := term.CursorY + 1; y < term.Rows; y++ {
							for x := range term.Cols {
								term.Grid[y][x] = widget.TextGridCell{Rune: ' '}
							}
						}
					} else if n == 2 {
						for y := range term.Rows {
							for x := range term.Cols {
								term.Grid[y][x] = widget.TextGridCell{Rune: ' '}
							}
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
				case 'L': // Insert lines
					n := getCSIParam(csiParams, 1)

					lines := (term.MarginBottom - term.CursorY) + 1
					if n > lines {
						n = lines
					}
					if n <= 0 {
						break
					}

					for y := term.MarginBottom; y >= term.CursorY + n; y-- {
						copy(term.Grid[y], term.Grid[y-1])
					}

					for y := term.CursorY; y < term.CursorY + n; y++ {
						for x := 0; x < term.Cols; x++ {
							term.Grid[y][x] = widget.TextGridCell{Rune: ' ', Style: term.currentStyle}
						}
					}
				case 'M': // delete lines
					n := getCSIParam(csiParams, 1)

					lines := (term.MarginBottom - term.CursorY) + 1

					if n > lines {
						n = lines
					}
					if n <= 0 {
						break
					}


					for y := term.CursorY; y <= term.MarginBottom - n; y++ {
						copy(term.Grid[y], term.Grid[y+n])
					}

					for y := term.MarginBottom - n + 1; y <= term.MarginBottom; y++ {
						for x := 0; x < term.Cols; x++ {
							term.Grid[y][x] = widget.TextGridCell{Rune: ' ', Style: term.currentStyle}
						}
					}
				case 'n':
					// Cursor Position Report (CSR)
					param := getCSIParam(csiParams, 0) 

					if param == 6 {
						row := term.CursorY + 1
						col := term.CursorX + 1

						response := fmt.Sprintf("\x1b[%d;%dR", row, col)

						_, err := p.WriteString(response)
						if err != nil {
							fyne.LogError("Failed to write CSR back to PTY", err)
						}
					} else if param == 5 {
						_, err := p.WriteString("\x1b[0n")
						if err != nil {
							fyne.LogError("Failed to write DSR", err)
						}
					}
				case 'P': // delete n chars from line
					n := getCSIParam(csiParams, 1)

					cols := (term.Cols - term.CursorX)

					if n > cols {
						n = cols
					}
					if n <= 0 {
						break
					}

					copy(term.Grid[term.CursorY][term.CursorX:], term.Grid[term.CursorY][term.CursorX+n:])

					for x := (term.CursorX - n) + 1; x < term.Cols; x++ {
						term.Grid[term.CursorY][x] = widget.TextGridCell{Rune: ' ', Style: term.currentStyle}
					}
				}
				term.inCSI = false
				term.mu.Unlock()
				continue
			}
		}

		// Grid Write Logic
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

			term.Grid[term.CursorY][term.CursorX] = widget.TextGridCell{Rune: r, Style: term.currentStyle}
			term.CursorX++
		}

		term.mu.Unlock()
	}
}
