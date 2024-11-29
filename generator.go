package main

import (
	"fmt"
	"math/rand/v2"
	"strconv"
)

func GenerateInput(nbNodes int) Input {
	nodes := make([]Node, nbNodes)
	for i := 0; i < nbNodes; i++ {
		parentId := 0
		if i > 0 {
			parentId = rand.IntN(i)
		}
		n := Node{
			Id:        strconv.Itoa(i),
			ParentId:  strconv.Itoa(parentId),
			Info:      fmt.Sprintf("Node %d\nIts parent is node %d", i, parentId),
			ShortInfo: fmt.Sprintf("Node %d", i),
		}

		nodes[i] = n
	}

	return Input{
		Nodes: nodes,
	}
}
