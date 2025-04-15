package main

import (
	"embed"
	"sort"
	"strings"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graph"
	"github.com/gverger/optimview/systems"
	"github.com/mlange-42/arche/ecs"
	"github.com/ncruces/zenity"
	"github.com/phuslu/log"
)

//go:embed data/Roboto.ttf
var f embed.FS

func importFile(events chan<- Event, filename string) {
	graphs := loadSearchTrees(filename)
	events <- SwitchSearchTree{graphs: graphs}
}

func computePositions(events chan<- Event, tree *GraphView) {
	positions := graph.ComputeLayeredCoordinates(*tree)
	offset := graph.Position{X: positions[0].X - rl.GetScreenWidth()/2, Y: positions[0].Y - rl.GetScreenHeight()/4}
	for i, p := range positions {
		positions[i] = graph.Position{
			X: p.X - offset.X,
			Y: p.Y - offset.Y,
		}
	}
	events <- MoveNodes{positions: positions}
}

type sceneType struct {
	sys   *systems.Systems
	world ecs.World
}

func (a app) loadTree(font rl.Font) sceneType {
	tree := a.trees[a.currentTree]
	sys := systems.New()
	sys.Add(systems.NewInitializer(tree))
	sys.Add(systems.NewTargeter())
	sys.Add(systems.NewViewport())
	sys.Add(systems.NewMouseSelector())
	sys.Add(systems.NewDrawEdges(font))
	sys.Add(systems.NewDrawNodes(font, len(tree.Tree.Nodes)))
	sys.Add(systems.NewNodeDetails(font))
	sys.Add(systems.NewTreeNavigator())
	w := ecs.NewWorld()
	sys.Initialize(&w)

	go computePositions(a.events, tree.Tree)

	return sceneType{
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

func newApp(trees map[string]systems.SearchTree) app {
	events := make(chan Event, 1)

	if len(trees) == 0 {
		files, err := zenity.SelectFileMultiple(
			zenity.Title("Search Tree Explorer"),
			zenity.Filename(""),
			zenity.FileFilters{
				{Name: "Tree file", Patterns: []string{"*.json", "*.json.gz", "*.tar.gz", "*.tgz"}, CaseFold: true},
			})
		if err != nil {
			log.Error().Err(err).Msg("opening file")
		}
		if err == nil {
			trees = make(map[string]systems.SearchTree)
			for _, f := range files {
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

func runVisu(input Input) {

	app := newApp(input.Trees)

	editMode := false

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

	scene := app.loadTree(font)

	rl.SetTargetFPS(60)

	allNodes := true

	for !rl.WindowShouldClose() {
		// Listen to events (graph changed for instance)
		found := true
		for found {
			select {
			case event := <-app.events:
				log.Info().Interface("event", event).Msg("event received")
				switch e := event.(type) {
				case MoveNodes:
					for _, node := range app.trees[app.currentTree].Tree.Nodes {
						if pos, ok := e.positions[node.Id]; ok {
							scene.sys.MoveNode(&scene.world, node.Id, pos.X, pos.Y)
						}
					}
				case SwitchSearchTree:
					close(app.events)
					app = newApp(e.graphs)
					scene.sys.Close()
					scene = app.loadTree(font)
					allNodes = true
				}

			default:
				found = false
			}
		}

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		scene.sys.Update(&scene.world)

		if gui.Button(rl.NewRectangle(float32(rl.GetScreenWidth()-200), 20, 150, 48), "load file") {

			file, err := zenity.SelectFile(
				zenity.Title("Search Tree Explorer"),
				zenity.Filename(""),
				zenity.FileFilters{
					{Name: "Tree file", Patterns: []string{"*.json", "*.json.gz", "*.tar.gz", "*.tgz"}, CaseFold: true},
				})
			log.Info().Err(err).Str("file", file).Msg("Importing...")
			if err == nil {
				go importFile(app.events, file)
			}
		}

		if editMode {
			gui.Lock()
		}

		at := app.currentTree

		if gui.DropdownBox(rl.NewRectangle(10, 10, 200, 30), strings.Join(app.treeNames, ";"), &app.currentTree, editMode) {
			log.Info().Int("active", int(app.currentTree)).Msg("DropdownBox")
			if editMode {
				if at != app.currentTree {
					scene.sys.Close()
					scene = app.loadTree(font)
					allNodes = true
				}

				// lastHovered = -1
			}
			editMode = !editMode
		}

		showAllTxt := "Show all Nodes"
		if allNodes {
			showAllTxt = "Nodes with children"
		}
		if gui.Button(rl.NewRectangle(float32(rl.GetScreenWidth()-200), 98, 150, 48), showAllTxt) {

			currentTree := app.trees[app.currentTree]
			if allNodes {
				tree := systems.SearchTree{
					Tree:   currentTree.Tree.StripNodesWithoutChildren(),
					Shapes: currentTree.Shapes,
				}

				toHide := make([]uint64, 0, len(currentTree.Tree.Nodes))
				for _, node := range currentTree.Tree.Nodes {
					if !tree.Tree.HasNode(node) {
						toHide = append(toHide, node.Id)
					}
				}

				scene.sys.Hide(&scene.world, toHide)

				go computePositions(app.events, tree.Tree)
			} else {
				scene.sys.ShowAll(&scene.world)

				go computePositions(app.events, currentTree.Tree)
			}

			allNodes = !allNodes
		}

		gui.Unlock()
		rl.DrawFPS(10, int32(rl.GetScreenHeight())-20)

		rl.EndDrawing()
	}
}

type Event any

type MoveNodes struct {
	positions map[uint64]graph.Position
}

type SwitchSearchTree struct {
	graphs map[string]systems.SearchTree
}
