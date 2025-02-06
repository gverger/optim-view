package systems

import (
	"context"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewNodeDetails(font rl.Font) *NodeDetails {
	return &NodeDetails{font: font}
}

type NodeDetails struct {
	font rl.Font

	nodes  generic.Filter2[Position, Node]
	mouse  generic.Resource[Mouse]
	camera generic.Resource[CameraHandler]
}

// Close implements System.
func (n *NodeDetails) Close() {
}

// Initialize implements System.
func (n *NodeDetails) Initialize(w *ecs.World) {
	n.nodes = *generic.NewFilter2[Position, Node]()
	n.mouse = generic.NewResource[Mouse](w)
	n.camera = generic.NewResource[CameraHandler](w)
}

// Update implements System.
func (n *NodeDetails) Update(ctx context.Context, w *ecs.World) {

	mouse := n.mouse.Get()
	query := n.nodes.Query(w)

	for query.Next() {
		pos, node := query.Get()

		if pos.X <= mouse.InWorld.X && pos.Y <= mouse.InWorld.Y && mouse.InWorld.X <= pos.X+node.SizeX && mouse.InWorld.Y <= pos.Y+node.SizeY {

			n.displayDetails(node, pos)
			query.Close()
			break
		}
	}

}

func (n *NodeDetails) displayDetails(hoveredNode *Node, pos *Position) {
	txtDims := rl.MeasureTextEx(n.font, hoveredNode.Text, 32, 0)

	mousePosition := rl.GetMousePosition()

	distX := float32(50)

	rightmostPointX := mousePosition.X + txtDims.X + 20

	if rightmostPointX > float32(rl.GetScreenWidth()-10) {
		distX = -50 - txtDims.X - 20
	} else if rightmostPointX+distX > float32(rl.GetScreenWidth()-10) {
		distX = float32(rl.GetScreenWidth()-10) - (mousePosition.X + txtDims.X + 20)
	}

	offsetX := rl.Clamp(mousePosition.X+distX, 10, float32(rl.GetScreenWidth())-10-txtDims.X-20)
	offsetY := rl.Clamp(mousePosition.Y, 10, float32(rl.GetScreenHeight())-10-txtDims.Y-20)

	savedBackgroundColor := gui.GetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR)
	gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, 0xDDDDDDDD)
	gui.Panel(rl.NewRectangle(offsetX, offsetY, txtDims.X+20, txtDims.Y+20), hoveredNode.Title)
	rl.DrawTextEx(n.font, hoveredNode.Text, rl.NewVector2(offsetX+10, offsetY+24), 32, 0, rl.Black)

	gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, savedBackgroundColor)
}

var _ System = &NodeDetails{}
