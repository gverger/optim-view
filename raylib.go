package main

import (
	"math"

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
	rl.SetConfigFlags(rl.FlagWindowMaximized)
	rl.SetConfigFlags(rl.FlagMsaa4xHint)

	rl.InitWindow(1200, 800, "Graph Visualization")
	defer rl.CloseWindow()

	camera := NewCameraHandler()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		camera.Update()

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode2D(*camera.Camera)

		for _, e := range g.Edges {
			for i := 0; i < len(e.Path)-1; i++ {
				rl.DrawLine(
					int32(e.Path[i][0]), int32(e.Path[i][1]),
					int32(e.Path[i+1][0]), int32(e.Path[i+1][1]),
					rl.Blue)
			}
		}

		for i, n := range g.Nodes {
			rl.DrawRectangle(int32(n.XY[0]), int32(n.XY[1]), int32(n.W), int32(n.H), rl.Maroon)
			// rl.DrawCircle(int32(n.XY[0]), int32(n.XY[1]), float32(n.H/2), rl.Maroon)
			rl.DrawText(input.Nodes[i].ShortInfo, int32(n.XY[0]), int32(n.XY[1]), 16, rl.Black)
		}
		rl.EndMode2D()
		rl.EndDrawing()
	}
}
