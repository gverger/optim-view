package systems

import (
	"context"

	"github.com/mlange-42/ark/ecs"
)

func NewGeometryCache() *GeometryCache {
	return &GeometryCache{}
}

type GeometryCache struct {
	boundingBoxes ecs.Resource[SubTreeBoundingBoxes]

	filterNodes    *ecs.Filter2[Position, Node]
	filterChildren *ecs.Filter4[Position, Node, Parent, ChildOf]
	filterRoot     *ecs.Filter2[Position, Node]

	debug ecs.Resource[DebugBoard]
}

// Close implements System.
func (s *GeometryCache) Close() {
}

func (s *GeometryCache) Initialize(w *ecs.World) {
	s.boundingBoxes = ecs.NewResource[SubTreeBoundingBoxes](w)

	s.boundingBoxes.Add(&SubTreeBoundingBoxes{
		boundingBoxes: make(map[ecs.Entity]*SubTreeBoundingBox, 0),
	})

	s.filterNodes = ecs.NewFilter2[Position, Node](w).With(ecs.C[VisibleElement]())
	s.filterChildren = ecs.NewFilter4[Position, Node, Parent, ChildOf](w).With(ecs.C[VisibleElement]())
	s.filterRoot = ecs.NewFilter2[Position, Node](w).With(ecs.C[VisibleElement]()).Without(ecs.C[Parent]())

	bb := s.boundingBoxes.Get()

	rootQuery := s.filterRoot.Query()
	for rootQuery.Next() {
		p, n := rootQuery.Get()

		bb.boundingBoxes[rootQuery.Entity()] = &SubTreeBoundingBox{
			parentNode: ecs.Entity{},
			rootNode:   rootQuery.Entity(),
			X:          p.X,
			Y:          p.Y,
			Width:      n.SizeX,
			Height:     n.SizeY,
			dirty:      true,
		}
	}

	q := s.filterChildren.Query()
	for q.Next() {
		p, n, parent, _ := q.Get()

		bb.boundingBoxes[q.Entity()] = &SubTreeBoundingBox{
			parentNode: parent.parent,
			rootNode:   q.Entity(),
			X:          p.X,
			Y:          p.Y,
			Width:      n.SizeX,
			Height:     n.SizeY,
			dirty:      true,
		}
	}
	s.debug = ecs.NewResource[DebugBoard](w)
}

type child struct {
	e      ecs.Entity
	p      *Position
	n      *Node
	parent *Parent
}

func (s *GeometryCache) updateCache(ctx context.Context, w *ecs.World, p *Position, n *Node, e ecs.Entity) {

	boundingBoxes := s.boundingBoxes.Get().boundingBoxes

	nodesInOrder := make([]child, 0)

	children := make([]child, 0, 100)

	children = append(children, child{e: e, p: p, n: n})
	for len(children) > 0 {
		countCache++
		if ctx.Err() != nil {
			return
		}

		c := children[len(children)-1]
		children = children[:len(children)-1]
		nodesInOrder = append(nodesInOrder, c)
		e := c.e
		p := c.p
		n := c.n

		bb := boundingBoxes[e]

		if !bb.dirty {
			continue
		}
		bb.dirty = false
		bb.X = p.X
		bb.Y = p.Y
		bb.Width = n.SizeX
		bb.Height = n.SizeY

		qChildren := s.filterChildren.Query(ecs.Rel[ChildOf](e))
		for qChildren.Next() {
			c := qChildren.Entity()

			p2, to, pa, _ := qChildren.Get()

			children = append(children, child{e: c, p: p2, n: to, parent: pa})
		}
	}

	for j := range nodesInOrder[1:] {
		i := len(nodesInOrder) - j - 1
		node := nodesInOrder[i]

		parent := node.parent.parent
		c := node.e

		newX2 := max(boundingBoxes[parent].X+boundingBoxes[parent].Width, boundingBoxes[c].X+boundingBoxes[c].Width)
		newY2 := max(boundingBoxes[parent].Y+boundingBoxes[parent].Height, boundingBoxes[c].Y+boundingBoxes[c].Height)

		boundingBoxes[parent].X = min(boundingBoxes[parent].X, boundingBoxes[c].X)
		boundingBoxes[parent].Y = min(boundingBoxes[parent].Y, boundingBoxes[c].Y)
		boundingBoxes[parent].Width = newX2 - boundingBoxes[parent].X
		boundingBoxes[parent].Height = newY2 - boundingBoxes[parent].Y
	}

}

var countCache = 0

func (s *GeometryCache) Update(ctx context.Context, w *ecs.World) {
	rootQuery := s.filterRoot.Query()
	countCache = 0
	for rootQuery.Next() {
		p, n := rootQuery.Get()
		s.updateCache(ctx, w, p, n, rootQuery.Entity())
	}
}

var _ System = &GeometryCache{}
