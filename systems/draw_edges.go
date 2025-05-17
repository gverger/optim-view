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

	filterNodes    *ecs.Filter2[Position, Node]
	filterChildren *ecs.Filter4[Position, Node, Parent, ChildOf]
	filterRoot     *ecs.Filter2[Position, Node]

	debug         ecs.Resource[DebugBoard]
	selected      ecs.Resource[NodeSelection]
	boundingBoxes ecs.Resource[SubTreeBoundingBoxes]
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
	d.filterChildren = ecs.NewFilter4[Position, Node, Parent, ChildOf](w).With(ecs.C[VisibleElement]())
	d.filterRoot = ecs.NewFilter2[Position, Node](w).With(ecs.C[VisibleElement]()).Without(ecs.C[Parent]())

	d.debug = ecs.NewResource[DebugBoard](w)
	d.selected = ecs.NewResource[NodeSelection](w)

	d.boundingBoxes = ecs.NewResource[SubTreeBoundingBoxes](w)
}

func (d *DrawEdges) drawLevel(ctx context.Context, w *ecs.World, visible VisibleWorld, p1 *Position, from *Node, e ecs.Entity) {

	boundingBoxes := d.boundingBoxes.Get().boundingBoxes

	children := make([]child, 0, 100)
	children = append(children, child{e: e, p: p1, n: from})
	for len(children) > 0 {
		if ctx.Err() != nil {
			return
		}

		c := children[len(children)-1]
		children = children[:len(children)-1]

		bb := boundingBoxes[c.e]
		if bb.X > visible.MaxX || bb.X+bb.Width < visible.X || bb.Y > visible.MaxY || bb.Y+bb.Height < visible.Y {
			continue
		}

		p1 := c.p
		from := c.n
		e := c.e

		x1 := p1.X + from.SizeX/2
		y1 := p1.Y + from.SizeY + 8
		if y1 > visible.MaxY {
			continue
		}

		src := rl.NewVector2(float32(x1), float32(y1))

		qChildren := d.filterChildren.Query(ecs.Rel[ChildOf](e))
		startDrawn := false

		cxLeft := visible.MaxX + 1
		cxRight := visible.X - 1
		cy := float32(0.0)

		// children := make([]child, 0)
		for qChildren.Next() {
			p2, to, _, _ := qChildren.Get()

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
			if x2 < cxLeft {
				cxLeft = x2
			}
			if x2 > cxRight {
				cxRight = x2
			}
			cy = ctrlA.Y
			children = append(children, child{e: qChildren.Entity(), p: p2, n: to})
		}

		if cy >= float32(visible.Y) && cy <= float32(visible.MaxY) && cxRight >= visible.X && cxLeft <= visible.MaxX {
			if cxLeft < visible.X {
				cxLeft = visible.X
			}
			if cxRight > visible.MaxX {
				cxRight = visible.MaxX
			}
			rl.DrawLineEx(rl.NewVector2(float32(cxLeft), cy), rl.NewVector2(float32(cxRight), cy), 2, rl.Gray)
		}

	}
}

func (d *DrawEdges) Update(ctx context.Context, w *ecs.World) {

	rl.BeginMode2D(*d.camera.Get().Camera)

	rootQ := d.filterRoot.Query()
	for rootQ.Next() {
		p, n := rootQ.Get()
		d.drawLevel(ctx, w, *d.visibleWorld.Get(), p, n, rootQ.Entity())
	}

	rl.EndMode2D()
}

var _ System = &DrawEdges{}
