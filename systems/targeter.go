package systems

import (
	"context"

	"github.com/gen2brain/raylib-go/easings"
	"github.com/mlange-42/ark/ecs"
)

func NewTargeter() *Targeter {
	return &Targeter{}
}

type Targeter struct {
	filterTarget2 *ecs.Filter2[Position, Target2]
	filterTarget1 *ecs.Filter2[Size, Target1]

	grid          ecs.Resource[Grid]
	boundingBoxes ecs.Resource[SubTreeBoundingBoxes]

	tick int
}

// Close implements System.
func (m *Targeter) Close() {
}

func (m *Targeter) Initialize(w *ecs.World) {
	m.filterTarget2 = ecs.NewFilter2[Position, Target2](w)
	m.filterTarget1 = ecs.NewFilter2[Size, Target1](w)
	m.grid = ecs.NewResource[Grid](w)
	m.boundingBoxes = ecs.NewResource[SubTreeBoundingBoxes](w)
}

func (m *Targeter) Update(ctx context.Context, w *ecs.World) {
	m.tick++
	updateTarget1(m, w)
	updateTarget2(m, w)
}

func updateTarget2(m *Targeter, w *ecs.World) {
	query := m.filterTarget2.Query()
	grid := m.grid.Get()
	boundingBoxes := m.boundingBoxes.Get()

	for query.Next() {
		pos, tar := query.Get()
		if tar.SinceTick == 0 {
			tar.SinceTick = m.tick
			tar.StartX = pos.X
			tar.StartY = pos.Y
			tar.Done = false
		}
		if tar.SinceTick+tar.Duration < m.tick {
			tar.Done = true
		}
		if tar.Done {
			continue
		}

		if tar.X != pos.X || tar.Y != pos.Y {
			oldX := pos.X
			oldY := pos.Y
			pos.X = float64(easings.QuadInOut(float32(m.tick-tar.SinceTick), float32(tar.StartX), float32(tar.X-tar.StartX), float32(tar.Duration)))
			pos.Y = float64(easings.QuadInOut(float32(m.tick-tar.SinceTick), float32(tar.StartY), float32(tar.Y-tar.StartY), float32(tar.Duration)))

			oldGPos := GridCoords(int(oldX), int(oldY))
			gpos := GridCoords(int(pos.X), int(pos.Y))
			grid.MoveEntity(query.Entity(), oldGPos, gpos)
			boundingBoxes.NodeMoved(query.Entity())
		} else {
			tar.Done = true
		}
	}
}

func updateTarget1(m *Targeter, w *ecs.World) {
	query := m.filterTarget1.Query()

	for query.Next() {
		size, tar := query.Get()
		if tar.SinceTick == 0 {
			tar.SinceTick = m.tick
			tar.StartX = size.Value
			tar.Done = false
		}
		if tar.SinceTick+tar.Duration < m.tick {
			tar.Done = true
		}
		if tar.Done {
			continue
		}

		if tar.X != size.Value {
			size.Value = easings.QuadInOut(float32(m.tick-tar.SinceTick), float32(tar.StartX), float32(tar.X-tar.StartX), float32(tar.Duration))
		} else {
			tar.Done = true
		}
	}
}

var _ System = &Targeter{}
