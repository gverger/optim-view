package systems

import (
	"context"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
	"github.com/phuslu/log"
)

func NewDrawNodes(font rl.Font, nbNodes int) *DrawNodes {
	return &DrawNodes{font: font, nbNodes: nbNodes}
}

type DrawNodes struct {
	font    rl.Font
	nbNodes int

	filter        generic.Filter3[Position, Node, VisibleElement]
	hovered       generic.Resource[ecs.Entity]
	visibleWorld  generic.Resource[VisibleWorld]
	shapes        generic.Resource[[]ShapeDefinition]
	NodesTextures generic.Resource[[]rl.RenderTexture2D]
	camera        generic.Resource[CameraHandler]
}

// Close implements System.
func (d *DrawNodes) Close() {
	for _, t := range *d.NodesTextures.Get() {
		rl.UnloadRenderTexture(t)
	}
	shapes := *d.shapes.Get()
	for i := range shapes {
		rl.UnloadRenderTexture(shapes[i].Texture)
		shapes[i].rendered = false
	}
}

const (
	NodesPerTextureLine = 64
	LinesPerTexture     = 64
	NodeTextureSize     = 128
)

func (d *DrawNodes) Initialize(w *ecs.World) {
	d.filter = *generic.NewFilter3[Position, Node, VisibleElement]()
	d.hovered = generic.NewResource[ecs.Entity](w)
	d.visibleWorld = generic.NewResource[VisibleWorld](w)
	d.camera = generic.NewResource[CameraHandler](w)
	d.shapes = generic.NewResource[[]ShapeDefinition](w)
	d.NodesTextures = generic.NewResource[[]rl.RenderTexture2D](w)
	nbTextureLines := (d.nbNodes-1)/NodesPerTextureLine + 1
	nbTextures := (nbTextureLines-1)/LinesPerTexture + 1
	textures := d.NodesTextures.Get()
	for i := 0; i < nbTextures; i++ {
		*textures = append(*textures, rl.LoadRenderTexture(NodeTextureSize*NodesPerTextureLine, int32(min(LinesPerTexture, nbTextureLines))*NodeTextureSize))
		rl.BeginTextureMode((*textures)[i])
		rl.ClearBackground(rl.Fade(rl.White, 0))
		rl.EndTextureMode()
		nbTextureLines -= LinesPerTexture
	}
}

func nodeTextureIdx(node int) int {
	return (node - 1) / (LinesPerTexture * NodesPerTextureLine)
}

func nodeTextureRec(node int) rl.Rectangle {
	n := (node - 1) % (LinesPerTexture * NodesPerTextureLine)
	x := n % NodesPerTextureLine
	y := n / NodesPerTextureLine
	return rl.NewRectangle(float32(x*NodeTextureSize), float32(y*NodeTextureSize), NodeTextureSize, NodeTextureSize)
}

