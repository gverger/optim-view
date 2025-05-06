package systems

import (
	"errors"
	"image/color"
)

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

var errInvalidFormat = errors.New("invalid format")

func StringToRGBA(s string) (c color.RGBA, err error) {
	c.A = 0xff

	if s[0] != '#' {
		return c, errInvalidFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = errInvalidFormat
	}
	return
}
