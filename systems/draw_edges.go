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
	filter       generic.Filter1[Edge]
	filterJoints generic.Filter2[Position, JointOf]
	font         rl.Font
}

func (d *DrawEdges) Initialize(w *ecs.World) {
	d.filter = *generic.NewFilter1[Edge]()
	d.filterJoints = *generic.NewFilter2[Position, JointOf]().WithRelation(generic.T[JointOf]())
}

func (d *DrawEdges) Update(w *ecs.World) {
	query := d.filter.Query(w)
	for query.Next() {
		edge := query.Entity()
		jointsQuery := d.filterJoints.Query(w, edge)
		joints := make([]Position, jointsQuery.Count())
		for jointsQuery.Next() {
			p, j := jointsQuery.Get()
			joints[j.Order] = *p
		}
		for i := 1; i < len(joints); i++ {
			src := rl.NewVector2(float32(joints[i-1].X), float32(joints[i-1].Y))
			ctrlA := rl.NewVector2(float32(joints[i-1].X), float32(joints[i-1].Y+50))
			ctrlB := rl.NewVector2(float32(joints[i].X), float32(joints[i].Y-50))
			dst := rl.NewVector2(float32(joints[i].X), float32(joints[i].Y))

			rl.DrawSplineSegmentBezierCubic(src, ctrlA, ctrlB, dst, 1, rl.Green)
		}
	}
}

var _ System = &DrawEdges{}
