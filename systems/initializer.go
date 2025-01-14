package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewInitializer(tree SearchTree) *Initializer {
	return &Initializer{tree: tree}
}

type Initializer struct {
	tree SearchTree
}

// Initialize implements System.
func (c *Initializer) Initialize(w *ecs.World) {
	nodes := generic.NewMap5[Position, Node, Velocity, Shape, Target](w)
	edges := generic.NewMap1[Edge](w)

	nodeLookup := make(map[uint64]ecs.Entity, 0)

	graph := c.tree.Tree

	for _, n := range graph.Nodes {
		e := nodes.NewWith(
			&Position{
				// X: float64(n.XY[0]),
				// Y: float64(n.XY[1]),
			}, &Node{
				color:           rl.Gray,
				Text:            n.Text,
				SizeX:           120,
				SizeY:           90,
				ShapeTransforms: n.Transform,
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

	for i, e := range graph.Edges {
		src := nodeLookup[graph.Nodes[i].Id]
		for j := range e {
			dst := nodeLookup[graph.Nodes[j].Id]
			edges.NewWith(&Edge{From: src, To: dst})
		}
	}

	shapes := generic.NewResource[[]ShapeDefinition](w)
	shapes.Add(&c.tree.Shapes)

	mappings := generic.NewResource[Mappings](w)
	mappings.Add(&Mappings{
		nodeLookup: nodeLookup,
	})
}

func (i *Initializer) Update(w *ecs.World) {}

var _ System = &Initializer{}
