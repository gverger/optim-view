package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewMouseSelector() *MouseSelector {
	return &MouseSelector{}
}

type MouseSelector struct {
	filter *generic.Filter2[Position, Shape]
	mapper generic.Map2[Position, Shape]

	mouse   generic.Resource[Mouse]
	hovered generic.Resource[ecs.Entity]
}

func (m *MouseSelector) Initialize(w *ecs.World) {
	m.filter = generic.NewFilter2[Position, Shape]()
	m.mapper = generic.NewMap2[Position, Shape](w)
	m.mouse = generic.NewResource[Mouse](w)
	m.hovered = generic.NewResource[ecs.Entity](w)
}

func (m *MouseSelector) Update(w *ecs.World) {
	mWorld := m.mouse.Get().InWorld
	mouse := rl.NewVector2(float32(mWorld.X), float32(mWorld.Y))

	if m.hovered.Has() {
		pos, shape := m.mapper.Get(*m.hovered.Get())
		points := make([]rl.Vector2, 0, len(shape.Points))
		for _, p := range shape.Points {
			points = append(points, rl.NewVector2(float32(p.X+pos.X), float32(p.Y+pos.Y)))
		}
		if rl.CheckCollisionPointPoly(mouse, points) {
			return
		}
		m.hovered.Remove()
	}

	query := m.filter.Query(w)
	for query.Next() {
		pos, shape := query.Get()
		points := make([]rl.Vector2, 0, len(shape.Points))
		for _, p := range shape.Points {
			points = append(points, rl.NewVector2(float32(p.X+pos.X), float32(p.Y+pos.Y)))
		}
		if rl.CheckCollisionPointPoly(mouse, points) {
			e := query.Entity()
			m.hovered.Add(&e)
			query.Close()
			return
		}
	}
}

var _ System = &MouseSelector{}
