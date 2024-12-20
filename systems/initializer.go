package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graph"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewInitializer(g graph.Graph[*DisplayableNode, uint64]) *Initializer {
	return &Initializer{g: g}
}

type Initializer struct {
	g graph.Graph[*DisplayableNode, uint64]
}

// Initialize implements System.
func (c *Initializer) Initialize(w *ecs.World) {
	nodes := generic.NewMap3[Position, Node, Velocity](w)
	edges := generic.NewMap1[Edge](w)

	nodeLookup := make(map[uint64]ecs.Entity, 0)

	for _, n := range c.g.Nodes {
		e := nodes.NewWith(
			&Position{
				// X: float64(n.XY[0]),
				// Y: float64(n.XY[1]),
			}, &Node{
				color: rl.Gray,
				Text:  n.Text,
				SizeX: 125,
				SizeY: 125,
			},
			&Velocity{
				Dx: 0,
				Dy: 0,
			},
		)
		nodeLookup[n.Id] = e
	}

	for i, e := range c.g.Edges {
		src := nodeLookup[c.g.Nodes[i].Id]
		for j := range e {
			dst := nodeLookup[c.g.Nodes[j].Id]
			edges.NewWith(&Edge{From: src, To: dst})
		}
	}
}

func (i *Initializer) Update(w *ecs.World) {}

var _ System = &Initializer{}
