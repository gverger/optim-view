package systems

import (
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

type Systems struct {
	systems []System

	mappings     generic.Resource[Mappings]
	mouse        generic.Resource[Mouse]
	visibleWorld generic.Resource[VisibleWorld]

	targetBuilder generic.Map1[Target]
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
	s.targetBuilder = generic.NewMap1[Target](w)
	s.edgeFilter = generic.NewFilter1[JointOf]().WithRelation(generic.T[JointOf]())

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

func (s Systems) SetMouse(windowX, windowY, worlX, worldY float64) {
	s.mouse.Get().InWorld = Position{worlX, worldY}
	s.mouse.Get().OnScreen = Position{windowX, windowY}
}

func (s Systems) SetVisibleWorld(x, y, maxX, maxY float64) {
	r := s.visibleWorld.Get()
	r.X = x
	r.Y = y
	r.MaxX = maxX
	r.MaxY = maxY
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
			s.targetBuilder.Assign(e, &Target{X: float64(newX), Y: float64(newY)})
		} else {
			t.X = float64(newX)
			t.Y = float64(newY)
			t.SinceTick = 0
		}
	}
}
