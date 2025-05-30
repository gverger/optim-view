package main

import (
	"embed"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graph"
	"github.com/gverger/optimview/systems"
	"github.com/mlange-42/ark/ecs"
	"github.com/ncruces/zenity"
	"github.com/phuslu/log"
)

//go:embed data/Roboto.ttf
var f embed.FS

type SceneID uint

const (
	ExitID      SceneID = 0
	TreeSceneID SceneID = 1
)

func importFile(events chan<- Event, filename string) {
	graphs := loadSearchTrees(filename)
	events <- SwitchSearchTree{graphs: graphs}
}

func computePositionsAsync(events chan<- Event, tree *GraphView) {
	events <- MoveNodes{positions: computePositions(tree)}
}

func computePositions(tree *GraphView) map[uint64]graph.Position {
	positions := graph.ComputeLayeredCoordinates(*tree)
	offset := graph.Position{X: positions[0].X - rl.GetScreenWidth()/2, Y: positions[0].Y - rl.GetScreenHeight()/4}
	for i, p := range positions {
		positions[i] = graph.Position{
			X: p.X - offset.X,
			Y: p.Y - offset.Y,
		}
	}
	return positions
}

type ecosystem struct {
	sys   *systems.Systems
	world ecs.World
}

func (a app) loadTree(font rl.Font) ecosystem {
	tree := a.trees[a.currentTree]
	positions := computePositions(tree.Tree)

	sys := systems.New(config.DebugMode)
	sys.Add(systems.NewDebug(font, 16))
	sys.Add(systems.NewInitializer(tree, positions))
	sys.Add(systems.NewGeometryCache())
	sys.Add(systems.NewTargeter())
	sys.Add(systems.NewViewport())
	sys.Add(systems.NewMouseSelector())
	sys.Add(systems.NewDrawEdges(font))
	sys.Add(systems.NewDrawNodes(font, len(tree.Tree.Nodes)))
	sys.Add(systems.NewNodeDetails(font))
	sys.Add(systems.NewTreeNavigator())
	w := ecs.NewWorld()
	sys.Initialize(&w)

	return ecosystem{
		sys:   sys,
		world: w,
	}
}

type app struct {
	events      chan Event
	treeNames   []string
	trees       []systems.SearchTree
	shapes      []ShapeDesc
	currentTree int32
}

var lastOpenFile = ""

func newApp(trees map[string]systems.SearchTree) app {
	events := make(chan Event, 1)

	if len(trees) == 0 {
		files, err := zenity.SelectFileMultiple(
			zenity.Title("Search Tree Explorer"),
			zenity.Filename(lastOpenFile),
			zenity.FileFilters{
				{Name: "Tree file", Patterns: []string{"*.json", "*.json.gz", "*.tar.gz", "*.tgz"}, CaseFold: true},
			})
		if err != nil {
			log.Error().Err(err).Msg("opening file")
		}
		if err == nil {
			trees = make(map[string]systems.SearchTree)
			for _, f := range files {
				lastOpenFile = f
				filetrees := loadSearchTrees(f)
				for k, v := range filetrees {
					trees[k] = v
				}
			}
		}
	}

	inputKeys := Keys(trees)
	sort.Strings(inputKeys)

	treeArray := make([]systems.SearchTree, len(trees))
	for i := 0; i < len(treeArray); i++ {
		treeArray[i] = trees[inputKeys[i]]
	}

	return app{
		events:      events,
		treeNames:   inputKeys,
		trees:       treeArray,
		currentTree: 0,
	}
}

type Engine interface {
	Step() Scene
}

type IScene interface {
	Id() SceneID
	Update() SceneID
	Draw()
}

type Scene struct {
	ID SceneID
}

func (s Scene) Id() SceneID {
	return s.ID
}

func runVisu(input Input) {

	app := newApp(input.Trees)

	// rl.SetConfigFlags(rl.TextureFilterNearestMipLinear)

	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagWindowMaximized | rl.FlagMsaa4xHint)
	rl.InitWindow(1600, 1000, "Graph Visualization")
	defer rl.CloseWindow()

	// rl.SetConfigFlags(rl.FlagFullscreenMode)
	// rl.ToggleFullscreen()
	fontData := Must(f.ReadFile("data/Roboto.ttf"))
	font := rl.LoadFontFromMemory(".ttf", fontData, 32, nil)

	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	rl.GenTextureMipmaps(&font.Texture)

	rl.SetTargetFPS(60)

	var scene IScene = NewTreeScene(app, font)
	for !rl.WindowShouldClose() {
		nextSceneID := scene.Update()
		if nextSceneID == ExitID {
			return
		}
		scene.Draw()
	}
}

type Event any

type MoveNodes struct {
	positions map[uint64]graph.Position
}

type SwitchSearchTree struct {
	graphs map[string]systems.SearchTree
}
