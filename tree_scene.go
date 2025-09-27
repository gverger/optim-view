package main

import (
	"strconv"
	"strings"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/systems"

	"github.com/ncruces/zenity"
	"github.com/phuslu/log"
)

type TreeScene struct {
	Scene

	engine *treeEngine
}

func (t *TreeScene) Draw() {
	t.engine.Step()
}

func (t *TreeScene) Update() SceneID {
	if nextId := t.engine.handleEvents(); nextId != t.ID {
		return nextId
	}
	return t.ID
}

func NewTreeScene(app app, font rl.Font) *TreeScene {
	return &TreeScene{
		Scene: Scene{
			ID: TreeSceneID,
		},
		engine: &treeEngine{
			font:          font,
			app:           app,
			ecosystem:     app.loadTree(font),
			allNodes:      true,
			editMode:      false,
			nodeToFind:    "",
			findMode:      false,
			uiTexture:     rl.LoadRenderTexture(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())),
			mouseCaptured: false,
		},
	}
}

type treeEngine struct {
	font rl.Font
	app  app

	ecosystem ecosystem
	allNodes  bool

	editMode bool

	nodeToFind string
	findMode   bool

	uiTexture     rl.RenderTexture2D
	mouseCaptured bool
}

func (e *treeEngine) handleEvents() SceneID {
	found := true
	for found {
		select {
		case event := <-e.app.events:
			log.Info().Interface("event", event).Msg("event received")
			switch event := event.(type) {
			case MoveNodes:
				for _, node := range e.app.trees[e.app.currentTree].Tree.Nodes {
					if pos, ok := event.positions[node.Id]; ok {
						e.ecosystem.sys.MoveNode(&e.ecosystem.world, node.Id, pos.X, pos.Y)
					}
				}
			case SwitchSearchTree:
				close(e.app.events)
				e.app = newApp(event.graphs)
				e.ecosystem.sys.Close()
				e.ecosystem = e.app.loadTree(e.font)
				e.allNodes = true
			}

		default:
			found = false
		}
	}
	return TreeSceneID
}

func navButton(text string) rl.Vector2 {
	size := rl.MeasureTextEx(rl.GetFontDefault(), text, 10, 1)
	size.X += 20 // padding 10 left and right
	return size
}

