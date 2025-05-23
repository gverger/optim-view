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

func (s *GeometryCache) recUpdateCache(ctx context.Context, w *ecs.World, p *Position, n *Node, e ecs.Entity, boundingBoxes map[ecs.Entity]*SubTreeBoundingBox) {
	if !boundingBoxes[e].dirty {
		return
	}
	if ctx.Err() != nil {
		return
	}
	parentBB := boundingBoxes[e]
	parentBB.X = p.X
	parentBB.Y = p.Y
	parentBB.Width = n.SizeX
	parentBB.Height = n.SizeY

	qChildren := s.filterChildren.Query(ecs.Rel[ChildOf](e))
	children := make([]child, 0, 16)
	for qChildren.Next() {
		c := qChildren.Entity()
		cP, cN, _, _ := qChildren.Get()
		children = append(children, child{
			e: c,
			p: cP,
			n: cN,
		})
	}

	for _, c := range children {
		s.recUpdateCache(ctx, w, c.p, c.n, c.e, boundingBoxes)
		childBB := boundingBoxes[c.e]
		if childBB.dirty {
			return
		}

		x2 := max(parentBB.X+parentBB.Width, childBB.X+childBB.Width)
		y2 := max(parentBB.Y+parentBB.Height, childBB.Y+childBB.Height)

		parentBB.X = min(parentBB.X, childBB.X)
		parentBB.Y = min(parentBB.Y, childBB.Y)
		parentBB.Width = x2 - parentBB.X
		parentBB.Height = y2 - parentBB.Y
	}
	parentBB.dirty = false
}

var countCache = 0

func (s *GeometryCache) Update(ctx context.Context, w *ecs.World) {
	rootQuery := s.filterRoot.Query()
	countCache = 0
	for rootQuery.Next() {
		p, n := rootQuery.Get()
		s.recUpdateCache(ctx, w, p, n, rootQuery.Entity(), s.boundingBoxes.Get().boundingBoxes)
	}
}

var _ System = &GeometryCache{}
