package systems

import "github.com/gverger/optimview/graph"

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

type ShapeDefinition struct {
	Shapes []Shape
}

type SearchTree struct {
	Tree   *graph.Graph[*DisplayableNode, uint64]
	Shapes []ShapeDefinition
}
