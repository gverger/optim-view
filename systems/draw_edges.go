package systems

import (
	"context"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

const (
	EdgeThickness = 2
)

func NewDrawEdges(font rl.Font) *DrawEdges {
	return &DrawEdges{font: font}
}

type DrawEdges struct {
	font         rl.Font
	filter       generic.Filter2[Edge, VisibleElement]
	filterNodes  generic.Map2[Position, Node]
	visibleWorld generic.Resource[VisibleWorld]
	camera       generic.Resource[CameraHandler]
}

// Close implements System.
func (d *DrawEdges) Close() {
}

func (d *DrawEdges) Initialize(w *ecs.World) {
	d.filter = *generic.NewFilter2[Edge, VisibleElement]()
	d.filterNodes = generic.NewMap2[Position, Node](w)
	d.visibleWorld = generic.NewResource[VisibleWorld](w)
	d.camera = generic.NewResource[CameraHandler](w)
}

func (d *DrawEdges) Update(ctx context.Context, w *ecs.World) {
	visible := d.visibleWorld.Get()
	query := d.filter.Query(w)

	lhlines := make(map[float32]float32)
	rhlines := make(map[float32]float32)

	rl.BeginMode2D(*d.camera.Get().Camera)

	for query.Next() {
		e, _ := query.Get()

		p1, from := d.filterNodes.Get(e.From)
		p2, to := d.filterNodes.Get(e.To)

		x1 := p1.X + from.SizeX/2
		y1 := p1.Y + from.SizeY + 8
		x2 := p2.X + to.SizeX/2
		y2 := p2.Y - 8

		src := rl.NewVector2(float32(x1), float32(y1))
		ctrlA := rl.NewVector2(float32(x1), float32((y1+y2)/2))
		ctrlB := rl.NewVector2(float32(x2), float32((y1+y2)/2))
		dst := rl.NewVector2(float32(x2), float32(y2))

		if x1 >= visible.X && x1 <= visible.MaxX && ctrlA.Y >= float32(visible.Y) && y1 <= visible.MaxY {
			rl.DrawLineEx(src, ctrlA, 2, rl.Gray)
		}

		if x2 >= visible.X && x2 <= visible.MaxX && ctrlB.Y >= float32(visible.Y) && y2 <= visible.MaxY {
			rl.DrawLineEx(rl.NewVector2(ctrlB.X, ctrlB.Y-EdgeThickness/2), rl.NewVector2(dst.X, dst.Y-8), 2, rl.Gray)
			rl.DrawTriangle(rl.NewVector2(dst.X, dst.Y-8), dst, rl.NewVector2(dst.X+4, dst.Y-10), rl.Gray)
			rl.DrawTriangle(dst, rl.NewVector2(dst.X, dst.Y-8), rl.NewVector2(dst.X-4, dst.Y-10), rl.Gray)
		}

		xl1 := ctrlA.X
		xl2 := ctrlB.X
		if xl1 > xl2 {
			xl1, xl2 = xl2, xl1
		}
		if ctrlA.Y >= float32(visible.Y) && ctrlA.Y <= float32(visible.MaxY) && xl2 >= float32(visible.X) && xl1 <= float32(visible.MaxX) {
			if xl1 >= float32(visible.X) && xl2 <= float32(visible.MaxX) {
				rl.DrawLineEx(ctrlA, ctrlB, 2, rl.Gray)
			} else {
				if xl1 < float32(visible.X) {
					if l, ok := lhlines[ctrlA.Y]; !ok || l < xl2 {
						lhlines[ctrlA.Y] = xl2
					}
				}
				if xl2 > float32(visible.MaxX) {
					if l, ok := rhlines[ctrlA.Y]; !ok || l > xl1 {
						rhlines[ctrlA.Y] = xl1
					}
				}
			}
		}
	}
	for y, x := range lhlines {
		rl.DrawLineEx(rl.NewVector2(float32(visible.X), y), rl.NewVector2(min(x, float32(visible.MaxX)), y), 2, rl.Gray)
	}
	for y, x := range rhlines {
		rl.DrawLineEx(rl.NewVector2(max(x, float32(visible.X)), y), rl.NewVector2(float32(visible.MaxX), y), 2, rl.Gray)
	}

	rl.EndMode2D()
}

var _ System = &DrawEdges{}
