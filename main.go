package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/nikolaydubina/go-graph-layout/layout"
)

type Node struct {
	Id        string `json:"id"`
	ParentId  string `json:"parentId"`
	Info      string `json:"info"`
	ShortInfo string `json:"shortInfo"`
}

type Input struct {
	Nodes []Node `json:"nodes"`
}

func readInput(filename string) Input {
	jsonFile := Must(os.Open(filename))
	defer jsonFile.Close()

	content := Must(io.ReadAll(jsonFile))
	var input Input
	json.Unmarshal(content, &input)

	return input
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
			W: 100,
			H: 50,
		}
	}

	for _, node := range input.Nodes {
		if node.ParentId != node.Id {
			g.Edges[[2]uint64{indices[node.ParentId], indices[node.Id]}] = layout.Edge{}
		}
	}

	gl := layout.SugiyamaLayersStrategyGraphLayout{
		CycleRemover:     layout.NewSimpleCycleRemover(),
		LevelsAssigner:   layout.NewLayeredGraph,
		OrderingAssigner: layout.WarfieldOrderingOptimizer{
			Epochs:                   10,
			LayerOrderingInitializer: layout.BFSOrderingInitializer{},
			LayerOrderingOptimizer: layout.CompositeLayerOrderingOptimizer{
				Optimizers: []layout.LayerOrderingOptimizer{
					// layout.WMedianOrderingOptimizer{},
					// layout.SwitchAdjacentOrderingOptimizer{},
				},
			},
		}.Optimize,
		NodesHorizontalCoordinatesAssigner: layout.BrandesKopfLayersNodesHorizontalAssigner{
			Delta: 125,
		},
		NodesVerticalCoordinatesAssigner: layout.BasicNodesVerticalCoordinatesAssigner{
			MarginLayers:   125,
			FakeNodeHeight: 125,
		},
		EdgePathAssigner: layout.StraightEdgePathAssigner{}.UpdateGraphLayout,
	}
	gl.UpdateGraphLayout(g)

	return g
}

func main() {
	// input := readInput("./data/small.json")
	input := GenerateInput(3000)
	g := PlaceNodes(input)

	// fmt.Printf("Input: %#+v\n", input)
	// fmt.Printf("Graph: %#+v\n", g)

	runVisu(input, g)
}
