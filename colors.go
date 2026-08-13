package main

import (
	"image/color"
)

type xtermColor struct {
	num int
	name string
	rgb color.RGBA
}

var xtermColors = []xtermColor{
	// System Colors - 16 color
	{
		num: 0,
		name: "Black",
		rgb: color.RGBA{0,0,0,1},
	},
	{
		num: 1,
		name: "Maroon",
		rgb: color.RGBA{128,0,0,1},
	},
	{
		num: 2,
		name: "Green",
		rgb: color.RGBA{0,128,0,1},
	},
	{
		num: 3,
		name: "Olive",
		rgb: color.RGBA{128,128,0,1},
	},
	{
		num: 4,
		name: "Navy",
		rgb: color.RGBA{0,0,128,1},
	},
	{
		num: 5,
		name: "Purple",
		rgb: color.RGBA{128,0,128,1},
	},
	{
		num: 6,
		name: "Teal",
		rgb: color.RGBA{0,128,128,1},
	},
	{
		num: 7,
		name: "Silver",
		rgb: color.RGBA{192,192,192,1},
	},
	{
		num: 8,
		name: "Grey",
		rgb: color.RGBA{128,128,128,1},
	},
	{
		num: 9,
		name: "Red",
		rgb: color.RGBA{255,0,0,1},
	},
	{
		num: 10,
		name: "Lime",
		rgb: color.RGBA{0,255,0,1},
	},
	{
		num: 11,
		name: "Yellow",
		rgb: color.RGBA{255,255,0,1},
	},
	{
		num: 12,
		name: "Blue",
		rgb: color.RGBA{0,0,255,1},
	},
	{
		num: 13,
		name: "Fuchsia",
		rgb: color.RGBA{255,0,255,1},
	},
	{
		num: 14,
		name: "Aqua",
		rgb: color.RGBA{0,255,255,1},
	},
	{
		num: 15,
		name: "White",
		rgb: color.RGBA{255,255,255,1},
	},
}