func (d *DrawNodes) Update(ctx context.Context, w *ecs.World) {
	visible := d.visibleWorld.Get()
	shapes := *d.shapes.Get()
	query := d.filter.Query(w)
	nodeTextures := *(d.NodesTextures.Get())

	visibleArea := (visible.MaxX - visible.X) * (visible.MaxY - visible.Y)

	rl.BeginMode2D(*d.camera.Get().Camera)

	for query.Next() {
		pos, n, _ := query.Get()

		if pos.X > visible.MaxX || pos.Y > visible.MaxY || pos.X+n.SizeX < visible.X || pos.Y+n.SizeY < visible.Y {
			continue
		}

		rl.DrawRectangle(int32(pos.X), int32(pos.Y), int32(n.SizeX), int32(n.SizeY), rl.LightGray)

		select {
		case <-ctx.Done():
			if n.rendered {
				rec := nodeTextureRec(n.idx)
				rec.Height = -rec.Height
				texture := nodeTextures[nodeTextureIdx(n.idx)].Texture
				rec.Y = float32(texture.Height) - rec.Y - NodeTextureSize // texture is upside down...
				rl.DrawTextureRec(texture, rec, rl.NewVector2(float32(pos.X), float32(pos.Y)), rl.White)
			} else {
				rl.DrawRectangleLines(int32(pos.X), int32(pos.Y), int32(n.SizeX), int32(n.SizeY), rl.LightGray)
			}
			continue
		default:
		}

		if n.SizeX*n.SizeY < visibleArea/40 && n.rendered {
			rec := nodeTextureRec(n.idx)
			rec.Height = -rec.Height
			texture := nodeTextures[nodeTextureIdx(n.idx)].Texture
			rec.Y = float32(texture.Height) - rec.Y - NodeTextureSize // texture is upside down...
			rl.DrawTextureRec(texture, rec, rl.NewVector2(float32(pos.X), float32(pos.Y)), rl.White)
			continue
		}

		drawFast := false
		if n.SizeX*n.SizeY < visibleArea/10 {
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
			minX = min(minX, tr.X+shapeList.MinX)
			minY = min(minY, tr.Y+shapeList.MinY)
			maxX = max(maxX, shapeList.MaxX+tr.X)
			maxY = max(maxY, shapeList.MaxY+tr.Y)
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

		reverseY := float32(-1)

		if n.DrawnSizeX == 0 && n.DrawnSizeY == 0 {
			n.DrawnSizeX = float64(scale * dimX)
			n.DrawnSizeY = float64(scale * dimY)
		}

		midX := (float32(n.SizeX) - scale*dimX) / 2
		midY := (float32(n.SizeY) - reverseY*scale*dimY) / 2

		// rl.DrawRectangleLines(int32(pos.X+(n.SizeX-n.DrawnSizeX)/2), int32(pos.Y+(n.SizeY-n.DrawnSizeY)/2), int32(n.DrawnSizeX), int32(n.DrawnSizeY), rl.Blue)
		// rl.DrawText(fmt.Sprintf("%v", n.idx), int32(pos.X), int32(pos.Y), 8, rl.Maroon)

		for _, tr := range n.ShapeTransforms {
			shapeList := shapes[tr.Id]

			for i := range shapeList.Shapes {
				s := &shapes[tr.Id].Shapes[i]
				if s.Triangles == nil {
					if err := s.ComputeTriangles(); err != nil || len(s.Triangles) == 0 {
						log.Warn().Int("item", tr.Id).Int("shape index", i).Err(err).Msg("cannot triangulate")
					}
				}
			}

			if !shapeList.rendered {
				// We render shapes with an offset of 1, to be sure they are surrounded by transparent,
				// They seem to create a thin line at the border otherwise
				// beware when displaying them, we should offset the draw by 1
				// We don't do it now since we want an approximation of the drawing and it seems fine
				shapes[tr.Id].Texture = rl.LoadRenderTexture(
					int32(math.Ceil(float64(tScale*(shapeList.MaxX-shapeList.MinX))))+2,
					int32(math.Ceil(float64(tScale*(shapeList.MaxY-shapeList.MinY))))+2)
				shapes[tr.Id].rendered = true

				offsetX := -shapeList.MinX
				offsetY := -shapeList.MinY
				rl.BeginTextureMode(shapes[tr.Id].Texture)
				rl.ClearBackground(rl.Fade(rl.White, 0.0))
				for _, s := range shapeList.Shapes {
					renderShape(s, false, tScale*offsetX, float32(shapes[tr.Id].Texture.Texture.Height)-tScale*offsetY, tScale, -tScale)
				}
				rl.EndTextureMode()
			}

			if drawFast && !tr.Highlight {
				offsetX := scale*tr.X + float32(pos.X) + midX + shapeList.MinX*scale
				offsetY := reverseY*scale*tr.Y + float32(pos.Y) + midY + shapeList.MinY*scale
				if reverseY < 0 {
					offsetY += scale * float32(-shapeList.MaxY-shapeList.MinY)
				}

				rl.DrawTexturePro(shapeList.Texture.Texture,
					rl.NewRectangle(0, 0, float32(shapeList.Texture.Texture.Width), reverseY*float32(shapeList.Texture.Texture.Height)),
					rl.NewRectangle(offsetX, offsetY, scale*float32(shapeList.Texture.Texture.Width)/tScale, scale*float32(shapeList.Texture.Texture.Height)/tScale),
					rl.Vector2Zero(), 0, rl.White)
			} else {
				offsetX := midX + scale*tr.X + float32(pos.X)
				offsetY := midY + reverseY*scale*tr.Y + float32(pos.Y)

				for _, s := range shapeList.Shapes {
					renderShape(s, tr.Highlight, offsetX, offsetY, scale, reverseY*scale)
				}
			}
		}
		if !n.rendered {
			texture := nodeTextures[nodeTextureIdx(n.idx)]
			rl.BeginTextureMode(texture)
			rec := nodeTextureRec(n.idx)
			for _, tr := range n.ShapeTransforms {
				shapeList := shapes[tr.Id]
				x := midX + scale*tr.X + shapeList.MinX*scale
				y := midY + reverseY*scale*tr.Y + shapeList.MinY*reverseY*scale
				if reverseY < 0 {
					y -= scale * float32(shapeList.MaxY-shapeList.MinY)
				}

				color := rl.White
				if tr.Highlight {
					for _, s := range shapeList.Shapes {
						renderShape(s, tr.Highlight, rec.X+x-scale*shapeList.MinX, rec.Y+y+scale*shapeList.MaxY, scale, reverseY*scale)
					}
				} else {
					rl.DrawTexturePro(shapeList.Texture.Texture,
						rl.NewRectangle(0, 0, float32(shapeList.Texture.Texture.Width), reverseY*float32(shapeList.Texture.Texture.Height)),

						rl.NewRectangle(rec.X+x, rec.Y+y, scale*(shapeList.MaxX-shapeList.MinX), scale*(shapeList.MaxY-shapeList.MinY)),
						rl.Vector2Zero(), 0, color)
				}
			}
			rl.EndTextureMode()
			n.rendered = true
		}
	}

	rl.EndMode2D()
}

type ShapeColor struct {
	border color.RGBA
	fill   color.RGBA
}

type HighlightableShapeColor struct {
	normal      ShapeColor
	highlighted ShapeColor
}

var shapeColors = map[string]HighlightableShapeColor{
	"blue": HighlightableShapeColor{
		normal:      ShapeColor{border: rl.Blue, fill: rl.SkyBlue},
		highlighted: ShapeColor{border: rl.DarkGreen, fill: rl.Green},
	},
	"red": HighlightableShapeColor{
		normal:      ShapeColor{border: rl.Maroon, fill: rl.Red},
		highlighted: ShapeColor{border: rl.DarkPurple, fill: rl.Purple},
	},
	"": HighlightableShapeColor{
		normal:      ShapeColor{border: rl.Black, fill: rl.RayWhite},
		highlighted: ShapeColor{border: rl.Black, fill: rl.RayWhite},
	},
}

func renderShape(s DrawableShape, highlight bool, offsetX, offsetY, scaleX, scaleY float32) {
	col, ok := shapeColors[s.Color]
	if !ok {
		log.Fatal().Str("color", s.Color).Msg("Unknown color")
	}

	color := col.normal
	if highlight {
		color = col.highlighted
	}

	scaled := func(x float64, y float64) rl.Vector2 {
		return rl.NewVector2(scaleX*float32(x)+offsetX, scaleY*float32(y)+offsetY)
	}

	// if s.Color != "" {
	for _, t := range s.Triangles {
		// need to be counter clockwise: depends on scaleY
		if scaleX*scaleY > 0 {
			rl.DrawTriangle(scaled(t.C.X, t.C.Y), scaled(t.B.X, t.B.Y), scaled(t.A.X, t.A.Y), color.fill)
		} else {
			rl.DrawTriangle(scaled(t.A.X, t.A.Y), scaled(t.B.X, t.B.Y), scaled(t.C.X, t.C.Y), color.fill)
		}
		// Fill the holes between adjacent triangles
		rl.DrawLineStrip(
			[]rl.Vector2{
				scaled(t.A.X, t.A.Y),
				scaled(t.B.X, t.B.Y),
				scaled(t.C.X, t.C.Y),
				scaled(t.A.X, t.A.Y),
			}, color.fill)
	}
	// }
	points := make([]rl.Vector2, 0, len(s.Points))
	for _, p := range s.Points {
		points = append(points, scaled(p.X, p.Y))
	}
	points = append(points, scaled(s.Points[0].X, s.Points[0].Y))
	rl.DrawLineStrip(points, color.border)
}

var _ System = &DrawNodes{}
