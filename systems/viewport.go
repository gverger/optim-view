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

	hovered generic.Resource[ecs.Entity]
	shape   generic.Map2[Position, Shape]
	move    generic.Map2[Position, Target2]
	zoom    generic.Map2[Size, Target1]

	cameraEntity       ecs.Entity
	cameraOffsetEntity ecs.Entity
	cameraZoomEntity   ecs.Entity
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

	v.hovered = generic.NewResource[ecs.Entity](w)
	v.shape = generic.NewMap2[Position, Shape](w)

	v.move = generic.NewMap2[Position, Target2](w)
	v.cameraEntity = v.move.NewWith(&Position{
		X: float64(v.cameraHandler.Camera.Target.X),
		Y: float64(v.cameraHandler.Camera.Target.Y),
	},
		NewTarget2Empty(16))

	v.cameraOffsetEntity = v.move.NewWith(&Position{
		X: float64(v.cameraHandler.Camera.Offset.X),
		Y: float64(v.cameraHandler.Camera.Offset.Y),
	},
		NewTarget2Empty(16))

	v.zoom = generic.NewMap2[Size, Target1](w)
	v.cameraZoomEntity = v.zoom.NewWith(&Size{
		Value: v.cameraHandler.Camera.Zoom,
	},
		NewTarget1Empty(16))
}

// Return target, zoom
func (h *CameraHandler) FocusOn(points ...Position) (rl.Vector2, float32) {
	// move camera for the whole scene
	if len(points) > 0 {
		minX := points[0].X
		minY := points[0].Y
		maxX := points[0].X
		maxY := points[0].Y

		for _, p := range points[1:] {
			minX = min(minX, p.X)
			minY = min(minY, p.Y)
			maxX = max(maxX, p.X)
			maxY = max(maxY, p.Y)
		}

		dx := float32(maxX-minX) / float32(rl.GetScreenWidth()-20)
		dy := float32(maxY-minY) / float32(rl.GetScreenHeight()-20)

		return rl.NewVector2(float32(minX+maxX)/2, float32(minY+maxY)/2), 1 / max(dx, dy)
	}

	return h.Camera.Target, h.Camera.Zoom
}

// Update implements System.
func (v *Viewport) Update(ctx context.Context, w *ecs.World) {
	camera := v.camera.Get().Camera
	cpos, target := v.move.Get(v.cameraEntity)
	// if moved elsewhere
	if !target.Done {
		camera.Target = rl.NewVector2(float32(cpos.X), float32(cpos.Y))
	}

	offsetpos, targetOffset := v.move.Get(v.cameraOffsetEntity)
	if !targetOffset.Done {
		camera.Offset = rl.NewVector2(float32(offsetpos.X), float32(offsetpos.Y))
	}

	zoom, targetZoom := v.zoom.Get(v.cameraZoomEntity)
	if !targetZoom.Done {
		camera.Zoom = zoom.Value
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		delta := rl.GetMouseDelta()
		delta = rl.Vector2Scale(delta, -1.0/camera.Zoom)
		camera.Target = rl.Vector2Add(camera.Target, delta)
		target.Done = true
		targetOffset.Done = true
		targetZoom.Done = true
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		if v.hovered.Has() {
			hovered := v.hovered.Get()
			pos, shape := v.shape.Get(*hovered)

			points := make([]Position, 0, len(shape.Points))
			for _, p := range shape.Points {
				points = append(points, Position{pos.X + p.X, pos.Y + p.Y})
			}

			target.StartX = float64(camera.Target.X)
			target.StartY = float64(camera.Target.Y)
			cpos.X = target.StartX
			cpos.Y = target.StartY


			targetZoom.StartX = camera.Zoom
			zoom.Value = camera.Zoom

			t, z := v.cameraHandler.FocusOn(points...)

			target.X = float64(t.X)
			target.Y = float64(t.Y)
			target.SinceTick = 0

			targetOffset.StartX = float64(camera.Offset.X)
			targetOffset.StartY = float64(camera.Offset.Y)
			offsetpos.X = targetOffset.StartX
			offsetpos.Y = targetOffset.StartY

			targetOffset.X = float64(rl.GetScreenWidth()) / 2
			targetOffset.Y = float64(rl.GetScreenHeight()) / 2
			targetOffset.SinceTick = 0

			targetZoom.X = z
			targetZoom.SinceTick = 0
		}
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
		targetZoom.Done = true
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
