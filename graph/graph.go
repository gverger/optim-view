package graph

import (
	"fmt"
	"iter"

	"github.com/phuslu/log"
)

type Graph[Node any, ID comparable] struct {
	Nodes  []Node
	Lookup map[ID]int
	Edges  []map[int]struct{}

	NodeID func(Node) ID
}

func NewGraph[Node any, ID comparable](idMapper func(Node) ID) *Graph[Node, ID] {
	return &Graph[Node, ID]{Lookup: make(map[ID]int), NodeID: idMapper}
}

func (g Graph[Node, ID]) NodeForId(id ID) Node {
	_, ok := g.Lookup[id]
	Assert(ok, "No node with id %v", id)

	return g.Nodes[g.Lookup[id]]
}

func (g *Graph[Node, ID]) AddNode(n Node) {
	Assert(!g.HasNode(n), "node exists: %v", g.NodeID(n))

	g.addNode(n)
}

func (g *Graph[Node, ID]) addNode(n Node) {
	g.Lookup[g.NodeID(n)] = len(g.Nodes)
	g.Nodes = append(g.Nodes, n)
	g.Edges = append(g.Edges, make(map[int]struct{}))
}

func (g *Graph[Node, ID]) AddEdgeId(a, b ID) {
	g.AddEdge(g.Nodes[g.Lookup[a]], g.Nodes[g.Lookup[b]])
}

func (g *Graph[Node, ID]) AddEdge(a, b Node) {
	if !g.HasNode(a) {
		g.addNode(a)
	}
	if !g.HasNode(b) {
		g.addNode(b)
	}
	Assert(!g.HasEdge(a, b), "edge exists: %v->%v", a, b)
	g.addEdge(a, b)
}

func (g *Graph[Node, ID]) Children(n Node) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		for c := range g.Edges[g.lookup(n)] {
			if !yield(g.Nodes[c]) {
				return
			}
		}
	}
}

func (g Graph[Node, ID]) lookup(n Node) int {
	return g.Lookup[g.NodeID(n)]
}

func (g *Graph[Node, ID]) addEdge(a, b Node) {
	g.Edges[g.lookup(a)][g.lookup(b)] = struct{}{}
}

func (g *Graph[Node, ID]) HasNode(n Node) bool {
	_, ok := g.Lookup[g.NodeID(n)]
	return ok
}

func (g *Graph[Node, ID]) HasEdge(a, b Node) bool {
	if !g.HasNode(a) {
		return false
	}
	if !g.HasNode(b) {
		return false
	}
	_, ok := g.Edges[g.lookup(a)][g.lookup(b)]
	return ok
}

func Assert(condition bool, msg string, args ...any) {
	if condition {
		return
	}

	log.Fatal().Msg(fmt.Sprint("Assertion failed: ", fmt.Sprintf(msg, args...)))
}

func AssertNoErr(err error, msg string) {
	if err == nil {
		return
	}
	log.Fatal().Err(err)
}
