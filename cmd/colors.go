package main

import (
	"image/color"
)
func getANSIColor(index int) color.Color {
	// 0-7: Standard, 8-15: Bright
	ansiColors := []color.Color{
		color.RGBA{0, 0, 0, 255},       // 0 Black
		color.RGBA{205, 0, 0, 255},     // 1 Red
		color.RGBA{0, 205, 0, 255},     // 2 Green
		color.RGBA{205, 205, 0, 255},   // 3 Yellow
		color.RGBA{0, 0, 238, 255},     // 4 Blue
		color.RGBA{205, 0, 205, 255},   // 5 Magenta
		color.RGBA{0, 205, 205, 255},   // 6 Cyan
		color.RGBA{229, 229, 229, 255}, // 7 White
		
		color.RGBA{127, 127, 127, 255}, // 8 Bright Black (Gray)
		color.RGBA{255, 0, 0, 255},     // 9 Bright Red
		color.RGBA{0, 255, 0, 255},     // 10 Bright Green
		color.RGBA{255, 255, 0, 255},   // 11 Bright Yellow
		color.RGBA{92, 92, 255, 255},   // 12 Bright Blue
		color.RGBA{255, 0, 255, 255},   // 13 Bright Magenta
		color.RGBA{0, 255, 255, 255},   // 14 Bright Cyan
		color.RGBA{255, 255, 255, 255}, // 15 Bright White
	}
	
	if index >= 0 && index < len(ansiColors) {
		return ansiColors[index]
	}
	return color.White
}

func get256Color(index int) color.Color {
	// Standard 16 colors
	if index >= 0 && index <= 15 {
		return getANSIColor(index)
	}
	
	// Grayscale ramp (232-255)
	if index >= 232 && index <= 255 {
		gray := uint8((index - 232) * 10 + 8)
		return color.RGBA{gray, gray, gray, 255}
	}
	
	// 6x6x6 RGB color cube (16-231)
	if index >= 16 && index <= 231 {
		index -= 16
		r := uint8((index / 36) * 51)
		g := uint8(((index % 36) / 6) * 51)
		b := uint8((index % 6) * 51)
		return color.RGBA{r, g, b, 255}
	}
	
	return color.White
}
