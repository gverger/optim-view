package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
)

type Position struct {
	X float64
	Y float64
}

type Shape struct {
	Points []Position
}

type Node struct {
	SizeX float64
	SizeY float64

	DrawnSizeX float64
	DrawnSizeY float64

	color  rl.Color
	Title  string
	Text   string
	hidden bool

	ShapeTransforms []ShapeTransform
	rendered        bool
	idx             int
}

type Edge struct {
	From ecs.Entity
	To   ecs.Entity
}

type Velocity struct {
	Dx float64
	Dy float64
}

type Target struct {
	X float64
	Y float64

	SinceTick int
	StartX    float64
	StartY    float64
}

type JointOf struct {
	ecs.Relation
	Order int
}
