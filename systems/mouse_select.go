package systems

import (
	"context"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/ark/ecs"
)

func NewMouseSelector() *MouseSelector {
	return &MouseSelector{}
}

type MouseSelector struct {
	shapes *ecs.Filter3[Position, Shape, VisibleElement]
	mapper *ecs.Map3[Position, Shape, VisibleElement]
	grid   ecs.Resource[Grid]

	input     ecs.Resource[Input]
	selection ecs.Resource[NodeSelection]
}

// Close implements System.
func (m *MouseSelector) Close() {
}

func (m *MouseSelector) Initialize(w *ecs.World) {
	m.shapes = ecs.NewFilter3[Position, Shape, VisibleElement](w)
	m.mapper = ecs.NewMap3[Position, Shape, VisibleElement](w)
	m.input = ecs.NewResource[Input](w)
	m.selection = ecs.NewResource[NodeSelection](w)
	m.grid = ecs.NewResource[Grid](w)
}

func (m *MouseSelector) Update(ctx context.Context, w *ecs.World) {
	selection := m.selection.Get()

	input := m.input.Get()

	if !input.Active {
		selection.Hovered = ecs.Entity{}
		return
	}

	mWorld := input.Mouse.InWorld
	mouse := rl.NewVector2(float32(mWorld.X), float32(mWorld.Y))

	if selection.HasHovered() {
		if m.mapper.HasAll(selection.Hovered) {
			pos, shape, _ := m.mapper.Get(selection.Hovered)
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
				if m.mapper.HasAll(e) {
					pos, shape, _ := m.mapper.Get(e)
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
