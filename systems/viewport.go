package systems

import (
	"context"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/ark/ecs"
)

func NewViewport() *Viewport {
	return &Viewport{}
}

type Viewport struct {
	camera       ecs.Resource[CameraHandler]
	mouse        ecs.Resource[Mouse]
	visibleWorld ecs.Resource[VisibleWorld]
	boundaries   ecs.Resource[Boundaries]
	navMode      ecs.Resource[NavigationMode]

	selection ecs.Resource[NodeSelection]
	shape     *ecs.Map2[Position, Shape]
	move      *ecs.Map2[Position, Target2]
	zoom      *ecs.Map2[Size, Target1]
	positions *ecs.Filter1[Position]
	root      *ecs.Filter1[Position]

	cameraEntity       ecs.Entity
	cameraOffsetEntity ecs.Entity
	cameraZoomEntity   ecs.Entity

	debug ecs.Resource[DebugBoard]
}

// Close implements System.
func (v *Viewport) Close() {
}

// Initialize implements System.
func (v *Viewport) Initialize(w *ecs.World) {
	v.debug = ecs.NewResource[DebugBoard](w)
	v.camera = ecs.NewResource[CameraHandler](w)
	v.mouse = ecs.NewResource[Mouse](w)
	v.visibleWorld = ecs.NewResource[VisibleWorld](w)
	cameraHandler := v.camera.Get()
	v.navMode = ecs.NewResource[NavigationMode](w)
	v.navMode.Add(&NavigationMode{Nav: FreeNav})

	v.selection = ecs.NewResource[NodeSelection](w)
	v.selection.Add(&NodeSelection{})

	v.shape = ecs.NewMap2[Position, Shape](w)
	v.positions = ecs.NewFilter1[Position](w).With(ecs.C[Node]()).With(ecs.C[VisibleElement]())
	v.root = ecs.NewFilter1[Position](w).With(ecs.C[Node]()).With(ecs.C[VisibleElement]()).Without(ecs.C[Parent]())

	v.move = ecs.NewMap2[Position, Target2](w)
	v.cameraEntity = v.move.NewEntity(&Position{
		X: float64(cameraHandler.Camera.Target.X),
		Y: float64(cameraHandler.Camera.Target.Y),
	},
		NewTarget2Empty(12))

	v.cameraOffsetEntity = v.move.NewEntity(&Position{
		X: float64(cameraHandler.Camera.Offset.X),
		Y: float64(cameraHandler.Camera.Offset.Y),
	},
		NewTarget2Empty(12))

	v.zoom = ecs.NewMap2[Size, Target1](w)
	v.cameraZoomEntity = v.zoom.NewEntity(&Size{
		Value: cameraHandler.Camera.Zoom,
	},
		NewTarget1Empty(12))
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

	selection := v.selection.Get()

	cameraHandler := v.camera.Get()
	camera := cameraHandler.Camera
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

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		delta := rl.GetMouseDelta()
		delta = rl.Vector2Scale(delta, -1.0/camera.Zoom)
		camera.Target = rl.Vector2Add(camera.Target, delta)
		target.Done = true
		targetOffset.Done = true
		targetZoom.Done = true
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && selection.HasHovered() {
		selection.Selected = selection.Hovered
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		nodeTarget := ecs.Entity{}
		if selection.HasSelected() {
			nodeTarget = selection.Selected
		} else if selection.HasHovered() {
			nodeTarget = selection.Hovered
		} else {
			rootQuery := v.root.Query()
			for rootQuery.Next() {
				nodeTarget = rootQuery.Entity()
			}
		}
		if !nodeTarget.IsZero() {
			pos, shape := v.shape.Get(nodeTarget)

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

			t, z := cameraHandler.FocusOn(points...)

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

			if target.X == target.StartX && target.Y == target.StartY {
				targetZoom.X = z
				targetZoom.SinceTick = 0
			}
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

	nodes := v.positions.Query()
	if nodes.Next() {
		p := nodes.Get()
		minPos := Position{p.X, p.Y}
		maxPos := Position{p.X + 100, p.Y + 100}
		for nodes.Next() {
			p := nodes.Get()
			if p.X < minPos.X {
				minPos.X = p.X
			}
			if p.X+100 > maxPos.X {
				maxPos.X = p.X + 100
			}
			if p.Y < minPos.Y {
				minPos.Y = p.Y
			}
			if p.Y+100 > maxPos.Y {
				maxPos.Y = p.Y + 100
			}
		}

		minWin := rl.GetWorldToScreen2D(rl.NewVector2(float32(minPos.X), float32(minPos.Y)), *camera)
		maxWin := rl.GetWorldToScreen2D(rl.NewVector2(float32(maxPos.X), float32(maxPos.Y)), *camera)

		if minWin.X > float32(rl.GetScreenWidth())/2 {
			camera.Offset.X += float32(rl.GetScreenWidth())/2 - minWin.X
		}

		if maxWin.X < float32(rl.GetScreenWidth())/2 {
			camera.Offset.X += float32(rl.GetScreenWidth())/2 - maxWin.X
		}

		if minWin.Y > float32(rl.GetScreenHeight())/2 {
			camera.Offset.Y += float32(rl.GetScreenHeight())/2 - minWin.Y
		}

		if maxWin.Y < float32(rl.GetScreenHeight())/2 {
			camera.Offset.Y += float32(rl.GetScreenHeight())/2 - maxWin.Y
		}
	}

	mousePos := rl.GetMousePosition()
	worldMousePos := rl.GetScreenToWorld2D(mousePos, *camera)

	// rl.DrawText(fmt.Sprintf("M: %d, %d", int32(worldMousePos.X), int32(worldMousePos.Y)), int32(mousePos.X)+4, int32(mousePos.Y)+4, 32, rl.Red)

	v.mouse.Get().InWorld = Position{float64(worldMousePos.X), float64(worldMousePos.Y)}
	v.mouse.Get().OnScreen = Position{float64(mousePos.X), float64(mousePos.Y)}

	topLeft := rl.GetScreenToWorld2D(rl.Vector2Zero(), *camera)
	botRight := rl.GetScreenToWorld2D(rl.NewVector2(float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())), *camera)

	visibleWorld := v.visibleWorld.Get()
	visibleWorld.X = float64(topLeft.X)
	visibleWorld.Y = float64(topLeft.Y)
	visibleWorld.MaxX = float64(botRight.X)
	visibleWorld.MaxY = float64(botRight.Y)

	// // Reduce visible part of the screen for debug
	// sizeX := botRight.X - topLeft.X
	// sizeY := botRight.Y - topLeft.Y
	// visibleWorld.X += float64(sizeX) / 4
	// visibleWorld.Y += float64(sizeY) / 4
	// visibleWorld.MaxX -= float64(sizeX) / 4
	// visibleWorld.MaxY -= float64(sizeY) / 4
	// rl.BeginMode2D(*camera)
	// rl.DrawRectangleLines(int32(visibleWorld.X), int32(visibleWorld.Y), int32(visibleWorld.MaxX-visibleWorld.X), int32(visibleWorld.MaxY-visibleWorld.Y), rl.Orange)
	// rl.EndMode2D()

	nav := v.navMode.Get()
	if selection.HasSelected() && target.Done {
		nav.Nav = KeyboardNav
	} else {
		nav.Nav = FreeNav
	}

	// Debug Grid cell for mouse
	// gpos := GridCoords(int(worldMousePos.X), int(worldMousePos.Y))
	// rl.BeginMode2D(*v.camera.Get().Camera)
	// rl.DrawRectangleLines(int32(gpos.X)*1000, int32(gpos.Y)*1000, 1000, 1000, rl.Blue)
	// rl.EndMode2D()
}
