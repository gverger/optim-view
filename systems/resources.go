package systems

import "github.com/mlange-42/arche/ecs"

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
