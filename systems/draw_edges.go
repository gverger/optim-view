package systems

import (
	"context"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/ark/ecs"
)

const (
	EdgeThickness = 2
)

func NewDrawEdges(font rl.Font) *DrawEdges {
	return &DrawEdges{font: font}
}

type DrawEdges struct {
	font         rl.Font
	filter       *ecs.Filter2[Edge, VisibleElement]
	mapNodes     *ecs.Map2[Position, Node]
	visibleWorld ecs.Resource[VisibleWorld]
	camera       ecs.Resource[CameraHandler]

	filterNodes *ecs.Filter2[Position, Node]
	filterEdges *ecs.Filter3[Position, Node, ChildOf]
}

// Close implements System.
func (d *DrawEdges) Close() {
}

func (d *DrawEdges) Initialize(w *ecs.World) {
	d.filter = ecs.NewFilter2[Edge, VisibleElement](w)
	d.mapNodes = ecs.NewMap2[Position, Node](w)
	d.visibleWorld = ecs.NewResource[VisibleWorld](w)
	d.camera = ecs.NewResource[CameraHandler](w)

	d.filterNodes = ecs.NewFilter2[Position, Node](w).With(ecs.C[VisibleElement]())
	d.filterEdges = ecs.NewFilter3[Position, Node, ChildOf](w).With(ecs.C[VisibleElement]())
}

func (d *DrawEdges) Update(ctx context.Context, w *ecs.World) {
	visible := d.visibleWorld.Get()

	qNodes := d.filterNodes.Query()
	rl.BeginMode2D(*d.camera.Get().Camera)
	for qNodes.Next() {
		p1, from := qNodes.Get()
		x1 := p1.X + from.SizeX/2
		y1 := p1.Y + from.SizeY + 8
		src := rl.NewVector2(float32(x1), float32(y1))

		qChildren := d.filterEdges.Query(ecs.Rel[ChildOf](qNodes.Entity()))
		startDrawn := false

		cx1 := visible.MaxX + 1
		cx2 := visible.X - 1
		cy := float32(0.0)

		for qChildren.Next() {
			p2, to, _ := qChildren.Get()

			x2 := p2.X + to.SizeX/2
			y2 := p2.Y - 8

			ctrlA := rl.NewVector2(float32(x1), float32((y1+y2)/2))
			ctrlB := rl.NewVector2(float32(x2), float32((y1+y2)/2))
			dst := rl.NewVector2(float32(x2), float32(y2))

			// Draw start of edge, from src
			if !startDrawn && x1 >= visible.X && x1 <= visible.MaxX && ctrlA.Y >= float32(visible.Y) && y1 <= visible.MaxY {
				rl.DrawLineEx(src, ctrlA, 2, rl.Gray)
				startDrawn = true
			}

			// Draw end of edge, the arrow to dst
			if x2 >= visible.X && x2 <= visible.MaxX && y2 >= visible.Y && float64(ctrlB.Y) <= visible.MaxY {
				rl.DrawLineEx(rl.NewVector2(ctrlB.X, ctrlB.Y-EdgeThickness/2), rl.NewVector2(dst.X, dst.Y-8), 2, rl.Gray)
				rl.DrawTriangle(rl.NewVector2(dst.X, dst.Y-8), dst, rl.NewVector2(dst.X+4, dst.Y-10), rl.Gray)
				rl.DrawTriangle(dst, rl.NewVector2(dst.X, dst.Y-8), rl.NewVector2(dst.X-4, dst.Y-10), rl.Gray)
			}

			// Update left and right of the edge horizontal line
			if x2 < cx1 {
				cx1 = x2
			}
			if x2 > cx2 {
				cx2 = x2
			}
			cy = ctrlA.Y
		}

		if cy >= float32(visible.Y) && cy <= float32(visible.MaxY) && cx2 >= visible.X && cx1 <= visible.MaxX {
			if cx1 < visible.X {
				cx1 = visible.X
			}
			if cx2 > visible.MaxX {
				cx2 = visible.MaxX
			}
			rl.DrawLineEx(rl.NewVector2(float32(cx1), cy), rl.NewVector2(float32(cx2), cy), 2, rl.Gray)
		}
	}

	rl.EndMode2D()
}

var _ System = &DrawEdges{}
