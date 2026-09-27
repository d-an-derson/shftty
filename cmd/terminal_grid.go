package main

import (
	"image/color"
)

type TerminalCell struct {
	Rune rune
	FGColor color.Color
	BGColor color.Color
	Bold bool
	Italic bool
	Underline bool
	Inverse bool
}


type TerminalGridStyle interface {
	TextColor() color.Color
	BackgroundColor() color.Color
	IsBold() bool
	IsItalic() bool
	IsUnderline() bool
}

type ExtendedGridStyle struct {
	FGColor   color.Color
	BGColor   color.Color
	Bold      bool
	Italic    bool
	Underline bool
}

func (s *ExtendedGridStyle) TextColor() color.Color {
	return s.FGColor
}

func (s *ExtendedGridStyle) BackgroundColor() color.Color {
	return s.BGColor
}

func (s *ExtendedGridStyle) IsBold() bool {
	return s.Bold
}

func (s *ExtendedGridStyle) IsItalic() bool {
	return s.Italic
}

func (s *ExtendedGridStyle) IsUnderline() bool {
	return s.Underline
}

