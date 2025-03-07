package systems

import "image/color"

type palette struct {
	Background color.RGBA
	Hovered    color.RGBA
	Selected   color.RGBA
	TextColor  color.RGBA
}

var Palette = palette{
	Background: HexToRGBA(0xDFD2D2),
	Hovered:    HexToRGBA(0xA9B5DF),
	Selected:   HexToRGBA(0x7886C7),
	TextColor:  HexToRGBA(0x2D336B),
}

func HexToRGBA(hex int) color.RGBA {
	return color.RGBA{
		R: uint8((hex >> 16) & 0xFF),
		G: uint8((hex >> 8) & 0xFF),
		B: uint8((hex) & 0xFFb),
		A: 255,
	}
}
