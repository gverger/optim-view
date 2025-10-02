package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"math"
	"path"
	"path/filepath"
	"strings"

	"fmt"
	"io"
	"os"

	"github.com/gverger/optimview/graph"
	"github.com/gverger/optimview/systems"
	jsoniter "github.com/json-iterator/go"

	"github.com/iancoleman/orderedmap"
	"github.com/phuslu/log"
)

func loadSearchTree(reader io.Reader) systems.SearchTree {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	dec := json.NewDecoder(reader)

	var tree Tree
	MustSucceed(dec.Decode(&tree))

	log.Info().Int("nodes", len(tree.Nodes)).Msg("Tree loaded")

	return systems.SearchTree{
		Tree:   tree.ToGraph(),
		Shapes: tree.Shapes(),
	}
}

func loadTarTrees(filename string) map[string]systems.SearchTree {
	file := Must(os.Open(filename))
	defer file.Close()

	log.Info().Str("file", filename).Msg("Opening file")

	trees := make(map[string]systems.SearchTree)
	gzipReader := Must(gzip.NewReader(file))
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		MustSucceed(err)

		if header.Typeflag == tar.TypeReg && path.Ext(header.Name) == ".json" {
			log.Info().Str("filename", header.Name).Msg("reading file")
			trees[header.Name[:len(header.Name)-5]] = loadSearchTree(tarReader)
		} else {
			log.Info().Str("filename", header.Name).Msg("skipping non json entry")
		}
	}
	return trees
}

func loadSearchTrees(filename string) map[string]systems.SearchTree {
	if path.Ext(filename) == ".tgz" || strings.HasSuffix(filename, ".tar.gz") {
		return loadTarTrees(filename)
	}

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

	trees := make(map[string]systems.SearchTree, 1)
	key := filepath.Base(filename)
	key = strings.TrimSuffix(key, ".gz")
	key = strings.TrimSuffix(key, ".json")
	trees[key] = loadSearchTree(reader)
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
	Id        int
	X         float32
	Y         float32
	FillColor string
}

type TNode struct {
	Id       uint64
	ParentId int64
	Plot     []ShapePos
	Data     orderedmap.OrderedMap
}

type Tree struct {
	Init  []ShapeList
	Name  string `json:"Name"`
	Nodes []*TNode
}

func (t Tree) ToGraph() *GraphView {
	g := graph.NewGraph[*DisplayableNode, uint64](func(n *DisplayableNode) uint64 { return n.Id })

	mapper := make(map[uint64]uint64)
	var index uint64
	for _, n := range t.Nodes {
		if n == nil && index == 0 {
			g.AddNode(&DisplayableNode{Id: 0, Text: "root"})
			mapper[0] = 0
			index++
			continue
		}
		if n == nil {
			continue
		}
		shapeTransforms := make([]ShapeTransform, 0, len(n.Plot))
		minX := float32(math.MaxFloat32)
		minY := float32(math.MaxFloat32)
		for _, p := range n.Plot {
			shapeTransforms = append(shapeTransforms, ShapeTransform{
				Id:        p.Id,
				X:         p.X,
				Y:         p.Y,
				Highlight: p.FillColor == "green",
			})
			minX = min(minX, p.X)
			minY = min(minY, p.Y)
		}
		for i := range shapeTransforms {
			shapeTransforms[i].X -= minX
			shapeTransforms[i].Y -= minY
		}

		g.AddNode(&DisplayableNode{Id: n.Id, Text: nodeDetailsText(*n), Transform: shapeTransforms})
		mapper[n.Id] = index
		index++
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

func (t Tree) Shapes() []systems.ShapeDefinition {
	shapes := make([]systems.ShapeDefinition, 0, len(t.Init))
	for iInit, s := range t.Init {
		polygons := make([]systems.DrawableShape, 0)
		minX := float32(math.MaxFloat32)
		minY := float32(math.MaxFloat32)
		maxX := float32(-math.MaxFloat32)
		maxY := float32(-math.MaxFloat32)
		for iShape, d := range s {
			polygon := make([]systems.Position, 0, len(d.Shape)+1)
			open := true
			if len(d.Shape) == 0 {
				log.Fatal().Int("object index", iInit).Int("shape index", iShape).Msg("shape has no edge")
			} else {
				e := d.Shape[0]
				polygon = append(polygon, systems.Position{X: float64(e.Start.X), Y: float64(e.Start.Y)})
				minX = min(minX, e.Start.X)
				minY = min(minY, e.Start.Y)
				maxX = max(maxX, e.Start.X)
				maxY = max(maxY, e.Start.Y)

				shape := d.Shape
				if shape[0].Start == shape[len(shape)-1].End {
					shape = shape[:len(shape)-1]
					open = false
				}

				for i, e := range shape {
					if i < len(shape)-1 && e.End != shape[(i+1)%len(shape)].Start {
						log.Fatal().Interface("shape", shape).Int("index", i).Msg("shape should be closed")
					}
					polygon = append(polygon, systems.Position{X: float64(e.End.X), Y: float64(e.End.Y)})
					minX = min(minX, e.End.X)
					minY = min(minY, e.End.Y)
					maxX = max(maxX, e.End.X)
					maxY = max(maxY, e.End.Y)
				}
			}

			shape := systems.DrawableShape{Open: open, Points: polygon, Color: d.FillColor}
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
	return shapes
}

func nodeDetailsText(n TNode) string {
	sep := ""
	sb := strings.Builder{}
	for _, key := range n.Data.Keys() {
		value, ok := n.Data.Get(key)
		if ok {
			sb.WriteString(fmt.Sprintf("%s%v: %v", sep, key, value))
			sep = "\n"
		}
	}

	return fmt.Sprintf(sb.String())
}
