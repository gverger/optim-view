package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
)

type Size struct {
	Value float32
}

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

	scale float32
	midX  float32
	midY  float32

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

type VisibleElement struct{}

type Velocity struct {
	Dx float64
	Dy float64
}

type Target1 struct {
	X float32

	SinceTick int
	Duration  int
	StartX    float32
	Done      bool
}

func NewTarget1Empty(duration int) *Target1 {
	return &Target1{
		Duration: duration,

		SinceTick: -1,
		Done:      true,
	}
}

type Target2 struct {
	X float64
	Y float64

	SinceTick int
	Duration  int
	StartX    float64
	StartY    float64
	Done      bool
}

func NewTarget2Empty(duration int) *Target2 {
	return &Target2{
		Duration: duration,

		SinceTick: 0,
		Done:      true,
	}
}

type JointOf struct {
	ecs.Relation
	Order int
}
