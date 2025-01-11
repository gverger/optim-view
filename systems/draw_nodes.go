package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewDrawNodes(font rl.Font) *DrawNodes {
	return &DrawNodes{font: font}
}

type DrawNodes struct {
	font    rl.Font
	filter  generic.Filter2[Position, Node]
	hovered generic.Resource[ecs.Entity]
}

func (d *DrawNodes) Initialize(w *ecs.World) {
	d.filter = *generic.NewFilter2[Position, Node]()
	d.hovered = generic.NewResource[ecs.Entity](w)
}

func (d *DrawNodes) Update(w *ecs.World) {
	query := d.filter.Query(w)
	for query.Next() {
		pos, n := query.Get()
		color := n.color
		if d.hovered.Has() && *d.hovered.Get() == query.Entity() {
			color = rl.Green
		}
		rl.DrawRectangle(int32(pos.X), int32(pos.Y), int32(n.SizeX), int32(n.SizeY), color)
		rl.DrawTextEx(d.font, n.Text, rl.NewVector2(float32(pos.X), float32(pos.Y)), 8, 0, rl.Black)
	}
}

var _ System = &DrawNodes{}
