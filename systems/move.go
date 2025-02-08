package systems

import (
	"context"

	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewMover() *Mover {
	return &Mover{}
}

type Mover struct {
	filter *generic.Filter2[Position, Target2]
}

func (m *Mover) Close() {
}

func (m *Mover) Initialize(w *ecs.World) {
	m.filter = generic.NewFilter2[Position, Target2]()
}

func (m *Mover) Update(ctx context.Context, w *ecs.World) {
	query := m.filter.Query(w)
	for query.Next() {
		pos, trg := query.Get()

		pos.X = trg.X
		pos.Y = trg.Y
	}
}

var _ System = &Mover{}
