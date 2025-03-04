package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
)

type Mappings struct {
	nodeLookup map[uint64]ecs.Entity
	edgeLookup map[[2]uint64]ecs.Entity
}

type Mouse struct {
	OnScreen Position
	InWorld  Position
}

type VisibleWorld struct {
	X    float64
	Y    float64
	MaxX float64
	MaxY float64
}

type Shapes struct {
	Polygons [][]Position
}

type CameraHandler struct {
	Camera *rl.Camera2D
}

type SelectedNode struct {
	Entity ecs.Entity
}

func (s SelectedNode) IsSet() bool {
	return !s.Entity.IsZero()
}

type NavType uint

const (
	FreeNav     NavType = 0
	KeyboardNav NavType = 1
)

type NavigationMode struct {
	Nav NavType
}
