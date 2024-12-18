package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graph"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
	"github.com/nikolaydubina/go-graph-layout/layout"
)

func NewInitializer(g graph.Graph[*DisplayableNode, uint64], l layout.Graph) *Initializer {
	return &Initializer{g: g, l: l}
}

type Initializer struct {
	g graph.Graph[*DisplayableNode, uint64]
	l layout.Graph
}

// Initialize implements System.
func (i *Initializer) Initialize(w *ecs.World) {
	nodes := generic.NewMap3[Position, Node, Velocity](w)
	edges := generic.NewMap1[Edge](w)
	path := generic.NewMap3[Position, Velocity, BelongsTo](w, generic.T[BelongsTo]())

	nodeLookup := make(map[uint64]ecs.Entity, 0)
	edgesLookup := make(map[[2]uint64]ecs.Entity, 0)

	for id, n := range i.l.Nodes {
		e := nodes.NewWith(
			&Position{
				// X: float64(n.XY[0]),
				// Y: float64(n.XY[1]),
			}, &Node{
				color: rl.Gray,
				Text:  i.g.NodeForId(id).Text,
				SizeX: float64(n.W),
				SizeY: float64(n.H),
			},
			&Velocity{
				Dx: 0,
				Dy: 0,
			},
		)
		nodeLookup[id] = e
	}

	for id, e := range i.l.Edges {
		entity := edges.NewWith(
			&Edge{
				From: nodeLookup[id[0]],
				To:   nodeLookup[id[1]],
			},
		)
		points := make([]ecs.Entity, 0, len(e.Path))
		pathQuery := path.NewBatchQ(len(e.Path), entity)
		for pathQuery.Next() {
			points = append(points, pathQuery.Entity())
		}

		edgesLookup[id] = entity
	}

	ecs.AddResource(w, &Mappings{nodeLookup: nodeLookup, edgeLookup: edgesLookup})
}

func (i *Initializer) Update(w *ecs.World) {}

var _ System = &Initializer{}
