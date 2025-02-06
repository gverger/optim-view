package systems

import (
	"context"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewDrawEdges(font rl.Font) *DrawEdges {
	return &DrawEdges{font: font}
}

type DrawEdges struct {
	font         rl.Font
	filter       generic.Filter1[Edge]
	filterNodes  generic.Map2[Position, Node]
	visibleWorld generic.Resource[VisibleWorld]
	camera       generic.Resource[CameraHandler]
}

// Close implements System.
func (d *DrawEdges) Close() {
}

func (d *DrawEdges) Initialize(w *ecs.World) {
	d.filter = *generic.NewFilter1[Edge]()
	d.filterNodes = generic.NewMap2[Position, Node](w)
	d.visibleWorld = generic.NewResource[VisibleWorld](w)
	d.camera = generic.NewResource[CameraHandler](w)
}

func (d *DrawEdges) Update(ctx context.Context, w *ecs.World) {
	visible := d.visibleWorld.Get()
	query := d.filter.Query(w)

	rl.BeginMode2D(*d.camera.Get().Camera)

	for query.Next() {
		e := query.Get()
		p1, from := d.filterNodes.Get(e.From)
		p2, to := d.filterNodes.Get(e.To)
		// Simple way of detecting a line is not shown:
		// both nodes are left of the screen, or right, or above or below
		// still some edges drawn whereas they shouldn't but it seems ok
		if (p1.X > visible.MaxX && p2.X > visible.MaxX) ||
			(p1.Y > visible.MaxY && p2.Y > visible.MaxX) ||
			(p1.X+from.SizeX < visible.X && p2.X+to.SizeX < visible.X) ||
			(p1.Y+from.SizeY < visible.Y && p2.Y+to.SizeY < visible.Y) {
			continue
		}

		x1 := p1.X + from.SizeX/2
		y1 := p1.Y + (from.SizeY+from.DrawnSizeY)/2 + 8
		x2 := p2.X + to.SizeX/2
		y2 := p2.Y + (from.SizeY-from.DrawnSizeY)/2 - 8

		src := rl.NewVector2(float32(x1), float32(y1))
		ctrlA := rl.NewVector2(float32(x1), float32((y1+y2)/2))
		ctrlB := rl.NewVector2(float32(x2), float32((y1+y2)/2))
		dst := rl.NewVector2(float32(x2), float32(y2))

		rl.DrawLineEx(src, ctrlA, 2, rl.Gray)
		rl.DrawLineEx(ctrlA, ctrlB, 2, rl.Gray)
		rl.DrawLineEx(ctrlB, rl.NewVector2(dst.X, dst.Y-8), 2, rl.Gray)

		rl.DrawLineStrip([]rl.Vector2{
			src, ctrlA, ctrlB, dst,
		}, rl.Gray)

		rl.DrawTriangle(rl.NewVector2(dst.X, dst.Y-8), dst, rl.NewVector2(dst.X+4, dst.Y-10), rl.Gray)
		rl.DrawTriangle(dst, rl.NewVector2(dst.X, dst.Y-8), rl.NewVector2(dst.X-4, dst.Y-10), rl.Gray)
	}

	rl.EndMode2D()
}

var _ System = &DrawEdges{}
