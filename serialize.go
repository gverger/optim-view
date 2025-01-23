package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"math"
	"path"

	"encoding/json"

	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/gverger/optimview/graph"
	"github.com/gverger/optimview/systems"
	jsoniter "github.com/json-iterator/go"

	"github.com/phuslu/log"
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

func loadSearchTrees(filename string) map[string]systems.SearchTree {
	file := Must(os.Open(filename))
	defer file.Close()

	log.Info().Str("file", filename).Msg("Opening file")

	var reader io.Reader
	if path.Ext(filename) == ".gz" {
		gzipReader := Must(gzip.NewReader(file))
		defer gzipReader.Close()
		reader = bufio.NewReader(gzipReader)
	} else {
		reader = bufio.NewReader(file)
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	dec := json.NewDecoder(reader)

	var st []Tree
	MustSucceed(dec.Decode(&st))

	if len(st) == 0 {
		log.Fatal().Msg("no tree")
	}
	trees := make(map[string]systems.SearchTree, len(st))
	for _, tree := range st {
		shapes := make([]systems.ShapeDefinition, 0, len(tree.Init))
		for _, s := range tree.Init {
			polygons := make([]systems.DrawableShape, 0)
			minX := float32(math.MaxFloat32)
			minY := float32(math.MaxFloat32)
			maxX := float32(-math.MaxFloat32)
			maxY := float32(-math.MaxFloat32)
			for _, d := range s {
				polygon := make([]systems.Position, 0, len(d.Shape))
				for i, e := range d.Shape {
					if e.End.X != d.Shape[(i+1)%len(d.Shape)].Start.X {
						log.Fatal().Interface("shape", d.Shape).Int("index", i).Msg("edges")
					}
					polygon = append(polygon, systems.Position{X: float64(e.Start.X), Y: float64(e.Start.Y)})
					minX = min(minX, e.Start.X)
					minY = min(minY, e.Start.Y)
					maxX = max(maxX, e.Start.X)
					maxY = max(maxY, e.Start.Y)
				}

				shape := systems.DrawableShape{Points: polygon, Color: d.FillColor}
				for _, edges := range d.Holes {
					hole := make([]systems.Position, 0, len(edges))
					for i, e := range edges {
						if e.End.X != edges[(i+1)%len(edges)].Start.X {
							log.Fatal().Interface("holes", edges).Int("index", i).Msg("edges")
						}
						hole = append(hole, systems.Position{X: float64(e.Start.X), Y: float64(e.Start.Y)})
					}
					shape.Holes = append(shape.Holes, hole)
				}
				polygons = append(polygons, shape)
			}
			shapes = append(shapes, systems.ShapeDefinition{
				Shapes: polygons,
				MinX:   minX,
				MinY:   minY,
				MaxX:   maxX,
				MaxY:   maxY,
			})
		}

		trees[tree.Name] = systems.SearchTree{
			Tree:   tree.ToGraph().StripNodesWithoutChildren(),
			Shapes: shapes,
		}
	}
	return trees
}

type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type Edge struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
	Type  string   `json:"type"`
}

type ShapeList []ShapeDesc
type ShapeDesc struct {
	FillColor string
	Shape     []Edge   `json:"Shape"`
	Holes     [][]Edge `json:"Holes"`
}

type ShapePos struct {
	Id int
	X  float32
	Y  float32
}

type TNode struct {
	Id                 uint64
	GuideArea          float32
	ItemArea           float32
	ItemConvexHullArea float32
	NumberOfBins       uint32
	NumberOfItems      uint32
	ParentId           int64
	Profit             float32
	TrapezoidSetId     int
	X                  float32
	Y                  float32
	Plot               []ShapePos
}

type Tree struct {
	Init  []ShapeList
	Name  string `json:"Name"`
	Nodes []*TNode
}

func (t Tree) ToGraph() *GraphView {
	g := graph.NewGraph[*DisplayableNode, uint64](func(n *DisplayableNode) uint64 { return n.Id })

	mapper := make(map[uint64]uint64)
	for i, n := range t.Nodes {
		if n == nil && i == 0 {
			g.AddNode(&DisplayableNode{Id: uint64(i), Text: "root"})
			mapper[0] = uint64(i)
			continue
		}
		shapeTransforms := make([]ShapeTransform, 0, len(n.Plot))
		for _, p := range n.Plot {
			shapeTransforms = append(shapeTransforms, ShapeTransform{
				Id:        p.Id,
				X:         p.X,
				Y:         p.Y,
				Highlight: p.X == n.X && p.Y == n.Y,
			})
		}
		g.AddNode(&DisplayableNode{Id: uint64(i), Text: fmt.Sprintf("Profit=%v", n.Profit), Transform: shapeTransforms})
		mapper[n.Id] = uint64(i)
	}

	for _, n := range t.Nodes {
		if n == nil {
			continue
		}
		parent := uint64(n.ParentId)
		if n.ParentId == -1 {
			parent = 0
		}
		g.AddEdgeId(mapper[parent], mapper[n.Id])
	}

	return g
}
