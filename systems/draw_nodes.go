package systems

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

func NewDrawNodes(font rl.Font) *DrawNodes {
	return &DrawNodes{font: font}
}

type DrawNodes struct {
	font         rl.Font
	filter       generic.Filter2[Position, Node]
	hovered      generic.Resource[ecs.Entity]
	visibleWorld generic.Resource[VisibleWorld]
	shapes       generic.Resource[[]ShapeDefinition]
}

func (d *DrawNodes) Initialize(w *ecs.World) {
	d.filter = *generic.NewFilter2[Position, Node]()
	d.hovered = generic.NewResource[ecs.Entity](w)
	d.visibleWorld = generic.NewResource[VisibleWorld](w)
	d.shapes = generic.NewResource[[]ShapeDefinition](w)
}

func (d *DrawNodes) Update(w *ecs.World) {
	visible := d.visibleWorld.Get()
	shapes := *d.shapes.Get()
	query := d.filter.Query(w)

	visibleArea := (visible.MaxX - visible.X) * (visible.MaxY - visible.Y)

	cpt := 0
	for query.Next() {
		pos, n := query.Get()

		if pos.X > visible.MaxX || pos.Y > visible.MaxY || pos.X+n.SizeX < visible.X || pos.Y+n.SizeY < visible.Y {
			continue
		}
		cpt++

		drawFast := false
		if n.SizeX*n.SizeY < visibleArea/100 {
			drawFast = true
		}

		// color := n.color
		// if d.hovered.Has() && *d.hovered.Get() == query.Entity() {
		// 	color = rl.Green
		// }
		// rl.DrawTextEx(d.font, n.Text, rl.NewVector2(float32(pos.X), float32(pos.Y)), 8, 0, rl.Black)

		minX := float32(math.MaxFloat32)
		minY := float32(math.MaxFloat32)
		maxX := float32(-math.MaxFloat32)
		maxY := float32(-math.MaxFloat32)
		for _, tr := range n.ShapeTransforms {
			shapeList := shapes[tr.Id]
			minX = min(minX, tr.X)
			minY = min(minY, tr.Y)
			maxX = max(maxX, shapeList.MaxX-shapeList.MinX+tr.X)
			maxY = max(maxY, shapeList.MaxY-shapeList.MinY+tr.Y)
			// log.Info().Int("TR", tr.Id).Float32("MINX", shapeList.MinX).Float32("MAXX", shapeList.MaxX).Float32("MINY", shapeList.MinY).Float32("MAXY", shapeList.MaxY).Msg("shape")
		}

		dimX := maxX - minX
		dimY := maxY - minY
		scale := float32(1)
		tScale := float32(1)
		if dimX > dimY {
			scale = float32(n.SizeX) / dimX
			tScale = 400 / dimX
		} else {
			scale = float32(n.SizeY) / dimY
			tScale = 400 / dimY
		}

		midX := (float32(n.SizeX) - scale*dimX) / 2
		midY := (float32(n.SizeY) - scale*dimY) / 2
		// rl.DrawRectangleLines(int32(pos.X), int32(pos.Y), int32(n.SizeX), int32(n.SizeY), rl.Green)
		// rl.DrawText(fmt.Sprintf("mid %v %v scale %v dim %v %v", midX, midY, scale, dimX, dimY), int32(pos.X), int32(pos.Y), 8, rl.Maroon)

		for _, tr := range n.ShapeTransforms {
			shapeList := shapes[tr.Id]

			for i, s := range shapeList.Shapes {
				if s.Triangles == nil {
					shapeList.Shapes[i].ComputeTriangles()
				}
			}

			if !shapeList.rendered {
				shapes[tr.Id].Texture = rl.LoadRenderTexture(
					int32(tScale*(shapeList.MaxX-shapeList.MinX))+2,
					int32(tScale*(shapeList.MaxY-shapeList.MinY))+2)
				shapes[tr.Id].rendered = true

				offsetX := -shapeList.MinX
				offsetY := -shapeList.MinY
				rl.BeginTextureMode(shapes[tr.Id].Texture)
				for _, s := range shapeList.Shapes {
					renderShape(s, tScale*offsetX+1, float32(shapes[tr.Id].Texture.Texture.Height-1)-tScale*offsetY, tScale, -tScale)
				}
				rl.EndTextureMode()
			}

			if drawFast {
				offsetX := scale*tr.X + float32(pos.X) + midX
				offsetY := scale*tr.Y + float32(pos.Y) + midY
				rl.DrawTextureEx(shapeList.Texture.Texture, rl.NewVector2(offsetX, offsetY), 0, scale/tScale, rl.White)

			} else {
				offsetX := scale*tr.X + float32(pos.X) - scale*shapeList.MinX + midX
				offsetY := scale*tr.Y + float32(pos.Y) - scale*shapeList.MinY + midY
				for _, s := range shapeList.Shapes {
					renderShape(s, offsetX, offsetY, scale, scale)
				}
			}
		}
	}
}

func renderShape(s DrawableShape, offsetX, offsetY, scaleX, scaleY float32) {
	color := rl.Green
	bgColor := rl.RayWhite
	switch s.Color {
	case "blue":
		color = rl.Blue
		bgColor = rl.SkyBlue
	case "red":
		color = rl.Red
		bgColor = rl.Maroon
	case "":
		color = rl.Black
		bgColor = rl.RayWhite
	}
	scaled := func(x float64, y float64) rl.Vector2 {
		return rl.NewVector2(scaleX*float32(x)+offsetX, scaleY*float32(y)+offsetY)
	}

	if s.Color != "" {
		for _, t := range s.Triangles {
			// need to be counter clockwise: depends on scaleY
			if scaleX*scaleY > 0 {
				rl.DrawTriangle(scaled(t.C.X, t.C.Y), scaled(t.B.X, t.B.Y), scaled(t.A.X, t.A.Y), bgColor)
			} else {
				rl.DrawTriangle(scaled(t.A.X, t.A.Y), scaled(t.B.X, t.B.Y), scaled(t.C.X, t.C.Y), bgColor)
			}
			// Fill the holes between adjacent triangles
			rl.DrawLineStrip(
				[]rl.Vector2{
					scaled(t.A.X, t.A.Y),
					scaled(t.B.X, t.B.Y),
					scaled(t.C.X, t.C.Y),
					scaled(t.A.X, t.A.Y),
				}, bgColor)
		}
	}
	points := make([]rl.Vector2, 0, len(s.Points))
	for _, p := range s.Points {
		points = append(points, scaled(p.X, p.Y))
	}
	points = append(points, scaled(s.Points[0].X, s.Points[0].Y))
	rl.DrawLineStrip(points, color)
}

var _ System = &DrawNodes{}
