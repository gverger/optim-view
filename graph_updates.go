package main

import (
	"github.com/nikolaydubina/go-graph-layout/layout"
	"github.com/phuslu/log"
)

func Delete(lg *layout.LayeredGraph, nodes []uint64) *layout.LayeredGraph {
	g := &layout.LayeredGraph{
		Segments: make(map[[2]uint64]bool),
		Dummy:    make(map[uint64]bool),
		NodeYX:   make(map[uint64][2]int),
		Edges:    make(map[[2]uint64][]uint64),
	}

	toDelete := make(map[uint64]bool)
	for _, n := range nodes {
		toDelete[n] = true
	}

	for k, nodes := range lg.Edges {
		log.Debug().Interface("k", k).Msg("checking")
		n1 := k[0]
		n2 := k[1]
		if toDelete[n1] || toDelete[n2] {
			for _, v := range nodes[1 : len(nodes)-1] {
				toDelete[v] = true
			}
		} else {
			g.Edges[k] = make([]uint64, len(nodes))
			copy(g.Edges[k], nodes)
			log.Debug().Interface("edges", g.Edges[k]).Interface("k", k).Msg("add edge")
		}
	}

	log.Debug().Interface("delete", toDelete).Msg("delete")

	for s := range lg.Segments {
		if !toDelete[s[0]] && !toDelete[s[1]] {
			g.Segments[s] = true
			log.Debug().Interface("segment", s).Msg("segment added")
		}
	}

	for d := range lg.Dummy {
		if !toDelete[d] {
			g.Dummy[d] = true
			log.Debug().Interface("dummy", d).Msg("dummy added")
		}
	}

	for k, v := range lg.NodeYX {
		if !toDelete[k] {
			g.NodeYX[k] = v
			log.Debug().Uint64("k", k).Interface("NodeYX", v).Msg("node added")
		}
	}

	layers := g.Layers()
	for layerIdx, l := range layers {
		for orderIdx, n := range l {
			g.NodeYX[n] = [2]int{layerIdx, orderIdx}
		}
	}

	return g
}
