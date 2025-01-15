package systems

import (
	"math"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
	"github.com/osuushi/triangulate"
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

	cpt := 0
	for query.Next() {
		pos, n := query.Get()

		if pos.X > visible.MaxX || pos.Y > visible.MaxY || pos.X+n.SizeX < visible.X || pos.Y+n.SizeY < visible.Y {
			continue
		}
		cpt++

		// color := n.color
		// if d.hovered.Has() && *d.hovered.Get() == query.Entity() {
		// 	color = rl.Green
		// }
		// rl.DrawRectangle(int32(pos.X), int32(pos.Y), int32(n.SizeX), int32(n.SizeY), color)
		rl.DrawTextEx(d.font, n.Text, rl.NewVector2(float32(pos.X), float32(pos.Y)), 8, 0, rl.Black)

		minX := float32(math.MaxFloat32)
		minY := float32(math.MaxFloat32)
		maxX := float32(-math.MaxFloat32)
		maxY := float32(-math.MaxFloat32)
		for _, tr := range n.ShapeTransforms {
			shapeList := shapes[tr.Id]
			minX = min(minX, shapeList.MinX+tr.X)
			minY = min(minY, shapeList.MinY+tr.Y)
			maxX = max(maxX, shapeList.MaxX+tr.X)
			maxY = max(maxY, shapeList.MaxX+tr.Y)
		}

		dimX := maxX - minX
		dimY := maxY - minY
		scale := float32(1)
		if dimX > dimY {
			scale = 100.0 / dimX
		} else {
			scale = 100.0 / dimY
		}

		// scale := int32(1)
		for _, tr := range n.ShapeTransforms {
			shapeList := shapes[tr.Id]
			// if !shapeList.rendered {
			// 	shapes[tr.Id].Texture = rl.LoadRenderTexture(
			// 		scale*int32(shapeList.MaxX-shapeList.MinX)+1,
			// 		scale*int32(shapeList.MaxY-shapeList.MinY)+1)
			// 	shapes[tr.Id].rendered = true

			offsetX := scale*tr.X + float32(pos.X)
			offsetY := scale*tr.Y + float32(pos.Y)
			// offsetX := -shapeList.MinX
			// offsetY := -shapeList.MinY
			// rl.BeginTextureMode(shapes[tr.Id].Texture)
			{
				for i, s := range shapeList.Shapes {
					if s.Triangles == nil {
						points := make([]*triangulate.Point, 0, len(s.Points))
						for _, p := range s.Points {
							points = append(points, &triangulate.Point{X: p.X, Y: p.Y})
						}
						slices.Reverse(points)
						triangles, err := triangulate.Triangulate(points)
						if err != nil {
							slices.Reverse(points)
							triangles, err = triangulate.Triangulate(points)
						}
						if err != nil {
							triangles = make([]*triangulate.Triangle, 0)
						}
						shapeList.Shapes[i].Triangles = triangles
					}

					color := rl.Green
					switch s.Color {
					case "blue":
						color = rl.Blue
					case "red":
						color = rl.Red
					case "":
						color = rl.Black
					}
					if s.Color != "" {
						for _, t := range s.Triangles {
							rl.DrawTriangle(
								rl.NewVector2(scale*float32(t.C.X)+offsetX, scale*float32(t.C.Y)+offsetY),
								rl.NewVector2(scale*float32(t.B.X)+offsetX, scale*float32(t.B.Y)+offsetY),
								rl.NewVector2(scale*float32(t.A.X)+offsetX, scale*float32(t.A.Y)+offsetY),
								color,
							)
							// Fill the holes between adjacent triangles
							rl.DrawLineStrip(
								[]rl.Vector2{
									rl.NewVector2(scale*float32(t.A.X)+offsetX, scale*float32(t.A.Y)+offsetY),
									rl.NewVector2(scale*float32(t.B.X)+offsetX, scale*float32(t.B.Y)+offsetY),
									rl.NewVector2(scale*float32(t.C.X)+offsetX, scale*float32(t.C.Y)+offsetY),
									rl.NewVector2(scale*float32(t.A.X)+offsetX, scale*float32(t.A.Y)+offsetY),
								}, color)
						}
					}
					points := make([]rl.Vector2, 0, len(s.Points))
					for _, p := range s.Points {
						points = append(points, rl.NewVector2(scale*float32(p.X)+offsetX, scale*float32(p.Y)+offsetY))
					}
					points = append(points, rl.NewVector2(scale*float32(s.Points[0].X)+offsetX, scale*float32(s.Points[0].Y)+offsetY))
					rl.DrawLineStrip(points, rl.Black)

				}
			}
			// rl.EndTextureMode()

			// }

			// offsetX := tr.X + float32(pos.X) + shapeList.MinX
			// offsetY := tr.Y + float32(pos.Y) + shapeList.MinY
			// rl.DrawTextureEx(shapeList.Texture.Texture, rl.NewVector2(offsetX, offsetY), 0, 0.1, rl.Black)

		}
	}
	// rl.DrawText(fmt.Sprintf("%d nodes", cpt), 10, int32(rl.GetScreenHeight())-100, 12, rl.Red)
}

var _ System = &DrawNodes{}
