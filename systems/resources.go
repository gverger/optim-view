package systems

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/ark/ecs"
)

type Mappings struct {
	nodeLookup map[uint64]ecs.Entity
	edgeLookup map[[2]uint64]ecs.Entity
}

type Input struct {
	Active bool

	Mouse  Mouse

	KeyPressed Keyboard
}

type Keyboard struct {
	Down  bool
	Up    bool
	Right bool
	Left  bool
	Space bool
}

type Mouse struct {
	OnScreen         Position
	InWorld          Position
	Delta            Position
	VerticalScroll   float32
	HorizontalScroll float32

	LeftButton  MouseButton
	RightButton MouseButton
}

type MouseButton struct {
	Pressed bool
	Down    bool
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
	d.TextLines = append(d.TextLines, fmt.Sprintf(s, a...))
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
