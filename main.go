package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

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

func readInput(filename string) Input {
	jsonFile := Must(os.Open(filename))
	defer jsonFile.Close()

	content := Must(io.ReadAll(jsonFile))
	var input Input
	json.Unmarshal(content, &input)

	return input
}

func saveInput(filename string, input Input) {
	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.Encode(input)
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

	return g
}

type JsonEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func readJsonL(filename string) (Input, error) {
	jsonFile := Must(os.Open(filename))
	defer jsonFile.Close()

	nodes := make([]Node, 0)
	nodeIds := make(map[string]bool)
	scanner := bufio.NewScanner(io.Reader(jsonFile))
	for scanner.Scan() {
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		var edge JsonEdge
		decoder.Decode(&edge)

		if !nodeIds[edge.From] {
			nodes = append(nodes, Node{
				Id:        edge.From,
				ParentIds: make([]string, 0),
				Info:      fmt.Sprintf("Node %s", edge.From),
				ShortInfo: fmt.Sprintf("Node %s", edge.From),
			})
			nodeIds[edge.From] = true
		}

		if !nodeIds[edge.To] {
			nodes = append(nodes, Node{
				Id:        edge.To,
				ParentIds: []string{edge.From},
				Info:      fmt.Sprintf("Node %s", edge.To),
				ShortInfo: fmt.Sprintf("Node %s", edge.To),
			})
			nodeIds[edge.To] = true
		} else {
			for _, v := range nodes {
				if v.Id == edge.To {
					v.ParentIds = append(v.ParentIds, edge.From)
				}
			}
		}

	}
	return Input{Nodes: nodes}, scanner.Err()

}

func main() {
	// input := readInput("./data/small.json")
	// input := readInput("/tmp/input.json")
	// input := Must(readJsonL("../go-graph-layout/layout/testdata/brandeskopf.jsonl"))
	// fmt.Println("Generating input")
	input := GenerateDeepInput(50)
	saveInput("/tmp/input.json", input)
	g := PlaceNodes(input)

	// fmt.Printf("Input: %#+v\n", input)
	// fmt.Printf("Graph: %#+v\n", g)

	runVisu(input, g)
}
