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
	font         rl.Font
	filter       generic.Filter2[Position, Node]
	hovered      generic.Resource[ecs.Entity]
	visibleWorld generic.Resource[VisibleWorld]
}

func (d *DrawNodes) Initialize(w *ecs.World) {
	d.filter = *generic.NewFilter2[Position, Node]()
	d.hovered = generic.NewResource[ecs.Entity](w)
	d.visibleWorld = generic.NewResource[VisibleWorld](w)
}

func (d *DrawNodes) Update(w *ecs.World) {
	visible := d.visibleWorld.Get()
	query := d.filter.Query(w)

	cpt := 0
	for query.Next() {
		pos, n := query.Get()

		if pos.X > visible.MaxX || pos.Y > visible.MaxY || pos.X+n.SizeX < visible.X || pos.Y+n.SizeY < visible.Y {
			continue
		}
		cpt++

		color := n.color
		if d.hovered.Has() && *d.hovered.Get() == query.Entity() {
			color = rl.Green
		}
		rl.DrawRectangle(int32(pos.X), int32(pos.Y), int32(n.SizeX), int32(n.SizeY), color)
		rl.DrawTextEx(d.font, n.Text, rl.NewVector2(float32(pos.X), float32(pos.Y)), 8, 0, rl.Black)
	}
	// rl.DrawText(fmt.Sprintf("%d nodes", cpt), 10, int32(rl.GetScreenHeight())-100, 12, rl.Red)
}

var _ System = &DrawNodes{}
