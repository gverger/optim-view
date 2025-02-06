package systems

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
)

type Mappings struct {
	nodeLookup map[uint64]ecs.Entity
	edgeLookup map[[2]uint64]ecs.Entity
}

type Mouse struct {
	OnScreen Position
	InWorld  Position
}

type VisibleWorld struct {
	X    float64
	Y    float64
	MaxX float64
	MaxY float64
}

type Shapes struct {
	Polygons [][]Position
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

type SelectedNode struct {
	Entity ecs.Entity
}

func (s SelectedNode) IsSet() bool {
	return !s.Entity.IsZero()
}

type NavType uint

const (
	FreeNav     NavType = 0
	KeyboardNav NavType = 1
)

type NavigationMode struct {
	Nav NavType
}
