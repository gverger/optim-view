package systems

import (
	"context"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/ark/ecs"
)

func NewDrawGraph(font rl.Font, nbNodes int) *DrawGraph {
	return &DrawGraph{
		Nodes: NewDrawNodes(font, nbNodes),
		Edges: NewDrawEdges(font),
	}
}

type DrawGraph struct {
	Nodes *DrawNodes
	Edges *DrawEdges
}

// Close implements System.
func (d *DrawGraph) Close() {
	d.Nodes.Close()
	d.Edges.Close()
}

func (d *DrawGraph) Initialize(w *ecs.World) {
	d.Nodes.Initialize(w)
	d.Edges.Initialize(w)
}

func (d *DrawGraph) Update(ctx context.Context, w *ecs.World) {
	d.Nodes.Update(ctx, w)
	d.Edges.Update(ctx, w)
}

var _ System = &DrawGraph{}
