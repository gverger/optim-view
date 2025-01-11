package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewMover() *Mover {
	return &Mover{
		MaxSpeed: 100,
		MaxAcc:   0.08,
		Damp:     0.975,
	}
}

type Mover struct {
	MaxSpeed float64
	MaxAcc   float64

	Damp float64

	filter *generic.Filter3[Position, Velocity, Target]
}

func (m *Mover) Initialize(w *ecs.World) {
	m.filter = generic.NewFilter3[Position, Velocity, Target]()
}

func (m *Mover) Update(w *ecs.World) {
	query := m.filter.Query(w)
	for query.Next() {
		pos, vel, trg := query.Get()

		dir := rl.NewVector2(float32(trg.X-pos.X), float32(trg.Y-pos.Y))
		if rl.Vector2Length(dir) < 100 {
			pos.X = trg.X
			pos.Y = trg.Y
			vel.Dx = 0
			vel.Dy = 0
			continue
		}

		acc := rl.Vector2Normalize(dir)

		vel.Dx += float64(acc.X) * m.MaxAcc
		vel.Dy += float64(acc.Y) * m.MaxAcc

		vel.Dx *= m.Damp
		vel.Dy *= m.Damp

		pos.X += vel.Dx * m.MaxSpeed
		pos.Y += vel.Dy * m.MaxSpeed
	}
}

var _ System = &Mover{}
