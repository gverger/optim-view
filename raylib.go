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

func NewCameraHandler() systems.CameraHandler {
	return systems.CameraHandler{
		Camera: &rl.Camera2D{
			Zoom: 1.0,
		},
	}
}

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

type scene struct {
	sys   *systems.Systems
	world ecs.World
}

func (a app) loadTree(font rl.Font) scene {
	tree := a.trees[a.currentTree]
	sys := systems.New()
	sys.Add(systems.NewInitializer(tree))
	sys.Add(systems.NewViewport(NewCameraHandler()))
	sys.Add(systems.NewMouseSelector())
	// sys.Add(systems.NewMover())
	sys.Add(systems.NewTargeter())
	sys.Add(systems.NewDrawEdges(font))
	sys.Add(systems.NewDrawNodes(font, len(tree.Tree.Nodes)))
	sys.Add(systems.NewNodeDetails(font))
	w := ecs.NewWorld()
	sys.Initialize(&w)

	go computePositions(a.events, tree.Tree)

	return scene{
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
		file, err := zenity.SelectFile(
			zenity.Title("Search Tree Explorer"),
			zenity.Filename(""),
			zenity.FileFilters{
				{Name: "Search Tree files", Patterns: []string{"*.json", "*.json.gz"}, CaseFold: true},
			})
		log.Info().Err(err).Str("file", file).Msg("Importing...")
		if err == nil {
			trees = loadSearchTrees(file)
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

	// currentLayer := input.Layers[inputKeys[activeTree]]
	// currentLayout := input.Layouts[inputKeys[activeTree]]

	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	// rl.SetConfigFlags(rl.TextureFilterNearestMipLinear)

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

	// lastHovered := -1
	var selectionTexture rl.Texture2D

	for !rl.WindowShouldClose() {
		found := true
		for found {
			select {
			case event := <-app.events:
				log.Info().Interface("event", event).Msg("event received")
				switch e := event.(type) {
				case MoveNodes:
					for _, node := range app.trees[app.currentTree].Tree.Nodes {
						scene.sys.MoveNode(&scene.world, node.Id, e.positions[node.Id].X, e.positions[node.Id].Y)
					}
				case SwitchSearchTree:
					close(app.events)
					app = newApp(e.graphs)
					scene.sys.Close()
					scene = app.loadTree(font)
				}

			default:
				found = false
			}
		}

		// hovered = nil
		// for i, n := range currentTree.Nodes {
		// 	if rl.CheckCollisionPointRec(worldMousePos, rl.NewRectangle(float32(n.XY[0]), float32(n.XY[1]), float32(n.W), float32(n.H))) {
		// 		hovered = currentTree.Nodes[i]
		// 		if lastHovered != int(i) {
		// 			// image := rl.LoadImageSvg(selected.SvgImage, 500, 500)
		// 			rl.UnloadTexture(selectionTexture)
		// 			// if img, ok := ImageFromSVG(selected.SvgImage); ok {
		// 			// 	selectionTexture = rl.LoadTextureFromImage(img)
		// 			// }
		// 			// rl.UnloadImage(image)
		// 			lastHovered = int(i)
		// 		}
		// 		break
		// 	}
		// }

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		scene.sys.Update(&scene.world)

		if gui.Button(rl.NewRectangle(float32(rl.GetScreenWidth()-200), 20, 150, 48), "load file") {

			file, err := zenity.SelectFile(
				zenity.Title("Search Tree Explorer"),
				zenity.Filename(""),
				zenity.FileFilters{
					{Name: "Search Tree files", Patterns: []string{"*.json", "*.json.gz"}, CaseFold: true},
				})
			log.Info().Err(err).Str("file", file).Msg("Importing...")
			if err == nil {
				go importFile(app.events, file)
			}
		}

		// if hovered != nil && !editMode {
		// 	txtDims := rl.MeasureTextEx(rl.GetFontDefault(), hovered.Text, 32, 4)
		//
		// 	shape := currentLayout.Nodes[uint64(lastHovered)]
		// 	corner := rl.GetWorldToScreen2D(rl.NewVector2(float32(shape.XY[0]+shape.W), float32(shape.XY[1])), *camera.Camera)
		//
		// 	distX := float32(50)
		// 	distY := -float32(60)
		//
		// 	rightmostPointX := corner.X + txtDims.X + 20
		//
		// 	if rightmostPointX > float32(rl.GetScreenWidth()-10) {
		// 		distX = -50 - txtDims.X - 20
		// 		corner.X = rl.GetWorldToScreen2D(rl.NewVector2(float32(shape.XY[0]), 0), *camera.Camera).X
		// 	} else if rightmostPointX+distX > float32(rl.GetScreenWidth()-10) {
		// 		distX = float32(rl.GetScreenWidth()-10) - (corner.X + txtDims.X + 20)
		// 	}
		//
		// 	offsetX := rl.Clamp(corner.X+distX, 10, float32(rl.GetScreenWidth())-10-txtDims.X-20)
		// 	offsetY := rl.Clamp(corner.Y-distY, 10, float32(rl.GetScreenHeight())-10-txtDims.Y-20)
		//
		// 	savedBackgroundColor := gui.GetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR)
		// 	gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, 0xDDDDDDDD)
		// 	gui.Panel(rl.NewRectangle(offsetX, offsetY, txtDims.X+20, txtDims.Y+20), "Properties")
		// 	rl.DrawTextEx(font, hovered.Text, rl.NewVector2(offsetX+10, offsetY+24), 32, 0, rl.Black)
		//
		// 	// rl.DrawTexture(selectionTexture, int32(offsetX+10), int32(offsetY+300), rl.White)
		//
		// 	gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, savedBackgroundColor)
		// }
		//
		if editMode {
			gui.Lock()
		}
		at := app.currentTree

		if gui.DropdownBox(rl.NewRectangle(10, 10, 200, 30), strings.Join(app.treeNames, ";"), &app.currentTree, editMode) {
			log.Info().Int("active", int(app.currentTree)).Msg("DropdownBox")
			if editMode {
				// currentLayout = input.Layouts[inputKeys[activeTree]]
				if at != app.currentTree {
					scene.sys.Close()
					scene = app.loadTree(font)
				}

				// lastHovered = -1
			}
			editMode = !editMode
		}
		gui.Unlock()
		rl.DrawFPS(10, int32(rl.GetScreenHeight())-20)

		rl.EndDrawing()
	}

	rl.UnloadTexture(selectionTexture)
}

type Event any

type MoveNodes struct {
	positions map[uint64]graph.Position
}

type SwitchSearchTree struct {
	graphs map[string]systems.SearchTree
}
