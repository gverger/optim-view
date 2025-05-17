package main

import (
	"os"

	"github.com/gverger/optimview/graph"
	"github.com/gverger/optimview/systems"
	"github.com/phuslu/log"
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

type Configuration struct {
	DebugMode bool
}

var config = Configuration{
	DebugMode: true,
}

func main() {

	for _, o := range os.Args[1:] {
		if o == "--debug" {
			config.DebugMode = true
		}
	}

	log.DefaultLogger = log.Logger{
		Level: log.DebugLevel,
		TimeFormat: "15:04:05",
		Writer: &log.ConsoleWriter{
			ColorOutput:    true,
			QuoteString:    true,
			EndWithMessage: false,
		},
	}
	runVisu(Input{})
}
