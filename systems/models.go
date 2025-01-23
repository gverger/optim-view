package systems

import (
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graph"
	"github.com/osuushi/triangulate"
	"github.com/phuslu/log"
	"github.com/tchayen/triangolatte"
)

type ShapeTransform struct {
	Id        int
	X         float32
	Y         float32
	Highlight bool
}

type DisplayableNode struct {
	Id   uint64
	Text string

	Transform []ShapeTransform
}

type DrawableShape struct {
	Color  string
	Points []Position
	Holes  [][]Position

	Triangles []*triangulate.Triangle
}

func (s *DrawableShape) computeTrianglesWithTriangolatte() error {
	points := make([]triangolatte.Point, 0, len(s.Points))
	for _, p := range s.Points {
		points = append(points, triangolatte.Point{X: p.X, Y: p.Y})
	}
	trPoints, err := triangolatte.Polygon(points)
	if err != nil {
		return err
	}
	s.Triangles = nil
	for i := 0; i < len(trPoints); i += 6 {
		s.Triangles = append(s.Triangles, &triangulate.Triangle{
			A: &triangulate.Point{X: trPoints[i], Y: trPoints[i+1]},
			B: &triangulate.Point{X: trPoints[i+2], Y: trPoints[i+3]},
			C: &triangulate.Point{X: trPoints[i+4], Y: trPoints[i+5]},
		})
	}

	return err

}

func (s *DrawableShape) ComputeTriangles() error {
	if err := s.computeTrianglesWithTriangolatte(); err == nil {
		return nil
	}
	points := make([]*triangulate.Point, 0, len(s.Points))
	for _, p := range s.Points {
		points = append(points, &triangulate.Point{X: p.X, Y: p.Y})
	}
	triangles, err := triangulate.Triangulate(points)
	if err != nil || len(triangles) == 0 {
		slices.Reverse(points)
		triangles, err = triangulate.Triangulate(points)
		if err != nil {
			log.Error().Err(err).Msg("computeTriangles second")
			triangles = make([]*triangulate.Triangle, 0)
		}
	}
	s.Triangles = triangles
	return err
}

func counterClockwise(t *triangulate.Triangle) *triangulate.Triangle {
	if t.SignedArea() < 0 {
		return &triangulate.Triangle{
			A: t.C,
			B: t.B,
			C: t.A,
		}
	}
	return t
}

type ShapeDefinition struct {
	Shapes []DrawableShape
	MinX   float32
	MinY   float32
	MaxX   float32
	MaxY   float32

	Texture  rl.RenderTexture2D
	rendered bool
}

type SearchTree struct {
	Tree   *graph.Graph[*DisplayableNode, uint64]
	Shapes []ShapeDefinition
}
