package systems

import (
	"context"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graph"
	"github.com/mlange-42/ark/ecs"
)

func NewInitializer(tree SearchTree, initialPositions map[uint64]graph.Position) *Initializer {
	return &Initializer{tree: tree, initialPositions: initialPositions}
}

type Initializer struct {
	tree             SearchTree
	initialPositions map[uint64]graph.Position
}

// Close implements System.
func (c *Initializer) Close() {
}

// Initialize implements System.
func (c *Initializer) Initialize(w *ecs.World) {
	gridResource := ecs.NewResource[Grid](w)
	grid := gridResource.Get()
	nodes := ecs.NewMap5[Position, Node, VisibleElement, Velocity, Shape](w)
	// edges := ecs.NewMap2[Edge, VisibleElement](w)
	// start := ecs.NewMap1[StartOf](w)
	end := ecs.NewMap2[Parent, ChildOf](w)

	nodeLookup := make(map[uint64]ecs.Entity, 0)

	graph := c.tree.Tree

	for i, n := range graph.Nodes {
		pos := c.initialPositions[graph.NodeID(n)]
		e := nodes.NewEntity(
			&Position{
				X: float64(pos.X),
				Y: float64(pos.Y),
				// X: float64(n.XY[0]),
				// Y: float64(n.XY[1]),
			},
			&Node{
				color:           rl.Gray,
				Title:           fmt.Sprintf("Node %v", n.Id),
				Text:            n.Text,
				SizeX:           100,
				SizeY:           100,
				ShapeTransforms: n.Transform,
				idx:             i + 1,
			},
			&VisibleElement{},
			&Velocity{
				Dx: 0,
				Dy: 0,
			},
			&Shape{
				Points: []Position{
					{0, 0},
					{100, 0},
					{100, 100},
					{0, 100},
				},
			},
		)
		nodeLookup[n.Id] = e
		grid.AddEntity(e, GridCoords(pos.X, pos.Y))
	}

	for i, e := range graph.Edges {
		src := nodeLookup[graph.Nodes[i].Id]
		for j := range e {
			dst := nodeLookup[graph.Nodes[j].Id]

			// _ = edges.NewEntity(
			// 	&Edge{From: src, To: dst},
			// 	&VisibleElement{},
			// )

			end.Add(dst, &Parent{parent: src}, &ChildOf{}, ecs.Rel[ChildOf](src))
		}
	}

	shapes := ecs.NewResource[[]ShapeDefinition](w)
	shapes.Add(&c.tree.Shapes)

	textures := ecs.NewResource[[]rl.RenderTexture2D](w)
	t := make([]rl.RenderTexture2D, 0)
	textures.Add(&t)

	mappings := ecs.NewResource[Mappings](w)
	mappings.Add(&Mappings{
		nodeLookup: nodeLookup,
	})

	camera := ecs.NewResource[CameraHandler](w)
	cameraHandler := NewCameraHandler()
	camera.Add(&cameraHandler)
}

func (i *Initializer) Update(ctx context.Context, w *ecs.World) {}

var _ System = &Initializer{}
