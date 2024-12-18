package systems

import "github.com/mlange-42/arche/ecs"

type Mappings struct {
	nodeLookup map[uint64]ecs.Entity
	edgeLookup map[[2]uint64]ecs.Entity
}
