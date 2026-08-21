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

func (t *Terminal) Scroll() {
	for y := 1; y < t.Rows; y++ {
		t.Grid[y-1] = t.Grid[y]
	}
	t.Grid[t.Rows-1] = make([]widget.TextGridCell, t.Cols)
	for x := range t.Cols {
		t.Grid[t.Rows-1][x] = widget.TextGridCell{Rune: ' '}
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
	term := &Terminal{ 
		CursorX: 0, 
		CursorY: 0,
		defaultStyle: defaultStyle,
		currentStyle: defaultStyle,
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


	// Goroutine that reads from pty and parses ANSI
	go func() {
		term.ReadPTY(reader)
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

	// Create display container
	t := container.New(tl, ui)

	// Set window content & run app
	w.SetContent(t)
	w.Resize(fyne.NewSize(1200, 800))
	w.ShowAndRun()
}

