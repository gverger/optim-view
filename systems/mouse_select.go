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
	mapper generic.Map3[Position, Shape, VisibleElement]
	grid   generic.Resource[Grid]

	mouse     generic.Resource[Mouse]
	selection generic.Resource[NodeSelection]
}

// Close implements System.
func (m *MouseSelector) Close() {
}

func (m *MouseSelector) Initialize(w *ecs.World) {
	m.shapes = generic.NewFilter3[Position, Shape, VisibleElement]()
	m.mapper = generic.NewMap3[Position, Shape, VisibleElement](w)
	m.mouse = generic.NewResource[Mouse](w)
	m.selection = generic.NewResource[NodeSelection](w)
	m.grid = generic.NewResource[Grid](w)
}

func (m *MouseSelector) Update(ctx context.Context, w *ecs.World) {
	mWorld := m.mouse.Get().InWorld
	mouse := rl.NewVector2(float32(mWorld.X), float32(mWorld.Y))

	selection := m.selection.Get()
	if selection.HasHovered() {
		pos, shape, visible := m.mapper.Get(selection.Hovered)
		if pos != nil && shape != nil && visible != nil {
			points := make([]rl.Vector2, 0, len(shape.Points))
			for _, p := range shape.Points {
				points = append(points, rl.NewVector2(float32(p.X+pos.X), float32(p.Y+pos.Y)))
			}
			slices.Reverse(points)
			if rl.CheckCollisionPointPoly(mouse, points) {
				return
			}
		}
		selection.Hovered = ecs.Entity{}
	}

	gpos := GridCoords(int(mouse.X), int(mouse.Y))

	grid := m.grid.Get()
	for i := gpos.X - 1; i < gpos.X+1; i++ {
		for j := gpos.Y - 1; j < gpos.Y+1; j++ {
			for _, e := range grid.At(GridPos{X: i, Y: j}) {
				pos, shape, visible := m.mapper.Get(e)
				if pos != nil && shape != nil && visible != nil {
					points := make([]rl.Vector2, 0, len(shape.Points))
					for _, p := range shape.Points {
						points = append(points, rl.NewVector2(float32(p.X+pos.X), float32(p.Y+pos.Y)))
					}
					slices.Reverse(points)
					if rl.CheckCollisionPointPoly(mouse, points) {
						selection.Hovered = e
						return
					}
				}
			}
		}
	}

	// query := m.shapes.Query(w)
	// for query.Next() {
	// 	pos, shape, _ := query.Get()
	//
	// 	points := make([]rl.Vector2, 0, len(shape.Points))
	// 	for _, p := range shape.Points {
	// 		points = append(points, rl.NewVector2(float32(p.X+pos.X), float32(p.Y+pos.Y)))
	// 	}
	// 	slices.Reverse(points)
	// 	if rl.CheckCollisionPointPoly(mouse, points) {
	// 		e := query.Entity()
	// 		m.hovered.Add(&e)
	// 		query.Close()
	// 		return
	// 	}
	// }
}

var _ System = &MouseSelector{}
