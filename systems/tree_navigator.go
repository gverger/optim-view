package systems

import (
	"context"
	"math"

	"github.com/mlange-42/ark/ecs"
)

func NewTreeNavigator() *TreeNavigator {
	return &TreeNavigator{}
}

type TreeNavigator struct {
	mode     ecs.Resource[NavigationMode]
	selected ecs.Resource[NodeSelection]
	edges    *ecs.Filter2[Edge, VisibleElement]
	children *ecs.Filter1[ChildOf]
	parent   *ecs.Map1[Parent]
	nodes    *ecs.Map1[Position]
	visible  *ecs.Map1[VisibleElement]

	input ecs.Resource[Input]
	debug ecs.Resource[DebugBoard]
}

// Initialize implements System.
func (t *TreeNavigator) Initialize(w *ecs.World) {
	t.mode = ecs.NewResource[NavigationMode](w)
	t.selected = ecs.NewResource[NodeSelection](w)
	t.edges = ecs.NewFilter2[Edge, VisibleElement](w)
	t.children = ecs.NewFilter1[ChildOf](w).With(ecs.C[VisibleElement]())
	t.parent = ecs.NewMap1[Parent](w)
	t.nodes = ecs.NewMap1[Position](w)
	t.visible = ecs.NewMap1[VisibleElement](w)

	t.input = ecs.NewResource[Input](w)
	t.debug = ecs.NewResource[DebugBoard](w)
}

// Update implements System.
func (t *TreeNavigator) Update(ctx context.Context, w *ecs.World) {
	if t.mode.Get().Nav != KeyboardNav {
		return
	}

	selection := t.selected.Get()
	if selection == nil || !selection.HasSelected() {
		return
	}

	input := t.input.Get()

	// debug := t.debug.Get()
	// debug.Write(fmt.Sprintf("SELECTED: %v", selection.Selected))
	children := make([]ecs.Entity, 0)

	childrenQuery := t.children.Query(ecs.Rel[ChildOf](selection.Selected))
	for childrenQuery.Next() {
		children = append(children, childrenQuery.Entity())
	}

	var parentEntity = t.parent.Get(selection.Selected)
	var parent ecs.Entity
	if parentEntity != nil {
		parent = parentEntity.parent
	}
	// debug.Writef("children: %d", len(children))

	siblings := make([]ecs.Entity, 0)
	siblingsQuery := t.children.Query(ecs.Rel[ChildOf](parent))
	for siblingsQuery.Next() {
		siblings = append(siblings, siblingsQuery.Entity())
	}

	bestNode := ecs.Entity{}

	if input.KeyPressed.Down {
		minX := math.MaxFloat64
		for _, c := range children {
			if t.visible.Get(c) == nil {
				continue
			}

			node := t.nodes.Get(c)
			if node.X < minX {
				minX = node.X
				bestNode = c
			}
		}
	}

	if input.KeyPressed.Up {
		bestNode = parent
	}

	if input.KeyPressed.Left {
		me := t.nodes.Get(selection.Selected)
		maxX := -math.MaxFloat64
		for _, s := range siblings {
			node := t.nodes.Get(s)
			if node.X > maxX && node.X < me.X {
				maxX = node.X
				bestNode = s
			}
		}
	}

	if input.KeyPressed.Right {
		me := t.nodes.Get(selection.Selected)
		minX := math.MaxFloat64
		for _, s := range siblings {
			node := t.nodes.Get(s)
			if node.X < minX && node.X > me.X {
				minX = node.X
				bestNode = s
			}
		}
	}

	if !bestNode.IsZero() {
		selection.Selected = bestNode
	}

}

// Close implements System.
func (t *TreeNavigator) Close() {
}

var _ System = &TreeNavigator{}
