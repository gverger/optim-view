package systems

import (
	"context"
	"time"

	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

type Systems struct {
	systems []System

	mappings     generic.Resource[Mappings]
	mouse        generic.Resource[Mouse]
	visibleWorld generic.Resource[VisibleWorld]
	camera       generic.Resource[CameraHandler]

	targetBuilder generic.Map1[Target2]

	positions generic.Map1[Position]
	edges     *generic.Filter1[Edge]
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
	for _, sys := range s.systems {
		sys.Update(ctx, w)
	}
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

func (s *Systems) Delete(w *ecs.World, nodeId uint64) {
	nodeEntity := s.mappings.Get().nodeLookup[nodeId]
	w.RemoveEntity(nodeEntity)

	query := s.edges.Query(w)
	toDelete := make([]ecs.Entity, 0)
	for query.Next() {
		e := query.Get()
		if e.From == nodeEntity || e.To == nodeEntity {
			toDelete = append(toDelete, query.Entity())
		}
	}

	for _, e := range toDelete {
		w.RemoveEntity(e)
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
