package systems

import (
	"context"
	"fmt"
	"time"

	"github.com/mlange-42/ark/ecs"
	"github.com/phuslu/log"
)

type Systems struct {
	systems   []System
	debug     System
	debugMode bool

	debugBoard   ecs.Resource[DebugBoard]
	mappings     ecs.Resource[Mappings]
	mouse        ecs.Resource[Mouse]
	visibleWorld ecs.Resource[VisibleWorld]
	boundaries   ecs.Resource[Boundaries]
	camera       ecs.Resource[CameraHandler]
	grid         ecs.Resource[Grid]

	targetBuilder *ecs.Map1[Target2]

	positions       *ecs.Map1[Position]
	edges           *ecs.Filter1[Edge]
	nodes           *ecs.Filter1[Node]
	visibleElements *ecs.Map1[VisibleElement]
	hiddenNodes     *ecs.Filter1[Node]
	hiddenEdges     *ecs.Filter1[Edge]
}

func New(debugMode bool) *Systems {
	return &Systems{
		systems:   make([]System, 0),
		debugMode: debugMode,
	}
}

func (s *Systems) Initialize(w *ecs.World) {
	s.mappings = ecs.NewResource[Mappings](w)
	s.mouse = ecs.NewResource[Mouse](w)
	s.mouse.Add(&Mouse{})

	s.visibleWorld = ecs.NewResource[VisibleWorld](w)
	s.visibleWorld.Add(&VisibleWorld{})
	s.boundaries = ecs.NewResource[Boundaries](w)
	s.boundaries.Add(&Boundaries{})
	s.targetBuilder = ecs.NewMap1[Target2](w)
	s.positions = ecs.NewMap1[Position](w)
	s.edges = ecs.NewFilter1[Edge](w)
	s.nodes = ecs.NewFilter1[Node](w)
	s.visibleElements = ecs.NewMap1[VisibleElement](w)
	s.hiddenNodes = ecs.NewFilter1[Node](w).Without(ecs.C[VisibleElement]())
	s.hiddenEdges = ecs.NewFilter1[Edge](w).Without(ecs.C[VisibleElement]())
	s.grid = ecs.NewResource[Grid](w)
	s.grid.Add(&Grid{grid: make(map[GridPos][]ecs.Entity)})

	s.debugBoard = ecs.NewResource[DebugBoard](w)
	s.debugBoard.Add(NewDebugBoard())
	s.debug.Initialize(w)

	for _, s := range s.systems {
		s.Initialize(w)
	}
}

func (s *Systems) Add(sys System) {
	switch sys.(type) {
	case *Debug:
		s.debug = sys
	default:
		s.systems = append(s.systems, sys)
	}
}

type IDebugBoard interface {
	Write(s string)
}

func (s Systems) Update(w *ecs.World) {
	ctx, cancel := context.WithTimeout(context.Background(), 96*time.Millisecond) // Should stop when at speed of 10 FPS
	defer cancel()
	var board IDebugBoard = NoDebugBoard{}
	if s.debugMode {
		board = s.debugBoard.Get()
	}
	for _, sys := range s.systems {
		start := time.Now()
		sys.Update(ctx, w)
		duration := time.Since(start)
		board.Write(fmt.Sprintf("%T: %dms", sys, duration.Milliseconds()))
	}
	s.debug.Update(ctx, w)
}

func (s Systems) Close() {
	for _, sys := range s.systems {
		sys.Close()
	}
	s.debug.Close()
}

type System interface {
	Initialize(w *ecs.World)
	Update(ctx context.Context, w *ecs.World)
	Close()
}

func (s *Systems) ShowAll(w *ecs.World) {
	s.visibleElements.AddBatch(s.hiddenEdges.Batch(), &VisibleElement{})
	s.visibleElements.AddBatch(s.hiddenNodes.Batch(), &VisibleElement{})
}

func (s *Systems) Hide(w *ecs.World, nodeIds []uint64) {
	todelete := make(map[ecs.Entity]bool, len(nodeIds))

	for _, id := range nodeIds {
		nodeEntity := s.mappings.Get().nodeLookup[id]
		if s.visibleElements.Get(nodeEntity) != nil {
			todelete[nodeEntity] = true
		}
	}

	query := s.edges.Query()
	deleted := 0
	total := 0
	for query.Next() {
		e := query.Get()
		ent := query.Entity()
		if s.visibleElements.Get(ent) == nil {
			continue
		}
		total++

		if todelete[e.From] || todelete[e.To] {
			todelete[ent] = true
			deleted++
		}
	}
	log.Info().Int("deleted edges", deleted).Int("total", total).Msg("Hide")

	for e := range todelete {
		s.visibleElements.Remove(e)
	}
}

func (s *Systems) Delete(w *ecs.World, nodeId uint64) {
	nodeEntity := s.mappings.Get().nodeLookup[nodeId]
	s.visibleElements.Remove(nodeEntity)

	query := s.edges.Query()
	for query.Next() {
		e := query.Get()
		if e.From == nodeEntity || e.To == nodeEntity {
			ent := query.Entity()
			if s.visibleElements.Get(ent) != nil {
				s.visibleElements.Remove(ent)
			}
		}
	}

}

func (s Systems) SamePositions(nodeId uint64, oldSystems Systems) {
	e, ok := s.mappings.Get().nodeLookup[nodeId]
	if !ok {
		return
	}

	oldE, ok := oldSystems.mappings.Get().nodeLookup[nodeId]
	if !ok {
		return
	}

	pos := oldSystems.positions.Get(oldE)

	p := s.positions.Get(e)
	p.X = pos.X
	p.Y = pos.Y

	s.MoveNode(nil, nodeId, int(pos.X), int(pos.Y))
}

func (s Systems) SetNodePos(w *ecs.World, nodeId uint64, newX, newY int) {
	e := s.mappings.Get().nodeLookup[nodeId]
	if t := s.targetBuilder.Get(e); t == nil {
		s.targetBuilder.Add(e, &Target2{X: float64(newX), Y: float64(newY), Duration: 0})
	} else {
		t.X = float64(newX)
		t.Y = float64(newY)
		t.SinceTick = 0
		t.Duration = 0
	}
}

func (s Systems) MoveNode(w *ecs.World, nodeId uint64, newX, newY int) {
	e := s.mappings.Get().nodeLookup[nodeId]
	if t := s.targetBuilder.Get(e); t == nil {
		s.targetBuilder.Add(e, &Target2{X: float64(newX), Y: float64(newY), Duration: 30})
	} else {
		t.X = float64(newX)
		t.Y = float64(newY)
		t.SinceTick = 0
		t.Duration = 30
	}
}
