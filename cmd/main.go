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

	// TODO: Reimplement this package
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

	a.Settings().SetTheme(&fontTheme{
		Theme: theme.DefaultTheme(),
		font: resourceFiraCodeNerdFontRegularTtf,
	})

	w := a.NewWindow("shftty")

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

	cfg, err := initConfig()
	if err != nil {
		fyne.LogError("Failed to parse config", err)
	}

	// Start pty
	os.Setenv("TERM", "xterm-256color")
	c := exec.Command(cfg.Shell)
	p, err := pty.StartWithSize(c, &pty.Winsize{
		Rows: uint16(44),
		Cols: uint16(120),
	})

	if err != nil {
		fyne.LogError("Failed to open pty", err)
		os.Exit(1)
	}

	defer c.Process.Kill()


	// Sus out font width for layout
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


	windowShortcuts(w, p)

	// buf reader for pty stream
	reader := bufio.NewReader(p)

	// Read & parse pty loop
	go func() {
		term.ReadPTY(reader, p)
	}()

	// Render loop
	// TODO: need to add an isDirty catch to skip redrawing valid screens
	// TODO: avoid re-allocating uiRows constantly

	gridRows := make([]widget.TextGridRow, 0)

	go func() {
		for {
			time.Sleep(16 * time.Millisecond)

			term.mu.Lock()

			if !term.isDirty {
				term.mu.Unlock()
				continue
			}

			// Resize rows slice if window resized
			if len(gridRows) != term.Rows || (term.Rows > 0 && len(gridRows[0].Cells) != term.Cols) {
				gridRows = make([]widget.TextGridRow, term.Rows)
				for y := 0; y < term.Rows; y++ {
					gridRows[y].Cells = make([]widget.TextGridCell, term.Cols)
				}
			}

			// mutate cells
			for y := 0; y < term.Rows; y++ {
				for x := 0; x < term.Cols; x++ {
					gridRows[y].Cells[x] = term.Grid[y][x]
				}
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

			// reset dirty flag
			term.isDirty = false
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

