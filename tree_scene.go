package main

import (
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
			font:      font,
			app:       app,
			ecosystem: app.loadTree(font),
			allNodes:  true,
			editMode:  false,
		},
	}
}

type treeEngine struct {
	font rl.Font
	app  app

	ecosystem ecosystem
	allNodes  bool

	editMode bool
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

func (e *treeEngine) Step() SceneID {
	rl.BeginDrawing()

	rl.ClearBackground(rl.RayWhite)

	e.ecosystem.sys.Update(&e.ecosystem.world)

	if gui.Button(rl.NewRectangle(float32(rl.GetScreenWidth()-200), 20, 150, 48), "load file") {

		file, err := zenity.SelectFile(
			zenity.Title("Search Tree Explorer"),
			zenity.Filename(""),
			zenity.FileFilters{
				{Name: "Tree file", Patterns: []string{"*.json", "*.json.gz", "*.tar.gz", "*.tgz"}, CaseFold: true},
			})
		log.Info().Err(err).Str("file", file).Msg("Importing...")
		if err == nil {
			go importFile(e.app.events, file)
		}
	}

	if e.editMode {
		gui.Lock()
	}

	at := e.app.currentTree

	if gui.DropdownBox(rl.NewRectangle(10, 10, 200, 30), strings.Join(e.app.treeNames, ";"), &e.app.currentTree, e.editMode) {
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
	if gui.Button(rl.NewRectangle(float32(rl.GetScreenWidth()-200), 98, 150, 48), showAllTxt) {

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

			go computePositions(e.app.events, tree.Tree)
		} else {
			e.ecosystem.sys.ShowAll(&e.ecosystem.world)

			go computePositions(e.app.events, currentTree.Tree)
		}

		e.allNodes = !e.allNodes
	}

	gui.Unlock()
	rl.DrawFPS(10, int32(rl.GetScreenHeight())-20)

	rl.EndDrawing()
	return TreeSceneID
}
