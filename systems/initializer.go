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
	nodes := generic.NewMap5[Position, Node, Velocity, Shape, Target](w)
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
				SizeX: 25,
				SizeY: 25,
			},
			&Velocity{
				Dx: 0,
				Dy: 0,
			},
			&Shape{
				Points: []Position{
					{0, 0},
					{25, 0},
					{25, 25},
					{0, 25},
				},
			},
			&Target{},
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

	mappings := generic.NewResource[Mappings](w)
	mappings.Add(&Mappings{
		nodeLookup: nodeLookup,
	})

	mouse := generic.NewResource[Mouse](w)
	mouse.Add(&Mouse{})
}

func (i *Initializer) Update(w *ecs.World) {}

var _ System = &Initializer{}
