package systems

import (
	"context"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/ark/ecs"
)

func NewInputs() *Inputs {
	return &Inputs{}
}

type Inputs struct {
	input  ecs.Resource[Input]
	camera ecs.Resource[CameraHandler]
}

// Close implements System.
func (s *Inputs) Close() {
}

// Initialize implements System.
func (s *Inputs) Initialize(w *ecs.World) {
	s.input = ecs.NewResource[Input](w)
	s.camera = ecs.NewResource[CameraHandler](w)
}

// Update implements System.
func (s *Inputs) Update(ctx context.Context, w *ecs.World) {

	input := s.input.Get()

	if !input.Active {
		input.Mouse = Mouse{}
		input.KeyPressed = Keyboard{}
		return
	}

	cameraHandler := s.camera.Get()
	camera := cameraHandler.Camera

	mousePos := rl.GetMousePosition()
	worldMousePos := rl.GetScreenToWorld2D(mousePos, *camera)

	input.Mouse = Mouse{
		OnScreen: Position{float64(mousePos.X), float64(mousePos.Y)},
		InWorld:  Position{float64(worldMousePos.X), float64(worldMousePos.Y)},
		Delta:    Position{float64(rl.GetMouseDelta().X), float64(rl.GetMouseDelta().Y)},

		LeftButton: MouseButton{
			Pressed: rl.IsMouseButtonPressed(rl.MouseButtonLeft),
			Down:    rl.IsMouseButtonDown(rl.MouseButtonLeft),
		},

		RightButton: MouseButton{
			Pressed: rl.IsMouseButtonPressed(rl.MouseButtonRight),
			Down:    rl.IsMouseButtonDown(rl.MouseButtonRight),
		},

		HorizontalScroll: rl.GetMouseWheelMoveV().X,
		VerticalScroll:   rl.GetMouseWheelMoveV().Y,
	}

	input.KeyPressed = Keyboard{
		Down:  rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyJ),
		Up:    rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyK),
		Right: rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressed(rl.KeyL),
		Left:  rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressed(rl.KeyH),

		Space: rl.IsKeyPressed(rl.KeySpace),
	}
}

var _ System = &Inputs{}
