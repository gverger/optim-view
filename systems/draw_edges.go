package systems

import (
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
}

func (d *DrawEdges) Initialize(w *ecs.World) {
	d.filter = *generic.NewFilter1[Edge]()
	d.filterNodes = generic.NewMap2[Position, Node](w)
	d.visibleWorld = generic.NewResource[VisibleWorld](w)
}

func (d *DrawEdges) Update(w *ecs.World) {
	visible := d.visibleWorld.Get()
	query := d.filter.Query(w)
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
		y1 := p1.Y + from.SizeY
		x2 := p2.X + to.SizeX/2
		y2 := p2.Y

		src := rl.NewVector2(float32(x1), float32(y1))
		ctrlA := rl.NewVector2(float32(x1), float32(y1+20))
		ctrlB := rl.NewVector2(float32(x2), float32(y2-20))
		dst := rl.NewVector2(float32(x2), float32(y2))

		rl.DrawSplineSegmentBezierCubic(src, ctrlA, ctrlB, dst, 1, rl.Gray)
	}

	// query := d.filter.Query(w)
	// edges := make([]ecs.Entity, 0, query.Count())
	// for query.Next() {
	// 	edges = append(edges, query.Entity())
	// }
	// for _, edge := range edges {
	// 	jointsQuery := d.filterJoints.Query(w, edge)
	// 	joints := make([]Position, jointsQuery.Count())
	// 	for jointsQuery.Next() {
	// 		p, j := jointsQuery.Get()
	// 		joints[j.Order] = *p
	// 	}
	// 	for i := 1; i < len(joints); i++ {
	// 		src := rl.NewVector2(float32(joints[i-1].X), float32(joints[i-1].Y))
	// 		ctrlA := rl.NewVector2(float32(joints[i-1].X), float32(joints[i-1].Y+50))
	// 		ctrlB := rl.NewVector2(float32(joints[i].X), float32(joints[i].Y-50))
	// 		dst := rl.NewVector2(float32(joints[i].X), float32(joints[i].Y))
	//
	// 		rl.DrawSplineSegmentBezierCubic(src, ctrlA, ctrlB, dst, 1, rl.Green)
	// 	}
	// }
}

var _ System = &DrawEdges{}
