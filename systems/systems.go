package systems

import (
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

type Systems struct {
	systems []System

	mappings generic.Resource[Mappings]

	targetBuilder generic.Map1[Target]
	edgeFilter    *generic.Filter1[BelongsTo]
}

func New() *Systems {
	return &Systems{systems: make([]System, 0)}
}

func (s *Systems) Initialize(w *ecs.World) {
	s.mappings = generic.NewResource[Mappings](w)
	s.targetBuilder = generic.NewMap1[Target](w)
	s.edgeFilter = generic.NewFilter1[BelongsTo]().WithRelation(generic.T[BelongsTo]())

	for _, s := range s.systems {
		s.Initialize(w)
	}
}

func (s *Systems) Add(sys System) {
	s.systems = append(s.systems, sys)
}

func (s Systems) Update(w *ecs.World) {
	for _, sys := range s.systems {
		sys.Update(w)
	}
}

type System interface {
	Initialize(w *ecs.World)
	Update(w *ecs.World)
}

func (s Systems) MoveNode(w *ecs.World, nodeId uint64, newX, newY int) {
	e := s.mappings.Get().nodeLookup[nodeId]
	if t := s.targetBuilder.Get(e); t == nil {
		s.targetBuilder.Assign(e, &Target{X: float64(newX), Y: float64(newY)})
	} else {
		t.X = float64(newX)
		t.Y = float64(newY)
		t.SinceTick = 0
	}
}

func (s Systems) MoveEdge(w *ecs.World, edgeId [2]uint64, newX [][2]int) {
	e := s.mappings.Get().edgeLookup[edgeId]

	query := s.edgeFilter.Query(w, e)
	defer query.Close()

	for query.Next() {
		e := query.Entity()
		if t := s.targetBuilder.Get(e); t == nil {
			s.targetBuilder.Assign(e, &Target{X: float64(newX), Y: float64(newY)})
		} else {
			t.X = float64(newX)
			t.Y = float64(newY)
			t.SinceTick = 0
		}
	}
}
