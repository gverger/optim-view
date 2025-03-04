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
	edgeFilter    *generic.Filter1[JointOf]
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
	s.edgeFilter = generic.NewFilter1[JointOf]().WithRelation(generic.T[JointOf]())

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

func (s Systems) MoveEdge(w *ecs.World, edgeId [2]uint64, newPos [][2]int) {
	e := s.mappings.Get().edgeLookup[edgeId]

	query := s.edgeFilter.Query(w, e)

	joints := make([]ecs.Entity, query.Count())
	query.Count()
	for query.Next() {
		e := query.Entity()
		joint := query.Get()
		joints[joint.Order] = e
	}

	for i, e := range joints {
		newX := newPos[i][0]
		newY := newPos[i][1]
		if t := s.targetBuilder.Get(e); t == nil {
			s.targetBuilder.Assign(e, &Target2{X: float64(newX), Y: float64(newY), Duration: 30})
		} else {
			t.X = float64(newX)
			t.Y = float64(newY)
			t.SinceTick = 0
			t.Duration = 30
		}
	}
}
