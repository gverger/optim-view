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
	filter *generic.Filter2[Position, Target]

	tick int
}

func (m *Targeter) Initialize(w *ecs.World) {
	m.filter = generic.NewFilter2[Position, Target]()
}

func (m *Targeter) Update(ctx context.Context, w *ecs.World) {
	query := m.filter.Query(w)

	for query.Next() {
		pos, tar := query.Get()
		if tar.SinceTick == 0 {
			tar.SinceTick = m.tick
			tar.StartX = pos.X
			tar.StartY = pos.Y
		}

		if tar.X != pos.X || tar.Y != pos.Y {
			pos.X = float64(easings.SineInOut(float32(m.tick), float32(tar.StartX), float32(tar.X-tar.StartX), 60))
			pos.Y = float64(easings.SineInOut(float32(m.tick), float32(tar.StartY), float32(tar.Y-tar.StartY), 60))
		}
	}

	m.tick++
}

var _ System = &Targeter{}
