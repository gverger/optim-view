package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

func readInput(filename string) InputTree {
	jsonFile := Must(os.Open(filename))
	defer jsonFile.Close()

	content := Must(io.ReadAll(jsonFile))
	var input InputTree
	json.Unmarshal(content, &input)

	return input
}

func saveInput(filename string, input InputTree) {
	fmt.Println("SAVING INPUT", filename)
	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	defer file.Close()
	encoder := json.NewEncoder(file)
	MustSucceed(encoder.Encode(input))
}

type JsonEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func saveJsonL(filename string, input InputTree) {
	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, n := range input.Nodes {
		for _, pId := range n.ParentIds {
			edge := JsonEdge{
				From: pId,
				To:   n.Id,
			}
			encoder.Encode(edge)
		}
	}
}

func readJsonL(filename string) (InputTree, error) {
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
	return InputTree{Nodes: nodes}, scanner.Err()

}

type Trace struct {
	Nodes []TreeNode
}

func (s Trace) ToInput() InputTree {
	nodes := make([]Node, 0, len(s.Nodes)+1)
	nodes = append(nodes, Node{
		Id:        "0",
		ParentIds: []string{},
		Info:      "Root",
		ShortInfo: "Root",
	})

	for _, n := range s.Nodes {
		node := n.ToNode()
		if len(node.ParentIds) == 0 {
			node.ParentIds = append(node.ParentIds, "0")
		}
		nodes = append(nodes, node)
	}
	return InputTree{
		Nodes: nodes,
	}
}

type TreeNode struct {
	ID       int
	ParentID int

	GuideArea          float32
	ItemArea           float32
	ItemConvexHullArea float32
	NumberOfBins       int
	Profit             float32

	TrapezoidSetId int
	X              float32
	Y              float32

	SVG string
}

func (s TreeNode) ToNode() Node {
	parentIds := make([]string, 0)
	if s.ParentID != -1 {
		parentIds = append(parentIds, strconv.Itoa(s.ParentID))
	}

	info := fmt.Sprintf("ID: %d\n", s.ID)
	info += fmt.Sprintf("Bins: %d\n", s.NumberOfBins)
	info += fmt.Sprintf("Profit: %v\n", s.Profit)
	info += fmt.Sprintf("Guide Area: %v\n", s.GuideArea)
	info += fmt.Sprintf("Item Area: %v\n", s.ItemArea)
	info += fmt.Sprintf("Item Convex Hull Area: %v\n", s.ItemConvexHullArea)

	short := fmt.Sprintf("ID: %d\n", s.ID)
	short += fmt.Sprintf("Profit: %v\n", s.Profit)

	return Node{
		Id:        strconv.Itoa(s.ID),
		ParentIds: parentIds,
		Info:      info,
		ShortInfo: short,
		SvgImage:  s.SVG,
	}
}

func decodeTreeNodes(r io.Reader) (map[string]Trace, error) {
	dec := json.NewDecoder(r)

	// Expect start of object as the first token.
	t := Must(dec.Token())
	if t != json.Delim('{') {
		return nil, fmt.Errorf("expected {, got %v", t)
	}

	res := make(map[string]Trace)

	// While there are more tokens in the JSON stream...
	for dec.More() {

		// Read the key.
		t := Must(dec.Token())
		key := t.(string)
		fmt.Println("key", key)

		// read open bracket
		t = Must(dec.Token())

		trace := Trace{
			Nodes: make([]TreeNode, 0, 100),
		}

		for dec.More() {
			var node TreeNode
			MustSucceed(dec.Decode(&node))

			if node.ID != 0 {
				trace.Nodes = append(trace.Nodes, node)
			}
		}

		// read closing bracket
		t = Must(dec.Token())

		res[key] = trace

		// Add your code to process the key and value here.
		// fmt.Printf("key %q, value %#v\n", key, value)
	}
	t = Must(dec.Token())

	return res, nil
}

func loadSearchTree(filename string) map[string]Trace {
	file := Must(os.Open(filename))
	defer file.Close()

	return Must(decodeTreeNodes(bufio.NewReader(file)))
}
