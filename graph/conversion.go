package graph

import (
	"github.com/nikolaydubina/go-graph-layout/layout"
)

type ConvertedLayoutGraph[ID comparable] struct {
	Mappings    map[ID]uint64
	LayoutGraph layout.Graph
	Layers      layout.LayeredGraph
}

type Position struct {
	X int
	Y int
}

func ComputeLayeredCoordinates[Node any, ID comparable](input Graph[Node, ID]) map[ID]Position {
	c := ConvertToLayoutGraph(input)

	x := layout.BrandesKopfLayersNodesHorizontalAssigner{Delta: 125}.NodesHorizontalCoordinates(c.LayoutGraph, c.Layers)
	y := layout.BasicNodesVerticalCoordinatesAssigner{
		MarginLayers:   125,
		FakeNodeHeight: 50,
	}.NodesVerticalCoordinates(c.LayoutGraph, c.Layers)

	positions := make(map[ID]Position, len(x))
	for id, i := range c.Mappings {
		positions[id] = Position{
			X: x[i],
			Y: y[i],
		}
	}

	return positions
}

func ConvertToLayoutGraph[Node any, ID comparable](input Graph[Node, ID]) ConvertedLayoutGraph[ID] {
	g := layout.Graph{
		Edges: make(map[[2]uint64]layout.Edge),
		Nodes: make(map[uint64]layout.Node),
	}

	mapping := make(map[ID]uint64, len(input.Nodes))

	for i, n := range input.Nodes {
		index := uint64(i)
		mapping[input.NodeID(n)] = index
		g.Nodes[index] = layout.Node{
			W: 50,
			H: 50,
		}
	}

	for a, dst := range input.Edges {
		aId := input.NodeID(input.Nodes[a])
		for b := range dst {
			bId := input.NodeID(input.Nodes[b])
			g.Edges[[2]uint64{mapping[aId], mapping[bId]}] = layout.Edge{}
		}
	}

	layers := layout.NewLayeredGraph(g)

	return ConvertedLayoutGraph[ID]{
		Mappings:    mapping,
		LayoutGraph: g,
		Layers:      layers,
	}
}
