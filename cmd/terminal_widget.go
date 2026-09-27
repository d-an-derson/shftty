package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type TerminalWidget struct {
	widget.BaseWidget
	term *Terminal
}

type termRenderer struct {
	widget *TerminalWidget
	backgrounds [][]*canvas.Rectangle
	texts [][]*canvas.Text
	cursor *canvas.Rectangle
	renderCache []fyne.CanvasObject
}

func NewTerminalWidget(term *Terminal) *TerminalWidget {
	t := &TerminalWidget{term: term}
	t.ExtendBaseWidget(t)
	return t
}

func (t *TerminalWidget) CreateRenderer() fyne.WidgetRenderer {
	term := t.term
	term.mu.Lock()
	defer term.mu.Unlock()

	r := &termRenderer{
		widget: t, 
		backgrounds: make([][]*canvas.Rectangle, term.Rows),
		texts: make([][]*canvas.Text, term.Rows),
		renderCache: make([]fyne.CanvasObject, 0, (term.Rows*term.Cols*2)+1),
	}

	for y := 0; y < term.Rows; y++ {
		r.backgrounds[y] = make([]*canvas.Rectangle, term.Cols)
		r.texts[y] = make([]*canvas.Text, term.Cols)

		for x := 0; x < term.Cols; x++ {
			bg := canvas.NewRectangle(color.Transparent)
			r.backgrounds[y][x] = bg
			r.renderCache = append(r.renderCache, bg)

			txt := canvas.NewText(" ", color.White)
			txt.TextStyle.Monospace = true
			r.texts[y][x] = txt
			r.renderCache = append(r.renderCache, txt)
		}
	}

	r.cursor = canvas.NewRectangle(color.RGBA{128, 128, 128, 255})
	r.renderCache = append(r.renderCache, r.cursor)

	return r
}

func (r *termRenderer) Objects() []fyne.CanvasObject {
	return r.renderCache
}

func (r *termRenderer) Destroy() {
	r.backgrounds = nil
	r.texts = nil
	r.cursor = nil
	r.renderCache = nil
}

func (r *termRenderer) MinSize() fyne.Size {
	term := r.widget.term
	measureChar := canvas.NewText("X", color.White)
	measureChar.TextStyle.Monospace = true
	charSize := measureChar.MinSize()

	return fyne.NewSize(
		charSize.Width * float32(term.Cols),
		charSize.Height * float32(term.Rows),
	)
}

func (r *termRenderer) syncCache() {
	term := r.widget.term

	if len(r.backgrounds) == term.Rows && (term.Rows == 0 || len(r.backgrounds[0]) == term.Cols) {
		return
	}
	
	r.backgrounds = make([][]*canvas.Rectangle, term.Rows)
	r.texts = make([][]*canvas.Text, term.Rows)
	r.renderCache = make([]fyne.CanvasObject, 0, (term.Rows*term.Cols*2)+1)

	for y := 0; y < term.Rows; y++ {
		r.backgrounds[y] = make([]*canvas.Rectangle, term.Cols)
		r.texts[y] = make([]*canvas.Text, term.Cols)

		for x := 0; x < term.Cols; x++ {
			bg := canvas.NewRectangle(color.Transparent)
			r.backgrounds[y][x] = bg
			r.renderCache = append(r.renderCache, bg)

			txt := canvas.NewText(" ", color.White)
			txt.TextStyle.Monospace = true
			r.texts[y][x] = txt
			r.renderCache = append(r.renderCache, txt)
		}

		if r.cursor == nil {
			r.cursor = canvas.NewRectangle(color.RGBA{128, 128, 128, 255})
		}
		r.renderCache = append(r.renderCache, r.cursor)
	}
}

func (r *termRenderer) Layout(size fyne.Size) {
	term := r.widget.term

	term.mu.Lock()
	defer term.mu.Unlock()

	if term.Cols <= 0 || term.Rows <= 0 {
		return
	}

	r.syncCache()

	charW := size.Width / float32(term.Cols)
	charH := size.Height / float32(term.Rows)
	cellSize := fyne.NewSize(charW, charH)

	for y := 0; y < term.Rows; y++ {
		for x := 0; x < term.Cols; x++ {
			pos := fyne.NewPos(float32(x)*charW, float32(y)*charH)

			r.backgrounds[y][x].Resize(cellSize)
			r.backgrounds[y][x].Move(pos)

			r.texts[y][x].Resize(cellSize)
			r.texts[y][x].Move(pos)
		}
	}
}

func (r *termRenderer) Refresh() {
	term := r.widget.term
	term.mu.Lock()
	defer term.mu.Unlock()

	r.syncCache()

	for y := 0; y < term.Rows; y++ {
		for x := 0; x < term.Cols; x++ {
			cell := term.Grid[y][x]
			bg := r.backgrounds[y][x]
			txt := r.texts[y][x]

			targetFG := cell.FGColor
			targetBG := cell.BGColor
			if cell.Inverse {
				targetFG, targetBG = targetBG, targetFG
			}

			isBlockCursor := !term.CursorHidden && (term.CursorShape == 0 || term.CursorShape == 1 || term.CursorShape == 2)
			if isBlockCursor && x == term.CursorX && y == term.CursorY {
				if targetBG == color.Transparent {
					targetBG = color.RGBA{128, 128, 128, 255}
					targetFG = color.RGBA{0, 0, 0, 255}
				} else {
					targetFG, targetBG = targetBG, targetFG
				}
			}

			targetRune := string(cell.Rune)
			if cell.Rune == 0 {
				targetRune = " "
			}

			if txt.Text != targetRune || txt.Color != targetFG || txt.TextStyle.Bold != cell.Bold || bg.FillColor != targetBG {
				bg.FillColor = targetBG
				bg.Refresh()

				txt.Text = targetRune
				txt.Color = targetFG
				txt.TextStyle.Bold = cell.Bold
				txt.TextStyle.Italic = cell.Italic
				txt.Refresh()
			}
		}
	}

	r.cursor.Hidden = term.CursorHidden
	isFloatingShape := term.CursorShape >= 3

	if !term.CursorHidden && isFloatingShape {
		r.cursor.Hidden = false
		size := r.widget.Size()
		charW := size.Width / float32(term.Cols)
		charH := size.Height / float32(term.Rows)

		cursorX := float32(term.CursorX) * charW
		cursorY := float32(term.CursorY) * charH

		switch term.CursorShape {
		case 3, 4:
			r.cursor.Resize(fyne.NewSize(charW, 2))
			cursorY += (charH - 2)
			r.cursor.FillColor = color.RGBA{128, 128, 128, 255}
		case 5, 6:
			r.cursor.Resize(fyne.NewSize(2, charH))
			r.cursor.FillColor = color.RGBA{128, 128, 128, 150}
		}

		r.cursor.Move(fyne.NewPos(cursorX, cursorY))
		r.cursor.Refresh()
	} else {
		if !r.cursor.Hidden {
			r.cursor.Hidden = true
			r.cursor.Refresh()
		}
	}
}



