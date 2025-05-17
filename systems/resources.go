package systems

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/ark/ecs"
)

type Mappings struct {
	nodeLookup map[uint64]ecs.Entity
	edgeLookup map[[2]uint64]ecs.Entity
}

type Mouse struct {
	OnScreen Position
	InWorld  Position
}

// Boundaries of the graph
type Boundaries struct {
	X    float64
	Y    float64
	MaxX float64
	MaxY float64
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

type NodeSelection struct {
	Hovered  ecs.Entity
	Selected ecs.Entity
}

func (s NodeSelection) HasHovered() bool {
	return !s.Hovered.IsZero()
}

func (s NodeSelection) HasSelected() bool {
	return !s.Selected.IsZero()
}

type NavType uint

const (
	FreeNav     NavType = 0
	KeyboardNav NavType = 1
)

type NavigationMode struct {
	Nav NavType
}

type DebugText struct {
	Text string
}

type GridPos struct {
	X int
	Y int
}

func GridCoords(x, y int) GridPos {
	gx := x / 1000
	if x < 0 {
		gx--
	}
	gy := y / 1000
	if y < 0 {
		gy--
	}
	return GridPos{
		X: gx,
		Y: gy,
	}
}

type Grid struct {
	grid map[GridPos][]ecs.Entity
}

func (g *Grid) AddEntity(e ecs.Entity, pos GridPos) {
	g.grid[pos] = append(g.grid[pos], e)
}

func (g *Grid) MoveEntity(e ecs.Entity, oldPos, newPos GridPos) {
	if oldPos == newPos {
		return
	}
	for i, entity := range g.grid[oldPos] {
		if entity == e {
			g.grid[oldPos][i] = g.grid[oldPos][len(g.grid[oldPos])-1]
			g.grid[oldPos] = g.grid[oldPos][:len(g.grid[oldPos])-1]
			break
		}
	}

	g.grid[newPos] = append(g.grid[newPos], e)
}

func (g Grid) At(pos GridPos) []ecs.Entity {
	return g.grid[pos]
}

type DebugBoard struct {
	TextLines []string
}

func NewDebugBoard() *DebugBoard {
	return &DebugBoard{TextLines: make([]string, 0, 100)}
}

func (d *DebugBoard) Write(s string) {
	d.TextLines = append(d.TextLines, s)
}

func (d *DebugBoard) Writef(s string, a ...any) {
	d.TextLines = append(d.TextLines, fmt.Sprintf(s, a))
}

func (d *DebugBoard) Clean() {
	d.TextLines = d.TextLines[:0]
}

type NoDebugBoard struct {
}

func (d NoDebugBoard) Write(s string) {}

type SubTreeBoundingBox struct {
	rootNode   ecs.Entity
	parentNode ecs.Entity
	X          float64
	Y          float64
	Width      float64
	Height     float64

	dirty bool
}

type SubTreeBoundingBoxes struct {
	boundingBoxes map[ecs.Entity]*SubTreeBoundingBox
}

func (boxes *SubTreeBoundingBoxes) NodeMoved(node ecs.Entity) {
	if _, ok := boxes.boundingBoxes[node]; ok {
		boxes.boundingBoxes[node].dirty = true
		for !boxes.boundingBoxes[node].parentNode.IsZero() {
			node = boxes.boundingBoxes[node].parentNode
			boxes.boundingBoxes[node].dirty = true
		}
	}
}
