package systems

import (
	"context"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewTreeNavigator() *TreeNavigator {
	return &TreeNavigator{}
}

type TreeNavigator struct {
	mode     generic.Resource[NavigationMode]
	selected generic.Resource[NodeSelection]
	edges    generic.Filter2[Edge, VisibleElement]
	nodes    generic.Map1[Position]
	visible  generic.Map1[VisibleElement]
}

// Initialize implements System.
func (t *TreeNavigator) Initialize(w *ecs.World) {
	t.mode = generic.NewResource[NavigationMode](w)
	t.selected = generic.NewResource[NodeSelection](w)
	t.edges = *generic.NewFilter2[Edge, VisibleElement]()
	t.nodes = generic.NewMap1[Position](w)
	t.visible = generic.NewMap1[VisibleElement](w)
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

	// rl.DrawText("SELECTED", 20, 300, 32, rl.Blue)

	edgeQuery := t.edges.Query(w)
	var parent ecs.Entity
	children := make([]ecs.Entity, 0)
	for edgeQuery.Next() {
		e, _ := edgeQuery.Get()

		if e.From.ID() == selection.Selected.ID() {
			children = append(children, e.To)
		} else if e.To.ID() == selection.Selected.ID() {
			parent = e.From
		}
	}

	siblings := make([]ecs.Entity, 0)
	if !parent.IsZero() {
		edgeQuery = t.edges.Query(w)
		for edgeQuery.Next() {
			e, _ := edgeQuery.Get()
			if e.From.ID() == parent.ID() {
				siblings = append(siblings, e.To)
			}
		}
	}

	bestNode := ecs.Entity{}

	if isPressed(rl.KeyJ) || isPressed(rl.KeyDown) {
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

	if isPressed(rl.KeyK) || isPressed(rl.KeyUp) {
		bestNode = parent
	}

	if isPressed(rl.KeyH) || isPressed(rl.KeyLeft) {
		me := t.nodes.Get(selection.Selected)
		maxX := -math.MaxFloat64
		for _, s := range siblings {
			if t.visible.Get(s) == nil {
				continue
			}

			node := t.nodes.Get(s)
			if node.X > maxX && node.X < me.X {
				maxX = node.X
				bestNode = s
			}
		}
	}

	if isPressed(rl.KeyL) || isPressed(rl.KeyRight) {
		me := t.nodes.Get(selection.Selected)
		minX := math.MaxFloat64
		for _, s := range siblings {
			if t.visible.Get(s) == nil {
				continue
			}

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

func isPressed(key int32) bool {
	return rl.IsKeyPressed(key) || (rl.IsKeyDown(rl.KeyLeftShift) && rl.IsKeyDown(key))
}

// Close implements System.
func (t *TreeNavigator) Close() {
}

var _ System = &TreeNavigator{}
