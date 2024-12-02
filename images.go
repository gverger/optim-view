package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

func ImageFromSVG(svg string) (*rl.Image, bool) {
	if len(svg) == 0 {
		return nil, false
	}
	// log.Info().Str("svg", svg).Msg("Tracing")
	fmt.Println(svg)
	c := Must(canvas.ParseSVG(strings.NewReader(svg)))
	c.WriteFile("./writefile.svg", renderers.PNG())

	ca := canvas.New(300, 300)
	ctx := canvas.NewContext(ca)

	c.RenderTo(ca)

	var data bytes.Buffer
	c.Write(bufio.NewWriter(&data), renderers.PNG(canvas.DPMM(3.2)))
	canvas.DrawPreview(ctx)

	return rl.LoadImageFromMemory(".png", data.Bytes(), int32(len(data.Bytes()))), true
}
