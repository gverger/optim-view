package systems

import (
	"context"

	"github.com/gen2brain/raylib-go/easings"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewTargeter() *Targeter {
	return &Targeter{}
}

type Targeter struct {
	filterTarget2 *generic.Filter2[Position, Target2]
	filterTarget1 *generic.Filter2[Size, Target1]

	tick int
}

// Close implements System.
func (m *Targeter) Close() {
}

func (m *Targeter) Initialize(w *ecs.World) {
	m.filterTarget2 = generic.NewFilter2[Position, Target2]()
	m.filterTarget1 = generic.NewFilter2[Size, Target1]()
}

func (m *Targeter) Update(ctx context.Context, w *ecs.World) {
	m.tick++
	updateTarget1(m, w)
	updateTarget2(m, w)
}

func updateTarget2(m *Targeter, w *ecs.World) {
	query := m.filterTarget2.Query(w)

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
			pos.X = float64(easings.QuadIn(float32(m.tick-tar.SinceTick), float32(tar.StartX), float32(tar.X-tar.StartX), float32(tar.Duration)))
			pos.Y = float64(easings.QuadIn(float32(m.tick-tar.SinceTick), float32(tar.StartY), float32(tar.Y-tar.StartY), float32(tar.Duration)))
		} else {
			tar.Done = true
		}
	}
}

func updateTarget1(m *Targeter, w *ecs.World) {
	query := m.filterTarget1.Query(w)

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
			size.Value = easings.ExpoIn(float32(m.tick-tar.SinceTick), float32(tar.StartX), float32(tar.X-tar.StartX), float32(tar.Duration))
		} else {
			tar.Done = true
		}
	}
}

var _ System = &Targeter{}
