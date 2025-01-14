package main

import (
	"github.com/gverger/optimview/graph"
	"github.com/gverger/optimview/systems"
)

type DisplayableNode = systems.DisplayableNode
type ShapeTransform = systems.ShapeTransform

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
	Trees map[string]systems.SearchTree
}

type InputTree struct {
	Nodes []Node `json:"nodes"`
}


func main() {
	runVisu(Input{})
}
