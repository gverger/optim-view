package main

import (
	"fmt"
	"time"

	"github.com/nikolaydubina/go-graph-layout/layout"
)

type Node struct {
	Id        string   `json:"id"`
	ParentIds []string `json:"parentId"`
	Info      string   `json:"info"`
	ShortInfo string   `json:"shortInfo"`
}

type Input struct {
	Nodes []Node `json:"nodes"`
}

func PlaceNodes(input Input) layout.Graph {
	g := layout.Graph{
		Edges: make(map[[2]uint64]layout.Edge),
		Nodes: make(map[uint64]layout.Node),
	}

	indices := make(map[string]uint64)

	for i, node := range input.Nodes {
		index := uint64(i)
		indices[node.Id] = index
		g.Nodes[index] = layout.Node{
			W: 50,
			H: 50,
		}
	}

	for _, node := range input.Nodes {
		for _, pId := range node.ParentIds {
			g.Edges[[2]uint64{indices[pId], indices[node.Id]}] = layout.Edge{}
		}
	}

	gl := layout.SugiyamaLayersStrategyGraphLayout{
		CycleRemover:   layout.NewSimpleCycleRemover(),
		LevelsAssigner: layout.NewLayeredGraph,
		OrderingAssigner: layout.WarfieldOrderingOptimizer{
			Epochs:                   500,
			LayerOrderingInitializer: layout.BFSOrderingInitializer{},
			LayerOrderingOptimizer: layout.CompositeLayerOrderingOptimizer{
				Optimizers: []layout.LayerOrderingOptimizer{
					layout.WMedianOrderingOptimizer{},
					// layout.SwitchAdjacentOrderingOptimizer{},
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

	return g
}

func main() {
	// input := readInput("./data/small.json")
	// input := readInput("/tmp/input.json")
	// input := Must(readJsonL("/tmp/input.jsonl.json"))
	// input := Must(readJsonL("../go-graph-layout/layout/testdata/brandeskopf.jsonl"))
	searches := loadSearchTree("search_tree.json")
	input := searches["guide_0_d_0"].ToInput()

	// fmt.Println("Generating input")
	// input := GenerateDeepInput(10000)
	//
	// // input.Nodes[2].ParentIds = append(input.Nodes[2].ParentIds, input.Nodes[7].Id)
	saveInput("input-trees.json", input)
	saveJsonL("input.jsonl", input)
	// // g := PlaceNodes(input)
	start := time.Now()
	g := PlaceNodes(input)
	fmt.Println("Total =", time.Since(start))
	//
	// // fmt.Printf("Input: %#+v\n", input)
	// // fmt.Printf("Graph: %#+v\n", g)
	//
	runVisu(input, g)
}
