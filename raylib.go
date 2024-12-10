package main

import (
	"embed"
	"math"
	"sort"
	"strings"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/nikolaydubina/go-graph-layout/layout"
	"github.com/phuslu/log"
)

//go:embed data/Roboto.ttf
var f embed.FS

type MouseHandler struct {
	LastClick int
}

type CameraHandler struct {
	Camera *rl.Camera2D
}

func NewCameraHandler() CameraHandler {
	return CameraHandler{
		Camera: &rl.Camera2D{
			Zoom: 1.0,
		},
	}
}

func (h CameraHandler) Update() {
	camera := h.Camera
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		delta := rl.GetMouseDelta()
		delta = rl.Vector2Scale(delta, -1.0/camera.Zoom)
		camera.Target = rl.Vector2Add(camera.Target, delta)
	}

	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		// Get the world point that is under the mouse
		mouseWorldPos := rl.GetScreenToWorld2D(rl.GetMousePosition(), *camera)

		// Set the offset to where the mouse is
		camera.Offset = rl.GetMousePosition()

		// Set the target to match, so that the camera maps the world space point
		// under the cursor to the screen space point under the cursor at any zoom
		camera.Target = mouseWorldPos

		// Zoom increment
		scaleFactor := float32(1.0 + (0.25 * math.Abs(float64(wheel))))
		if wheel < 0 {
			scaleFactor = 1.0 / scaleFactor
		}
		camera.Zoom = rl.Clamp(camera.Zoom*scaleFactor, 0.0125, 1024.0)
	}
}

func runSingleVisu(tree *GraphView, layer layout.LayeredGraph, g layout.Graph) {
	runVisu(Input{
		Trees:   map[string]*GraphView{"Graph": tree},
		Layouts: map[string]layout.Graph{"Graph": g},
		Layers:  map[string]layout.LayeredGraph{"Graph": layer},
	})
}

