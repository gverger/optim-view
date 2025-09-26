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

// Return true if mouse captured
func (e *treeEngine) drawUI() {
	rl.BeginTextureMode(e.uiTexture)
	rl.ClearBackground(rl.Fade(rl.White, 0.0))

	reloadButtonRec := rl.NewRectangle(float32(rl.GetScreenWidth()-380), 20, 150, 48)
	if gui.Button(reloadButtonRec, "reload file") {
		log.Info().Str("file", lastOpenFile).Msg("importing...")
		go importFile(e.app.events, lastOpenFile)
	}

	loadFileRec := rl.NewRectangle(float32(rl.GetScreenWidth()-200), 20, 150, 48)
	if gui.Button(loadFileRec, "load file") {

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

	if e.editMode {
		gui.Lock()
	}

	at := e.app.currentTree

	dropDownRec := rl.NewRectangle(10, 10, 200, 30)
	if gui.DropdownBox(dropDownRec, strings.Join(e.app.treeNames, ";"), &e.app.currentTree, e.editMode) {
		log.Info().Int("active", int(e.app.currentTree)).Msg("DropdownBox")
		if e.editMode {
			if at != e.app.currentTree {
				e.ecosystem.sys.Close()
				e.ecosystem = e.app.loadTree(e.font)
				e.allNodes = true
			}

			// lastHovered = -1
		}
		e.editMode = !e.editMode
	}

	showAllTxt := "Show all Nodes"
	if e.allNodes {
		showAllTxt = "Nodes with children"
	}
	allChildrenRec := rl.NewRectangle(float32(rl.GetScreenWidth()-200), 98, 150, 48)
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

	findButtonRec := rl.NewRectangle(float32(rl.GetScreenWidth())-250, 98, 30, 48)
	if gui.Button(findButtonRec, gui.IconText(gui.ICON_LENS_BIG, "")) || (e.findMode && rl.IsKeyPressed(rl.KeyEnter)) {
		log.Info().Str("node to find", e.nodeToFind).Msg("FIND: ")
		id, err := strconv.Atoi(e.nodeToFind)
		if err == nil {
			e.ecosystem.sys.GoToNode(uint64(id))
			e.nodeToFind = ""
		}
	}

	findRec := rl.NewRectangle(float32(rl.GetScreenWidth())-400, 98, 150, 48)
	rl.DrawRectangleRec(findRec, rl.LightGray)
	if gui.TextBox(findRec, &e.nodeToFind, 10, e.findMode) {
		e.findMode = !e.findMode
	}

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
		// rl.CheckCollisionPointRec(rl.GetMousePosition(), findRec) ||
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
	rl.ClearBackground(rl.RayWhite)

	e.ecosystem.sys.Update(&e.ecosystem.world)

	rl.DrawTextureRec(e.uiTexture.Texture, rl.NewRectangle(0, 0, float32(rl.GetScreenWidth()), -float32(rl.GetScreenHeight())), rl.Vector2Zero(), rl.White)

	rl.EndDrawing()

	return TreeSceneID
}
