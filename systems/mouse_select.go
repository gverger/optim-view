package systems

import (
	"context"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewMouseSelector() *MouseSelector {
	return &MouseSelector{}
}

type MouseSelector struct {
	shapes *generic.Filter3[Position, Shape, VisibleElement]
	mapper generic.Map2[Position, Shape]

	mouse   generic.Resource[Mouse]
	hovered generic.Resource[ecs.Entity]
}

// Close implements System.
func (m *MouseSelector) Close() {
}

func (m *MouseSelector) Initialize(w *ecs.World) {
	m.shapes = generic.NewFilter3[Position, Shape, VisibleElement]()
	m.mapper = generic.NewMap2[Position, Shape](w)
	m.mouse = generic.NewResource[Mouse](w)
	m.hovered = generic.NewResource[ecs.Entity](w)
}

func (m *MouseSelector) Update(ctx context.Context, w *ecs.World) {
	mWorld := m.mouse.Get().InWorld
	mouse := rl.NewVector2(float32(mWorld.X), float32(mWorld.Y))

	if m.hovered.Has() {
		pos, shape := m.mapper.Get(*m.hovered.Get())
		points := make([]rl.Vector2, 0, len(shape.Points))
		for _, p := range shape.Points {
			points = append(points, rl.NewVector2(float32(p.X+pos.X), float32(p.Y+pos.Y)))
		}
		slices.Reverse(points)
		if rl.CheckCollisionPointPoly(mouse, points) {
			return
		}
		m.hovered.Remove()
	}

	query := m.shapes.Query(w)
	for query.Next() {
		pos, shape, _ := query.Get()

		points := make([]rl.Vector2, 0, len(shape.Points))
		for _, p := range shape.Points {
			points = append(points, rl.NewVector2(float32(p.X+pos.X), float32(p.Y+pos.Y)))
		}
		slices.Reverse(points)
		if rl.CheckCollisionPointPoly(mouse, points) {
			e := query.Entity()
			m.hovered.Add(&e)
			query.Close()
			return
		}
	}
}

var _ System = &MouseSelector{}