func runVisu(input Input) {
	events := make(chan Event, 1)

	inputKeys := Keys(input.Trees)
	sort.Strings(inputKeys)
	keys := strings.Join(inputKeys, ";")
	log.Info().Msg(keys)

	activeTree := int32(0)
	editMode := false

	currentTree := input.Trees[inputKeys[activeTree]]
	currentLayer := input.Layers[inputKeys[activeTree]]
	currentLayout := input.Layouts[inputKeys[activeTree]]

	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	// rl.SetConfigFlags(rl.TextureFilterNearestMipLinear)

	rl.InitWindow(1600, 1000, "Graph Visualization")
	defer rl.CloseWindow()

	// rl.SetConfigFlags(rl.FlagFullscreenMode)
	// rl.ToggleFullscreen()

	fontData := Must(f.ReadFile("data/Roboto.ttf"))
	font := rl.LoadFontFromMemory(".ttf", fontData, 96, nil)

	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	rl.GenTextureMipmaps(&font.Texture)

	camera := NewCameraHandler()

	rl.SetTargetFPS(60)
	nbChildren := make(map[string]int, len(currentTree.Nodes))
	for _, n := range currentTree.Nodes {
		for _, p := range n.ParentIds {
			nbChildren[p]++
		}
	}

	var hovered *Node
	lastHovered := -1
	var selectionTexture rl.Texture2D

	for !rl.WindowShouldClose() {
		found := true
		for found {
			select {
			case event := <-events:
				log.Info().Interface("event", event).Msg("event received")
				switch e := event.(type) {
				case NewGraph:
					log.Info().Msgf("new graph %d nodes", len(e.Tree.Nodes))
					currentTree = e.Tree
					currentLayout = e.Layout
				}
			default:
				found = false
			}
		}

		camera.Update()

		gesture := rl.GetGestureDetected()

		mousePos := rl.GetMousePosition()
		worldMousePos := rl.GetScreenToWorld2D(mousePos, *camera.Camera)

		hovered = nil
		for i, n := range currentLayout.Nodes {
			if int(i) >= len(currentTree.Nodes) {
				log.Info().Int("i", int(i)).Int("tree nodes", len(currentTree.Nodes)).Interface("node", n).Msg("crash")
			}
			node := currentTree.Nodes[i]
			if node.Hidden {
				continue
			}
			if rl.CheckCollisionPointRec(worldMousePos, rl.NewRectangle(float32(n.XY[0]), float32(n.XY[1]), float32(n.W), float32(n.H))) {
				hovered = node
				if lastHovered != int(i) {
					// image := rl.LoadImageSvg(selected.SvgImage, 500, 500)
					rl.UnloadTexture(selectionTexture)
					// if img, ok := ImageFromSVG(selected.SvgImage); ok {
					// 	selectionTexture = rl.LoadTextureFromImage(img)
					// }
					// rl.UnloadImage(image)
					lastHovered = int(i)
				}
				break
			}
		}

		if hovered != nil && rl.IsMouseButtonPressed(rl.MouseLeftButton) && gesture == rl.GestureDoubletap {
			log.Info().Msg("clicked")
			go HideChildren(currentTree, currentLayout, currentLayer, hovered.Id, events)
		}

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode2D(*camera.Camera)

		for n, e := range currentLayout.Edges {
			n1 := currentTree.Nodes[n[0]]
			n2 := currentTree.Nodes[n[1]]
			if n1.Hidden || n2.Hidden {
				continue
			}
			for i := 1; i < len(e.Path); i++ {
				src := rl.NewVector2(float32(e.Path[i-1][0]), float32(e.Path[i-1][1]))
				ctrlA := rl.NewVector2(float32(e.Path[i-1][0]), float32(e.Path[i-1][1]+50))
				ctrlB := rl.NewVector2(float32(e.Path[i][0]), float32(e.Path[i][1]-50))
				dst := rl.NewVector2(float32(e.Path[i][0]), float32(e.Path[i][1]))

				rl.DrawSplineSegmentBezierCubic(src, ctrlA, ctrlB, dst, 1, rl.Blue)
			}
		}

		for i, n := range currentLayout.Nodes {
			node := currentTree.Nodes[i]
			if node.Hidden {
				continue
			}
			color := rl.Maroon
			if nbChildren[node.Id] > 0 {
				color = rl.DarkGreen
			}
			if hovered != nil && hovered.Id == node.Id {
				color = rl.DarkBlue
			}

			rl.DrawRectangle(int32(n.XY[0]), int32(n.XY[1]), int32(n.W), int32(n.H), color)
			// rl.DrawCircle(int32(n.XY[0]), int32(n.XY[1]), float32(n.H/2), rl.Maroon)
			rl.DrawTextEx(font, node.ShortInfo, rl.NewVector2(float32(n.XY[0]), float32(n.XY[1])), 11, 0, rl.Black)
		}
		rl.EndMode2D()

		if hovered != nil && !editMode {
			txtDims := rl.MeasureTextEx(rl.GetFontDefault(), hovered.Info, 32, 4)

			shape := currentLayout.Nodes[uint64(lastHovered)]
			corner := rl.GetWorldToScreen2D(rl.NewVector2(float32(shape.XY[0]+shape.W), float32(shape.XY[1])), *camera.Camera)

			distX := float32(50)
			distY := -float32(60)

			rightmostPointX := corner.X + txtDims.X + 20

			if rightmostPointX > float32(rl.GetScreenWidth()-10) {
				distX = -50 - txtDims.X - 20
				corner.X = rl.GetWorldToScreen2D(rl.NewVector2(float32(shape.XY[0]), 0), *camera.Camera).X
			} else if rightmostPointX+distX > float32(rl.GetScreenWidth()-10) {
				distX = float32(rl.GetScreenWidth()-10) - (corner.X + txtDims.X + 20)
			}

			offsetX := rl.Clamp(corner.X+distX, 10, float32(rl.GetScreenWidth())-10-txtDims.X-20)
			offsetY := rl.Clamp(corner.Y-distY, 10, float32(rl.GetScreenHeight())-10-txtDims.Y-20)

			savedBackgroundColor := gui.GetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR)
			gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, 0xDDDDDDDD)
			gui.Panel(rl.NewRectangle(offsetX, offsetY, txtDims.X+20, txtDims.Y+20), "Properties")
			rl.DrawTextEx(font, hovered.Info, rl.NewVector2(offsetX+10, offsetY+24), 32, 0, rl.Black)

			// rl.DrawTexture(selectionTexture, int32(offsetX+10), int32(offsetY+300), rl.White)

			gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, savedBackgroundColor)
		}

		if editMode {
			gui.Lock()
		}
		if gui.DropdownBox(rl.NewRectangle(10, 10, 200, 30), keys, &activeTree, editMode) {
			log.Info().Int("active", int(activeTree)).Msg("DropdownBox")
			if editMode {
				currentTree = input.Trees[inputKeys[activeTree]]
				currentLayout = input.Layouts[inputKeys[activeTree]]

				nbChildren = make(map[string]int, len(currentTree.Nodes))
				for _, n := range currentTree.Nodes {
					for _, p := range n.ParentIds {
						nbChildren[p]++
					}
				}

				hovered = nil
				lastHovered = -1
			}
			editMode = !editMode
		}
		gui.Unlock()

		rl.EndDrawing()
	}

	rl.UnloadTexture(selectionTexture)
}

