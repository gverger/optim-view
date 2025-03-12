package systems

import (
	"context"
	"time"

	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
	"github.com/phuslu/log"
)

type Systems struct {
	systems []System

	mappings     generic.Resource[Mappings]
	mouse        generic.Resource[Mouse]
	visibleWorld generic.Resource[VisibleWorld]
	camera       generic.Resource[CameraHandler]
	debugTxt     generic.Resource[DebugText]
	grid         generic.Resource[Grid]

	targetBuilder generic.Map1[Target2]

	positions       generic.Map1[Position]
	edges           *generic.Filter1[Edge]
	nodes           *generic.Filter1[Node]
	visibleElements generic.Map1[VisibleElement]
	hiddenNodes     *generic.Filter1[Node]
	hiddenEdges     *generic.Filter1[Edge]
}

func New() *Systems {
	return &Systems{systems: make([]System, 0)}
}

func (s *Systems) Initialize(w *ecs.World) {
	s.mappings = generic.NewResource[Mappings](w)
	s.mouse = generic.NewResource[Mouse](w)
	s.mouse.Add(&Mouse{})

	s.visibleWorld = generic.NewResource[VisibleWorld](w)
	s.visibleWorld.Add(&VisibleWorld{})
	s.targetBuilder = generic.NewMap1[Target2](w)
	s.positions = generic.NewMap1[Position](w)
	s.edges = generic.NewFilter1[Edge]()
	s.nodes = generic.NewFilter1[Node]()
	s.visibleElements = generic.NewMap1[VisibleElement](w)
	s.hiddenNodes = generic.NewFilter1[Node]().Without(generic.T[VisibleElement]())
	s.hiddenEdges = generic.NewFilter1[Edge]().Without(generic.T[VisibleElement]())
	s.grid = generic.NewResource[Grid](w)
	s.grid.Add(&Grid{grid: make(map[GridPos][]ecs.Entity)})

	s.debugTxt = generic.NewResource[DebugText](w)
	s.debugTxt.Add(&DebugText{})

	for _, s := range s.systems {
		s.Initialize(w)
	}
}

func (s *Systems) Add(sys System) {
	s.systems = append(s.systems, sys)
}

func (s Systems) Update(w *ecs.World) {
	ctx, cancel := context.WithTimeout(context.Background(), 96*time.Millisecond) // Should stop when at speed of 10 FPS
	defer cancel()
	// txt := s.debugTxt.Get()
	// txt.Text = ""
	for _, sys := range s.systems {
		// start := time.Now()
		sys.Update(ctx, w)
		// duration := time.Since(start)
		// txt.Text += fmt.Sprintf("%T: %dms\n", sys, duration.Milliseconds())
	}

	// rl.DrawText(txt.Text, 10, 200, 10, rl.Red)
}

func (s Systems) Close() {
	for _, sys := range s.systems {
		sys.Close()
	}
}

type System interface {
	Initialize(w *ecs.World)
	Update(ctx context.Context, w *ecs.World)
	Close()
}

func (s *Systems) ShowAll(w *ecs.World) {
	s.visibleElements.AddBatch(s.hiddenEdges.Filter(w))
	s.visibleElements.AddBatch(s.hiddenNodes.Filter(w))
}

func (s *Systems) Hide(w *ecs.World, nodeIds []uint64) {
	todelete := make(map[ecs.Entity]bool, len(nodeIds))

	for _, id := range nodeIds {
		nodeEntity := s.mappings.Get().nodeLookup[id]
		if s.visibleElements.Get(nodeEntity) != nil {
			todelete[nodeEntity] = true
		}
	}

	query := s.edges.Query(w)
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

	query := s.edges.Query(w)
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
		s.targetBuilder.Assign(e, &Target2{X: float64(newX), Y: float64(newY), Duration: 0})
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
		s.targetBuilder.Assign(e, &Target2{X: float64(newX), Y: float64(newY), Duration: 30})
	} else {
		t.X = float64(newX)
		t.Y = float64(newY)
		t.SinceTick = 0
		t.Duration = 30
	}
}
