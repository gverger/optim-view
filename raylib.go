package main

import (
	"math"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/nikolaydubina/go-graph-layout/layout"
)

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

func runVisu(input Input, g layout.Graph) {
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	// rl.SetConfigFlags(rl.TextureFilterNearestMipLinear)

	rl.InitWindow(1600, 1000, "Graph Visualization")
	defer rl.CloseWindow()

	camera := NewCameraHandler()

	rl.SetTargetFPS(60)
	nbChildren := make(map[string]int, len(input.Nodes))
	for _, n := range input.Nodes {
		for _, p := range n.ParentIds {
			nbChildren[p]++
		}
	}

	var selected *Node
	lastSelected := -1
	var selectionTexture rl.Texture2D

	for !rl.WindowShouldClose() {
		camera.Update()

		mousePos := rl.GetMousePosition()
		worldMousePos := rl.GetScreenToWorld2D(mousePos, *camera.Camera)
		selected = nil
		for i, n := range g.Nodes {
			if rl.CheckCollisionPointRec(worldMousePos, rl.NewRectangle(float32(n.XY[0]), float32(n.XY[1]), float32(n.W), float32(n.H))) {
				selected = &input.Nodes[i]
				if lastSelected != int(i) {
					// image := rl.LoadImageSvg(selected.SvgImage, 500, 500)
					rl.UnloadTexture(selectionTexture)
					if img, ok := ImageFromSVG(selected.SvgImage); ok {
						selectionTexture = rl.LoadTextureFromImage(img)
					}
					// rl.UnloadImage(image)
					lastSelected = int(i)
				}
				break
			}
		}

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode2D(*camera.Camera)

		for _, e := range g.Edges {
			for i := 0; i < len(e.Path)-1; i++ {
				rl.DrawLine(
					int32(e.Path[i][0]), int32(e.Path[i][1]),
					int32(e.Path[i+1][0]), int32(e.Path[i+1][1]),
					rl.DarkBlue)
			}
		}

		for i, n := range g.Nodes {
			color := rl.Maroon
			if nbChildren[input.Nodes[i].Id] > 0 {
				color = rl.DarkGreen
			}
			if selected != nil && selected.Id == input.Nodes[i].Id {
				color = rl.DarkBlue
			}

			rl.DrawRectangle(int32(n.XY[0]), int32(n.XY[1]), int32(n.W), int32(n.H), color)
			// rl.DrawCircle(int32(n.XY[0]), int32(n.XY[1]), float32(n.H/2), rl.Maroon)
			rl.DrawText(input.Nodes[i].ShortInfo, int32(n.XY[0]), int32(n.XY[1]), 16, rl.Black)
		}
		rl.EndMode2D()

		if selected != nil {
			txtDims := rl.MeasureTextEx(rl.GetFontDefault(), selected.Info, 32, 4)

			shape := g.Nodes[uint64(lastSelected)]
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
			rl.DrawText(selected.Info, int32(offsetX+10), int32(offsetY+24), 32, rl.Black)

			rl.DrawTexture(selectionTexture, int32(offsetX+10), int32(offsetY+300), rl.White)

			gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, savedBackgroundColor)
		}

		rl.EndDrawing()
	}

	rl.UnloadTexture(selectionTexture)
}