type Event any

type NewGraph struct {
	Tree   *GraphView
	Layout layout.Graph
}

func HideChildren(tree *GraphView, oldG layout.Graph, oldL layout.LayeredGraph, nodeId string, events chan<- Event) {
	nodes := make([]*Node, 0, len(tree.Nodes))
	nodes = append(nodes, tree.NodeForId(nodeId))
	idx := 0
	for idx < len(nodes) {
		current := nodes[idx]
		idx++
		for c := range tree.Children(current) {
			if !c.Hidden {
				nodes = append(nodes, c)
				c.Hidden = true
				log.Debug().Str("node", c.Id).Msg("hide")
			}
		}
	}

	indices := make([]uint64, 0, len(nodes[1:]))
	for _, node := range nodes[1:] {
		idx := uint64(tree.Lookup[tree.NodeID(node)])
		log.Debug().Str("node", tree.NodeID(node)).Uint64("indice", idx).Msg("mapping")
		indices = append(indices, idx)
	}

	newL := Delete(&oldL, indices)

	g := oldG.Copy()

	newX := layout.
		BrandesKopfLayersNodesHorizontalAssigner{Delta: 150}.
		NodesHorizontalCoordinates(g, *newL)

	newY := layout.BasicNodesVerticalCoordinatesAssigner{
		MarginLayers:   125,
		FakeNodeHeight: 50,
	}.NodesVerticalCoordinates(g, *newL)

	nodeIdx := uint64(tree.Lookup[nodeId])
	offset := [2]int{
		oldG.Nodes[nodeIdx].XY[0] - (newX[nodeIdx] - g.Nodes[nodeIdx].W/2),
		oldG.Nodes[nodeIdx].XY[1] - (newY[nodeIdx] - g.Nodes[nodeIdx].H/2),
	}

	allNodesXY := make(map[uint64][2]int, len(g.Nodes))
	for n := range newL.NodeYX {
		allNodesXY[n] = [2]int{newX[n] + offset[0], newY[n] + offset[1]}
	}

	for e, nodes := range newL.Edges {
		path := make([][2]int, len(nodes))
		for i, n := range nodes {
			xy := allNodesXY[n]
			path[i] = [2]int{xy[0], xy[1]}
		}

		g.Edges[e] = layout.Edge{Path: path}
	}

	for k := range g.Nodes {
		g.Nodes[k] = layout.Node{
			XY: [2]int{newX[k] - g.Nodes[k].W/2 + offset[0], newY[k] - g.Nodes[k].H/2 + offset[1]},
			W:  g.Nodes[k].W,
			H:  g.Nodes[k].H,
		}
	}

	events <- NewGraph{Tree: tree, Layout: g}
}
