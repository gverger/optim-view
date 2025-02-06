package systems

import (
	"context"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewViewport(cameraHandler CameraHandler) *Viewport {
	return &Viewport{
		cameraHandler: &cameraHandler,
	}
}

type Viewport struct {
	cameraHandler *CameraHandler

	camera       generic.Resource[CameraHandler]
	mouse        generic.Resource[Mouse]
	visibleWorld generic.Resource[VisibleWorld]
}

// Close implements System.
func (v *Viewport) Close() {
}

// Initialize implements System.
func (v *Viewport) Initialize(w *ecs.World) {
	v.camera = generic.NewResource[CameraHandler](w)
	v.mouse = generic.NewResource[Mouse](w)
	v.visibleWorld = generic.NewResource[VisibleWorld](w)
	v.camera.Add(v.cameraHandler)
}

// Update implements System.
func (v *Viewport) Update(ctx context.Context, w *ecs.World) {
	camera := v.camera.Get().Camera
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

	mousePos := rl.GetMousePosition()
	worldMousePos := rl.GetScreenToWorld2D(mousePos, *camera)

	v.mouse.Get().InWorld = Position{float64(worldMousePos.X), float64(worldMousePos.Y)}
	v.mouse.Get().OnScreen = Position{float64(mousePos.X), float64(mousePos.Y)}

	topLeft := rl.GetScreenToWorld2D(rl.Vector2Zero(), *camera)
	botRight := rl.GetScreenToWorld2D(rl.NewVector2(float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())), *camera)

	visibleWorld := v.visibleWorld.Get()
	visibleWorld.X = float64(topLeft.X)
	visibleWorld.Y = float64(topLeft.Y)
	visibleWorld.MaxX = float64(botRight.X)
	visibleWorld.MaxY = float64(botRight.Y)
}
