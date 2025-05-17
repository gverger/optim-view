package systems

import (
	"context"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/ark/ecs"
)

func NewDebug(font rl.Font, fontSize int) *Debug {
	return &Debug{font: font, fontSize: fontSize, spacing: 1}
}

type Debug struct {
	font     rl.Font
	fontSize int
	spacing  int

	board ecs.Resource[DebugBoard]

	mapper *ecs.Map1[Position]
	panel  ecs.Entity
}

// Close implements System.
func (d *Debug) Close() {
}

func (d *Debug) Initialize(w *ecs.World) {
	d.board = ecs.NewResource[DebugBoard](w)

	d.mapper = ecs.NewMap1[Position](w)
	d.panel = d.mapper.NewEntity(&Position{10, float64(rl.GetScreenHeight() - 300)})
}

func (d *Debug) Update(ctx context.Context, w *ecs.World) {
	if !d.board.Has() {
		return
	}

	board := d.board.Get()

	text := strings.Join(board.TextLines, "\n")
	pos := d.mapper.Get(d.panel)

	topLeft := rl.NewVector2(float32(pos.X), float32(pos.Y))
	size := rl.MeasureTextEx(d.font, text, float32(d.fontSize), float32(d.spacing))

	rl.DrawRectangleRec(rl.NewRectangle(topLeft.X, topLeft.Y, size.X, size.Y), rl.Fade(rl.Gray, 0.3))

	rl.DrawTextEx(
		d.font,
		text,
		rl.NewVector2(float32(pos.X), float32(pos.Y)),
		float32(d.fontSize),
		float32(d.spacing),
		rl.DarkBlue,
	)

	board.Clean()
}

var _ System = &Debug{}
