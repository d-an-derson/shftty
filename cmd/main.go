package main

import (
	"bufio"
	// "image/color"
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

type Session struct {
	Term *Terminal
	PTY *os.File
	UI *widget.TextGrid
	Tab *container.TabItem
}

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

func createTab(terminals []*Terminal) (*container.TabItem, []*Terminal) {
	tab := container.NewTabItem("Tab n", widget.NewTextGrid())
	term := defaultTerminal()
	terminals = append(terminals, term)

	return tab, terminals
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

func createNewSession(title string, shell string) *Session {
	ui := widget.NewTextGrid()
	term := defaultTerminal()

	term.Resize(120, 44)

	c := exec.Command(shell)
	p, _ := pty.StartWithSize(c, &pty.Winsize{Rows: 44, Cols: 120})

	session := &Session{ Term: term, PTY: p, UI: ui }

	reader := bufio.NewReader(p)
	go term.ReadPTY(reader, p)

	go func() {
		gridRows := make([]widget.TextGridRow, 0)

		for {
			time.Sleep(16 * time.Millisecond)

			term.mu.Lock()

			if !term.isDirty || !term.isVisible {
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
					cell := term.Grid[y][x]

					if cell.Rune == 0 {
						cell.Rune = ' '
					}

					gridRows[y].Cells[x] = cell
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

	tabItem := container.NewTabItem(title, ui)
	session.Tab = tabItem

	return session
}

// =====================
// ------- Main --------
func main() {
	a := app.New()

	a.Settings().SetTheme(&fontTheme{
		Theme: theme.DefaultTheme(),
		font: resourceFiraCodeNerdFontRegularTtf,
	})

	var activeSession *Session
	var allSessions []*Session

	w := a.NewWindow("shftty")

	cfg, err := initConfig()
	if err != nil {
		fyne.LogError("Failed to parse config", err)
	}

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
				for _, session := range allSessions {
					if newCols == session.Term.Cols && newRows == session.Term.Rows {
						continue
					}

					session.Term.Resize(newCols, newRows)

					_ = pty.Setsize(session.PTY, &pty.Winsize{
						Rows: uint16(newRows),
						Cols: uint16(newCols),
					})
			}
			})

			resizeMu.Unlock()
		},
	}

	tabs := container.NewDocTabs()

	tabs.OnSelected = func(selected *container.TabItem) {
		for _, s := range allSessions {
			if s.Tab == selected {
				activeSession = s
				s.Term.isVisible = true
				s.Term.isDirty = true
			} else {
				s.Term.isVisible = false
			}
		}
	}

	initialSession := createNewSession("shftty", cfg.Shell)
	activeSession = initialSession
	allSessions = append(allSessions, initialSession)
	tabs.Append(initialSession.Tab)

	// Create display container
	t := container.New(tl, tabs)

	windowShortcuts(w, activeSession)

	// Callback function that handles special keypresses
	onTypedKey := func(e *fyne.KeyEvent) {
		if activeSession == nil { return }
		p := activeSession.PTY

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
		if activeSession == nil { return }
		_, _ = activeSession.PTY.WriteString(string(r))
	}

	w.Canvas().SetOnTypedKey(onTypedKey)
	w.Canvas().SetOnTypedRune(onTypedRune)


	// Set window content & run app
	w.SetContent(t)
	w.Resize(fyne.NewSize(1200, 800))
	w.ShowAndRun()
}

