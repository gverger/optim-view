package main

import (
	"embed"
	"math"
	"sort"
	"strings"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graph"
	"github.com/gverger/optimview/systems"
	"github.com/mlange-42/arche/ecs"
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

func (h *CameraHandler) Update() {
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
	// currentLayer := input.Layers[inputKeys[activeTree]]
	// currentLayout := input.Layouts[inputKeys[activeTree]]

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

	sys := systems.New()
	sys.Add(systems.NewInitializer(*currentTree))
	sys.Add(systems.NewMouseSelector())
	// sys.Add(systems.NewTargeter())
	sys.Add(systems.NewMover())
	sys.Add(systems.NewDrawEdges(font))
	sys.Add(systems.NewDrawNodes(font))
	w := ecs.NewWorld()
	sys.Initialize(&w)

	camera := NewCameraHandler()

	positions := graph.ComputeLayeredCoordinates(*currentTree)
	offset := graph.Position{X: positions[0].X - rl.GetScreenWidth()/2, Y: positions[0].Y - rl.GetScreenHeight()/4}
	for _, node := range currentTree.Nodes {
		sys.MoveNode(&w, node.Id, positions[node.Id].X-offset.X, positions[node.Id].Y-offset.Y)
	}

	rl.SetTargetFPS(60)

	var hovered *DisplayableNode
	// lastHovered := -1
	var selectionTexture rl.Texture2D

	for !rl.WindowShouldClose() {
		found := true
		for found {
			select {
			case event := <-events:
				log.Info().Interface("event", event).Msg("event received")
			default:
				found = false
			}
		}

		camera.Update()

		gesture := rl.GetGestureDetected()

		mousePos := rl.GetMousePosition()
		worldMousePos := rl.GetScreenToWorld2D(mousePos, *camera.Camera)
		sys.SetMouse(float64(mousePos.X), float64(mousePos.Y), float64(worldMousePos.X), float64(worldMousePos.Y))

		// hovered = nil
		// for i, n := range currentTree.Nodes {
		// 	if rl.CheckCollisionPointRec(worldMousePos, rl.NewRectangle(float32(n.XY[0]), float32(n.XY[1]), float32(n.W), float32(n.H))) {
		// 		hovered = currentTree.Nodes[i]
		// 		if lastHovered != int(i) {
		// 			// image := rl.LoadImageSvg(selected.SvgImage, 500, 500)
		// 			rl.UnloadTexture(selectionTexture)
		// 			// if img, ok := ImageFromSVG(selected.SvgImage); ok {
		// 			// 	selectionTexture = rl.LoadTextureFromImage(img)
		// 			// }
		// 			// rl.UnloadImage(image)
		// 			lastHovered = int(i)
		// 		}
		// 		break
		// 	}
		// }

		if hovered != nil && rl.IsMouseButtonPressed(rl.MouseLeftButton) && gesture == rl.GestureDoubletap {
			log.Info().Msg("clicked")
		}

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode2D(*camera.Camera)

		sys.Update(&w)


		rl.EndMode2D()

		// if hovered != nil && !editMode {
		// 	txtDims := rl.MeasureTextEx(rl.GetFontDefault(), hovered.Text, 32, 4)
		//
		// 	shape := currentLayout.Nodes[uint64(lastHovered)]
		// 	corner := rl.GetWorldToScreen2D(rl.NewVector2(float32(shape.XY[0]+shape.W), float32(shape.XY[1])), *camera.Camera)
		//
		// 	distX := float32(50)
		// 	distY := -float32(60)
		//
		// 	rightmostPointX := corner.X + txtDims.X + 20
		//
		// 	if rightmostPointX > float32(rl.GetScreenWidth()-10) {
		// 		distX = -50 - txtDims.X - 20
		// 		corner.X = rl.GetWorldToScreen2D(rl.NewVector2(float32(shape.XY[0]), 0), *camera.Camera).X
		// 	} else if rightmostPointX+distX > float32(rl.GetScreenWidth()-10) {
		// 		distX = float32(rl.GetScreenWidth()-10) - (corner.X + txtDims.X + 20)
		// 	}
		//
		// 	offsetX := rl.Clamp(corner.X+distX, 10, float32(rl.GetScreenWidth())-10-txtDims.X-20)
		// 	offsetY := rl.Clamp(corner.Y-distY, 10, float32(rl.GetScreenHeight())-10-txtDims.Y-20)
		//
		// 	savedBackgroundColor := gui.GetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR)
		// 	gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, 0xDDDDDDDD)
		// 	gui.Panel(rl.NewRectangle(offsetX, offsetY, txtDims.X+20, txtDims.Y+20), "Properties")
		// 	rl.DrawTextEx(font, hovered.Text, rl.NewVector2(offsetX+10, offsetY+24), 32, 0, rl.Black)
		//
		// 	// rl.DrawTexture(selectionTexture, int32(offsetX+10), int32(offsetY+300), rl.White)
		//
		// 	gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, savedBackgroundColor)
		// }
		//
		if editMode {
			gui.Lock()
		}
		if gui.DropdownBox(rl.NewRectangle(10, 10, 200, 30), keys, &activeTree, editMode) {
			log.Info().Int("active", int(activeTree)).Msg("DropdownBox")
			if editMode {
				currentTree = input.Trees[inputKeys[activeTree]]
				// currentLayout = input.Layouts[inputKeys[activeTree]]

				hovered = nil
				// lastHovered = -1
			}
			editMode = !editMode
		}
		gui.Unlock()

		rl.EndDrawing()
	}

	rl.UnloadTexture(selectionTexture)
}

type Event any
