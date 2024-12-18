package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
)

func NewDrawEdges(font rl.Font) *DrawEdges {
	return &DrawEdges{font: font}
}

type DrawEdges struct {
	font   rl.Font
}

func (d *DrawEdges) Initialize(w *ecs.World) {
}

func (d *DrawEdges) Update(w *ecs.World) {
}

var _ System = &DrawEdges{}
