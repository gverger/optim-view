package graph

import (
	"sort"
	"time"

	"github.com/gverger/go-graph-layout/layout"
	"github.com/phuslu/log"
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
	log.Info().Msg("Computing coordinates")
	now := time.Now()
	c := ConvertToLayoutGraph(input)
	log.Info().Dur("duration", time.Since(now)).Msg("converted to layout")
	now = time.Now()

	x := layout.BrandesKopfLayersNodesHorizontalAssigner{Delta: 120, TopDownOnly: true}.NodesHorizontalCoordinates(c.LayoutGraph, c.Layers)
	log.Info().Dur("duration", time.Since(now)).Msg("brandeskopf")
	now = time.Now()

	y := layout.BasicNodesVerticalCoordinatesAssigner{
		MarginLayers: 60,
	}.NodesVerticalCoordinates(c.LayoutGraph, c.Layers)
	log.Info().Dur("duration", time.Since(now)).Msg("vertical")
	now = time.Now()

	positions := make(map[ID]Position, len(x))
	for id, i := range c.Mappings {
		positions[id] = Position{
			X: x[i],
			Y: y[i],
		}
	}

	log.Info().Dur("duration", time.Since(now)).Msg("coordinates computed")
	return positions
}

func ConvertToLayoutGraph[Node any, ID comparable](input Graph[Node, ID]) ConvertedLayoutGraph[ID] {
	g := layout.Graph{
		Edges: make(map[[2]uint64]layout.Edge, len(input.Edges)),
		Nodes: make(map[uint64]layout.Node, len(input.Nodes)),
	}

	mapping := make(map[ID]uint64, len(input.Nodes))

	for i, n := range input.Nodes {
		index := uint64(i)
		mapping[input.NodeID(n)] = index
		g.Nodes[index] = layout.Node{
			W: 110,
			H: 90,
		}
	}

	parent := make(map[uint64]uint64, 0)
	for a, dst := range input.Edges {
		aId := input.NodeID(input.Nodes[a])
		for b := range dst {
			bId := input.NodeID(input.Nodes[b])
			g.Edges[[2]uint64{mapping[aId], mapping[bId]}] = layout.Edge{}
			parent[mapping[bId]] = mapping[aId]
		}
	}

	layers := layout.NewLayeredGraph(g)

	ll := layers.Layers()
	for i := range ll {
		sort.Slice(ll[i], func(a, b int) bool {
			pa := layers.NodePosition[parent[ll[i][a]]].Order
			pb := layers.NodePosition[parent[ll[i][b]]].Order
			if pa < pb {
				return true
			}
			if pa > pb {
				return false
			}
			return ll[i][a] < ll[i][b]
		})
		for j, v := range ll[i] {
			layers.NodePosition[v] = layout.LayerPosition{Layer: i, Order: j}
		}
	}

	return ConvertedLayoutGraph[ID]{
		Mappings:    mapping,
		LayoutGraph: g,
		Layers:      layers,
	}
}
