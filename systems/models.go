package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graph"
	"github.com/osuushi/triangulate"
)

type ShapeTransform struct {
	Id int
	X  float32
	Y  float32
}

type DisplayableNode struct {
	Id   uint64
	Text string

	Transform []ShapeTransform
}

type DrawableShape struct {
	Color  string
	Points []Position

	Triangles []*triangulate.Triangle
}

type ShapeDefinition struct {
	Shapes []DrawableShape
	MinX   float32
	MinY   float32
	MaxX   float32
	MaxY   float32

	Texture  rl.RenderTexture2D
	rendered bool
}

type SearchTree struct {
	Tree   *graph.Graph[*DisplayableNode, uint64]
	Shapes []ShapeDefinition
}
