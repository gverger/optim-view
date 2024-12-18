package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/gverger/optimview/graph"
	"github.com/gverger/optimview/systems"
	"github.com/nikolaydubina/go-graph-layout/layout"
)

type DisplayableNode = systems.DisplayableNode

type GraphView = graph.Graph[*DisplayableNode, uint64]

type Node struct {
	Id        string   `json:"id"`
	ParentIds []string `json:"parentIds"`
	Info      string   `json:"info"`
	ShortInfo string   `json:"shortInfo"`
	SvgImage  string   `json:"svg"`
	Hidden    bool     `json:"hidden"`
}

type Input struct {
	Trees   map[string]*GraphView
	Layouts map[string]layout.Graph
	Layers  map[string]layout.LayeredGraph
}

type InputTree struct {
	Nodes []Node `json:"nodes"`
}

func (t InputTree) ToGraph() *GraphView {
	g := graph.NewGraph[*DisplayableNode, uint64](func(n *DisplayableNode) uint64 { return n.Id })

	mapper := make(map[string]uint64)
	for i, n := range t.Nodes {
		g.AddNode(&DisplayableNode{Id: uint64(i), Text: n.ShortInfo})
		mapper[n.Id] = uint64(i)
	}

	for _, n := range t.Nodes {
		for _, pId := range n.ParentIds {
			g.AddEdgeId(mapper[pId], mapper[n.Id])
		}
	}

	return g
}

func PlaceNodes(input *GraphView) (layout.LayeredGraph, layout.Graph) {
	g := layout.Graph{
		Edges: make(map[[2]uint64]layout.Edge),
		Nodes: make(map[uint64]layout.Node),
	}

	for i := range input.Nodes {
		index := uint64(i)
		g.Nodes[index] = layout.Node{
			W: 50,
			H: 50,
		}
	}

	for a, dst := range input.Edges {
		for b := range dst {
			g.Edges[[2]uint64{input.Nodes[a].Id, input.Nodes[b].Id}] = layout.Edge{}
		}
	}

	var layeredGraph layout.LayeredGraph

	var layerer = func(g layout.Graph) layout.LayeredGraph {
		layeredGraph = layout.NewLayeredGraph(g)
		return layeredGraph
	}

	gl := layout.SugiyamaLayersStrategyGraphLayout{
		CycleRemover:   layout.NewSimpleCycleRemover(),
		LevelsAssigner: layerer,
		OrderingAssigner: layout.WarfieldOrderingOptimizer{
			Epochs:                   500,
			LayerOrderingInitializer: layout.BFSOrderingInitializer{},
			LayerOrderingOptimizer: layout.CompositeLayerOrderingOptimizer{
				Optimizers: []layout.LayerOrderingOptimizer{
					// layout.WMedianOrderingOptimizer{},
					layout.SwitchAdjacentOrderingOptimizer{},
				},
			},
		}.Optimize,
		NodesHorizontalCoordinatesAssigner: layout.BrandesKopfLayersNodesHorizontalAssigner{
			Delta: 125,
		},
		NodesVerticalCoordinatesAssigner: layout.BasicNodesVerticalCoordinatesAssigner{
			MarginLayers:   125,
			FakeNodeHeight: 50,
		},
		EdgePathAssigner: layout.StraightEdgePathAssigner{}.UpdateGraphLayout,
	}
	gl.UpdateGraphLayout(g)

	return layeredGraph, g
}

type IdOrder struct {
	g *GraphView
}

// Init implements layout.LayerOrderingInitializer.
func (o IdOrder) Init(segments map[[2]uint64]bool, layers [][]uint64) {
	for l := range layers {
		sort.Slice(layers[l], func(i, j int) bool {
			return o.g.Nodes[layers[l][i]].Id < o.g.Nodes[layers[l][j]].Id
		})
	}
}

var _ layout.LayerOrderingInitializer = IdOrder{}

func runSearchTrees() {
	searches := loadSearchTree("search_tree.json")

	trees := make(map[string]*GraphView)
	for key, tree := range searches {
		trees[key] = tree.ToInput().ToGraph()
	}

	start := time.Now()
	layouts := make(map[string]layout.Graph)
	layers := make(map[string]layout.LayeredGraph)
	for k, t := range trees {
		layers[k], layouts[k] = PlaceNodes(t)
	}
	fmt.Println("Total =", time.Since(start))
	runVisu(Input{Trees: trees, Layouts: layouts, Layers: layers})
}

func main() {
	// runSearchTrees()
	// return
	// input := readInput("./data/small.json")
	// input := readInput("brandeskopf.json")
	// input := Must(readJsonL("../go-graph-layout/layout/testdata/brandeskopf.jsonl"))

	fmt.Println("Generating input")
	// fmt.Println("input node", len(input.Nodes))
	input := GenerateDeepInput(100)
	//
	// input.Nodes[27].ParentIds = append(input.Nodes[27].ParentIds, input.Nodes[2].Id)
	// saveInput("input-trees.json", input)
	// saveJsonL("input.jsonl", input)
	g := input.ToGraph()
	start := time.Now()
	layer, layout := PlaceNodes(g)
	fmt.Println("Total =", time.Since(start))
	//
	// // fmt.Printf("Input: %#+v\n", input)
	// // fmt.Printf("Graph: %#+v\n", g)
	//
	runSingleVisu(g, layer, layout)
}
