package systems

import (
	"context"
	"fmt"

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

// Close implements System.
func (c *Initializer) Close() {
}

// Initialize implements System.
func (c *Initializer) Initialize(w *ecs.World) {
	gridResource := generic.NewResource[Grid](w)
	grid := gridResource.Get()
	nodes := generic.NewMap6[Position, Node, VisibleElement, Velocity, Shape, Target2](w)
	edges := generic.NewMap2[Edge, VisibleElement](w)

	nodeLookup := make(map[uint64]ecs.Entity, 0)

	graph := c.tree.Tree

	for i, n := range graph.Nodes {
		e := nodes.NewWith(
			&Position{
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
			&Target2{},
		)
		nodeLookup[n.Id] = e
		grid.AddEntity(e, GridPos{})
	}

	for i, e := range graph.Edges {
		src := nodeLookup[graph.Nodes[i].Id]
		for j := range e {
			dst := nodeLookup[graph.Nodes[j].Id]

			edges.NewWith(
				&Edge{From: src, To: dst},
				&VisibleElement{},
			)
		}
	}

	shapes := generic.NewResource[[]ShapeDefinition](w)
	shapes.Add(&c.tree.Shapes)

	textures := generic.NewResource[[]rl.RenderTexture2D](w)
	t := make([]rl.RenderTexture2D, 0)
	textures.Add(&t)

	mappings := generic.NewResource[Mappings](w)
	mappings.Add(&Mappings{
		nodeLookup: nodeLookup,
	})

	camera := generic.NewResource[CameraHandler](w)
	cameraHandler := NewCameraHandler()
	camera.Add(&cameraHandler)
}

func (i *Initializer) Update(ctx context.Context, w *ecs.World) {}

var _ System = &Initializer{}
