package main

import (
	"fmt"
	"math/rand/v2"
	"strconv"
)

func GenerateLargeInput(nbNodes int) InputTree {
	nodes := make([]Node, nbNodes)
	for i := 0; i < nbNodes; i++ {
		parentIds := make([]string, 0)
		if i > 0 {
			parentIds = append(parentIds, strconv.Itoa(rand.IntN(i)))
		}
		n := Node{
			Id:        strconv.Itoa(i),
			ParentIds: parentIds,
			Info:      fmt.Sprintf("Node %d\nIts parent is node %d", i, parentIds),
			ShortInfo: fmt.Sprintf("Node %d", i),
		}

		nodes[i] = n
	}

	return InputTree{
		Nodes: nodes,
	}
}

func GenerateDeepInput(nbNodes int) InputTree {
	nodes := make([]Node, nbNodes)
	parentId := 0
	rate := float32(0.7)
	for i := 0; i < nbNodes; i++ {
		n := Node{
			Id:        strconv.Itoa(i),
			Info:      fmt.Sprintf("Node %d\nIts parent is node %d", i, parentId),
			ShortInfo: fmt.Sprintf("Node %d", i),
		}
		if parentId != i {
			n.ParentIds = []string{strconv.Itoa(parentId)}
		}

		for parentId != i && rand.Float32() > rate {
			parentId++
			rate -= 1 / float32(nbNodes)
		}
		// if parentId > 0 && parentId != i && rand.Float32() > 0.9 {
		// 	n.ParentIds = append(n.ParentIds, nodes[rand.IntN(parentId)].Id)
		// 	fmt.Println("2 parents: ", n.Id)
		// }
		nodes[i] = n

	}

	return InputTree{
		Nodes: nodes,
	}
}