// Return true if mouse captured
func (e *treeEngine) drawUI() {
	if e.uiTexture.Texture.Width != int32(rl.GetScreenWidth()) || e.uiTexture.Texture.Height != int32(rl.GetScreenHeight()) {
		rl.UnloadTexture(e.uiTexture.Texture)
		e.uiTexture = rl.LoadRenderTexture(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
	}
	rl.BeginTextureMode(e.uiTexture)
	rl.ClearBackground(rl.Fade(rl.White, 0.0))

	navRec := rl.NewRectangle(0, 0, float32(rl.GetScreenWidth()), 40)
	rl.DrawRectangleRec(navRec, rl.NewColor(246, 248, 250, 255))
	rl.DrawLineEx(
		rl.NewVector2(navRec.X, navRec.Y+navRec.Height),
		rl.NewVector2(navRec.X+navRec.Width, navRec.Y+navRec.Height),
		1, rl.NewColor(209, 217, 224, 255))

	offsetX := 10.0

	// load file
	loadFileSize := navButton("Load File")
	loadFileRec := rl.NewRectangle(float32(offsetX), 2, loadFileSize.X, navRec.Height-4)
	if gui.Button(loadFileRec, "Load File") {
		file, err := zenity.SelectFile(
			zenity.Title("Search Tree Explorer"),
			zenity.Filename(lastOpenFile),
			zenity.FileFilters{
				{Name: "Tree file", Patterns: []string{"*.json", "*.json.gz", "*.tar.gz", "*.tgz"}, CaseFold: true},
			})
		if err != nil {
			log.Info().Err(err).Str("file", file).Msg("importing")
		} else {
			log.Info().Str("file", file).Msg("importing...")
			lastOpenFile = file
			go importFile(e.app.events, file)
		}
	}

	offsetX += float64(loadFileRec.Width) + 10

	// Reload
	reloadButtonSize := navButton("Reload File")
	reloadButtonRec := rl.NewRectangle(float32(offsetX), 2, reloadButtonSize.X, navRec.Height-4)
	if gui.Button(reloadButtonRec, "Reload File") {
		log.Info().Str("file", lastOpenFile).Msg("importing...")
		go importFile(e.app.events, lastOpenFile)
	}
	offsetX += float64(reloadButtonRec.Width) + 10

	if e.editMode {
		gui.Lock()
	}

	// Drop down
	at := e.app.currentTree

	dropDownSize := navButton(e.app.treeNames[0])
	for _, name := range e.app.treeNames[1:] {
		size := navButton(name)
		if size.X > dropDownSize.X {
			dropDownSize = size
		}
	}
	dropDownSize.X += 20 // some room for the arrow on the right
	dropDownRec := rl.NewRectangle(float32(offsetX), 2, dropDownSize.X, navRec.Height-4)
	if gui.DropdownBox(dropDownRec, strings.Join(e.app.treeNames, ";"), &e.app.currentTree, e.editMode) {
		if e.editMode {
			if at != e.app.currentTree {
				e.ecosystem.sys.Close()
				e.ecosystem = e.app.loadTree(e.font)
				e.allNodes = true
			}
		}
		e.editMode = !e.editMode
	}
	offsetX += float64(dropDownRec.Width) + 10

	showAllTxt := "Show all Nodes"
	if e.allNodes {
		showAllTxt = "Nodes with children"
	}
	allChildrenSize := navButton(showAllTxt)
	allChildrenRec := rl.NewRectangle(float32(offsetX), 2, allChildrenSize.X, navRec.Height-4)
	if gui.Button(allChildrenRec, showAllTxt) {

		currentTree := e.app.trees[e.app.currentTree]
		if e.allNodes {
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

			e.ecosystem.sys.Hide(&e.ecosystem.world, toHide)

			go computePositionsAsync(e.app.events, tree.Tree)
		} else {
			e.ecosystem.sys.ShowAll(&e.ecosystem.world)

			go computePositionsAsync(e.app.events, currentTree.Tree)
		}

		e.allNodes = !e.allNodes
	}

	rightOffsetX := float32(rl.GetScreenWidth())
	findButtonSize := float32(36.0)
	rightOffsetX -= float32(findButtonSize) + 10
	findButtonRec := rl.NewRectangle(rightOffsetX, 2, findButtonSize, 36)
	if gui.Button(findButtonRec, gui.IconText(gui.ICON_LENS_BIG, "")) || (e.findMode && rl.IsKeyPressed(rl.KeyEnter)) {
		log.Info().Str("node to find", e.nodeToFind).Msg("FIND: ")
		id, err := strconv.Atoi(e.nodeToFind)
		if err == nil {
			e.ecosystem.sys.GoToNode(uint64(id))
			e.nodeToFind = ""
		}
	}

	findSize := float32(100.0)
	rightOffsetX -= findSize - 1
	findRec := rl.NewRectangle(rightOffsetX, 2, findSize, 36)
	rl.DrawRectangleRec(findRec, rl.LightGray)
	if gui.TextBox(findRec, &e.nodeToFind, 10, e.findMode) {
		e.findMode = !e.findMode
	}

	findLabelSize := navButton("Find Node ID")
	rightOffsetX -= findLabelSize.X + 5
	findLabelRec := rl.NewRectangle(rightOffsetX, 2, findLabelSize.X, 36)
	gui.Label(findLabelRec, "Find Node ID")

	if e.nodeToFind != "" {
		nodeId, err := strconv.Atoi(e.nodeToFind)
		if err != nil || !e.ecosystem.sys.HasNode(uint64(nodeId)) {
			rl.DrawRectangleLinesEx(findRec, 3, rl.Red)
		}
	}

	gui.Unlock()
	rl.DrawFPS(10, int32(rl.GetScreenHeight())-20)

	rl.EndTextureMode()

	e.mouseCaptured = e.findMode ||
		rl.CheckCollisionPointRec(rl.GetMousePosition(), findRec) ||
		rl.CheckCollisionPointRec(rl.GetMousePosition(), allChildrenRec) ||
		rl.CheckCollisionPointRec(rl.GetMousePosition(), loadFileRec) ||
		rl.CheckCollisionPointRec(rl.GetMousePosition(), reloadButtonRec)
}

func (e *treeEngine) Step() SceneID {
	e.drawUI()

	if e.mouseCaptured {
		e.ecosystem.sys.CaptureInput()
	} else {
		e.ecosystem.sys.ReleaseInput()
	}

	rl.BeginDrawing()
	rl.ClearBackground(rl.White)

	e.ecosystem.sys.Update(&e.ecosystem.world)

	rl.DrawTextureRec(e.uiTexture.Texture, rl.NewRectangle(0, 0, float32(rl.GetScreenWidth()), -float32(rl.GetScreenHeight())), rl.Vector2Zero(), rl.White)

	rl.EndDrawing()

	return TreeSceneID
}
