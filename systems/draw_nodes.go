package systems

import (
	"context"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gverger/optimview/graphics"
	"github.com/mlange-42/ark/ecs"
	"github.com/phuslu/log"
)

const reverseY = float32(-1)

func NewDrawNodes(font rl.Font, nbNodes int) *DrawNodes {
	return &DrawNodes{font: font, nbNodes: nbNodes}
}

type DrawNodes struct {
	font         rl.Font
	nbNodes      int
	shapes       []ShapeDefinition
	nodeTextures graphics.TextureArray

	filter        ecs.Filter3[Position, Node, VisibleElement]
	visibleWorld  ecs.Resource[VisibleWorld]
	camera        ecs.Resource[CameraHandler]
	selection     ecs.Resource[NodeSelection]
}

// Close implements System.
func (d *DrawNodes) Close() {
	d.nodeTextures.Unload()
	shapes := d.shapes
	for i := range shapes {
		rl.UnloadRenderTexture(shapes[i].Texture)
		shapes[i].rendered = false
	}
}

const (
	NodeTextureSize = 100 // Nodes are 100x100

	NodeMinBorderSize = 5
)

func (d *DrawNodes) Initialize(w *ecs.World) {
	d.filter = *ecs.NewFilter3[Position, Node, VisibleElement](w)
	d.visibleWorld = ecs.NewResource[VisibleWorld](w)
	d.camera = ecs.NewResource[CameraHandler](w)
	d.selection = ecs.NewResource[NodeSelection](w)

	shapes := generic.NewResource[[]ShapeDefinition](w)
	d.shapes = *shapes.Get()

	d.nodeTextures = graphics.NewTextureArray(d.nbNodes, NodeTextureSize)

	for i := range d.shapes {
		for j := range d.shapes[i].Shapes {
			s := &d.shapes[i].Shapes[j]
			if s.Triangles == nil {
				if err := s.ComputeTriangles(); err != nil || len(s.Triangles) == 0 {
					log.Warn().Int("item", i).Int("shape index", j).Err(err).Msg("cannot triangulate")
				}
			}
		}

		dimX := d.shapes[i].MaxX - d.shapes[i].MinX
		dimY := d.shapes[i].MaxY - d.shapes[i].MinY
		// 800 seems like a good compromise: the shape is not too pixelated
		tScale := 800.0 / dimX
		if dimX < dimY {
			tScale = 800.0 / dimY
		}

		// We render shapes with an offset of 2, to be sure they are surrounded by transparent,
		// They seem to create a thin line at the border otherwise
		// beware when displaying them, we should offset the draw by 2
		// We don't do it now since we want an approximation of the drawing and it seems fine
		d.shapes[i].Texture = rl.LoadRenderTexture(int32(tScale*dimX)+4, int32(tScale*dimY)+4)
		d.shapes[i].rendered = true

		offsetX := -d.shapes[i].MinX
		offsetY := -d.shapes[i].MinY
		rl.BeginTextureMode(d.shapes[i].Texture)
		rl.ClearBackground(rl.Fade(rl.White, 0.0))
		for _, s := range d.shapes[i].Shapes {
			renderShape(s, false, tScale*offsetX+2, float32(d.shapes[i].Texture.Texture.Height)-tScale*offsetY-2, tScale, -tScale)
		}
		rl.EndTextureMode()
	}

	query := d.filter.Query(w)
	for query.Next() {
		_, n, _ := query.Get()

		minX := float32(math.MaxFloat32)
		minY := float32(math.MaxFloat32)
		maxX := float32(-math.MaxFloat32)
		maxY := float32(-math.MaxFloat32)
		for _, tr := range n.ShapeTransforms {
			shapeList := d.shapes[tr.Id]
			minX = min(minX, tr.X+shapeList.MinX)
			minY = min(minY, tr.Y+shapeList.MinY)
			maxX = max(maxX, shapeList.MaxX+tr.X)
			maxY = max(maxY, shapeList.MaxY+tr.Y)
		}

		dimX := maxX - minX
		dimY := maxY - minY
		scale := float32(1)
		if dimX > dimY {
			scale = float32(n.SizeX-NodeMinBorderSize) / dimX
		} else {
			scale = float32(n.SizeY-NodeMinBorderSize) / dimY
		}

		if n.DrawnSizeX == 0 && n.DrawnSizeY == 0 {
			n.DrawnSizeX = float64(scale * dimX)
			n.DrawnSizeY = float64(scale * dimY)
		}

		n.scale = scale
		n.midX = (float32(n.SizeX)-scale*dimX)/2 - scale*minX
		n.midY = (float32(n.SizeY)-reverseY*scale*dimY)/2 - reverseY*scale*minY
	}
}

func (d *DrawNodes) Update(ctx context.Context, w *ecs.World) {
	visible := d.visibleWorld.Get()
	query := d.filter.Query(w)
	selected := ecs.Entity{}
	hovered := ecs.Entity{}
	if d.selection.Has() {
		selected = d.selection.Get().Selected
		hovered = d.selection.Get().Hovered
	}

	visibleArea := (visible.MaxX - visible.X) * (visible.MaxY - visible.Y)

	rl.BeginMode2D(*d.camera.Get().Camera)

	toRender := make([]func(), 0)
	toRenderLater := make([]func(), 0)
	for query.Next() {
		pos, n, _ := query.Get()

		if pos.X > visible.MaxX || pos.Y > visible.MaxY || pos.X+n.SizeX < visible.X || pos.Y+n.SizeY < visible.Y {
			// render node texture if there is still time
			if !n.rendered && len(toRenderLater) < 100 {
				toRenderLater = append(toRenderLater, func() {
					d.renderNodeInTexture(n)
				})
			}
			continue
		}

		nodeColor := Palette.Background
		if hovered == query.Entity() {
			nodeColor = Palette.Hovered
		}
		if selected == query.Entity() {
			nodeColor = Palette.Selected
		}

		rl.DrawRectangle(int32(pos.X), int32(pos.Y), int32(n.SizeX), int32(n.SizeY), nodeColor)

		select {
		case <-ctx.Done():
			if n.rendered {
				d.drawOnTexture(n, pos)
			} else {
				rl.DrawRectangleLines(int32(pos.X), int32(pos.Y), int32(n.SizeX), int32(n.SizeY), rl.LightGray)
			}
			continue
		default:
		}

		if n.SizeX*n.SizeY < visibleArea/40 && n.rendered {
			d.drawOnTexture(n, pos)
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

		// rl.DrawRectangleLines(int32(pos.X+(n.SizeX-n.DrawnSizeX)/2), int32(pos.Y+(n.SizeY-n.DrawnSizeY)/2), int32(n.DrawnSizeX), int32(n.DrawnSizeY), rl.Blue)
		// rl.DrawText(fmt.Sprintf("%v", n.idx), int32(pos.X), int32(pos.Y), 8, rl.Maroon)

		for _, tr := range n.ShapeTransforms {
			shapeList := d.shapes[tr.Id]

			if drawFast && !tr.Highlight {
				d.drawShapeFromTexture(n.scale, tr, pos, n.midX, shapeList, reverseY, n.midY)
			} else {
				offsetX := n.midX + n.scale*tr.X + float32(pos.X)
				offsetY := n.midY + reverseY*n.scale*tr.Y + float32(pos.Y)

				for _, s := range shapeList.Shapes {
					renderShape(s, tr.Highlight, offsetX, offsetY, n.scale, reverseY*n.scale)
				}
			}
		}
		if !n.rendered {
			// We don't want to draw in the texture here since we are in the middle of a Mode2D
			// We delay the call then
			toRender = append(toRender, func() {
				d.renderNodeInTexture(n)
			})
		}
	}

	rl.EndMode2D()

	for _, renderNode := range toRender {
		renderNode()
	}

	for _, renderNode := range toRenderLater {
		if ctx.Err() != nil {
			break
		}
		renderNode()
	}
}

func (*DrawNodes) drawShapeFromTexture(scale float32, tr ShapeTransform, pos *Position, midX float32, shapeList ShapeDefinition, reverseY float32, midY float32) {
	offsetX := scale*tr.X + float32(pos.X) + midX + shapeList.MinX*scale
	offsetY := reverseY*scale*tr.Y + float32(pos.Y) + midY + shapeList.MinY*scale
	if reverseY < 0 {
		offsetY += scale * float32(-shapeList.MaxY-shapeList.MinY)
	}

	rl.DrawTexturePro(shapeList.Texture.Texture,
		rl.NewRectangle(2, 2, float32(shapeList.Texture.Texture.Width-4), reverseY*float32(shapeList.Texture.Texture.Height-4)),
		rl.NewRectangle(offsetX, offsetY, scale*(shapeList.MaxX-shapeList.MinX), -scale*(shapeList.MaxY-shapeList.MinY)),
		rl.Vector2Zero(), 0, rl.White)
}

func (s *DrawNodes) renderNodeInTexture(n *Node) {
	texture := s.nodeTextures.At(n.idx)
	rec := s.nodeTextures.NodeTextureRec(n.idx)
	rl.BeginTextureMode(texture)
	for _, tr := range n.ShapeTransforms {
		shapeList := s.shapes[tr.Id]
		x := n.midX + n.scale*tr.X + shapeList.MinX*n.scale
		y := n.midY + reverseY*n.scale*tr.Y + shapeList.MinY*reverseY*n.scale
		if reverseY < 0 {
			y -= n.scale * float32(shapeList.MaxY-shapeList.MinY)
		}

		color := rl.White
		if tr.Highlight {
			for _, s := range shapeList.Shapes {
				renderShape(s, tr.Highlight, rec.X+x-n.scale*shapeList.MinX, rec.Y+y+n.scale*shapeList.MaxY, n.scale, reverseY*n.scale)
			}
		} else {
			rl.DrawTexturePro(shapeList.Texture.Texture,
				rl.NewRectangle(2, 2, float32(shapeList.Texture.Texture.Width)-4, reverseY*float32(shapeList.Texture.Texture.Height-4)),

				rl.NewRectangle(rec.X+x, rec.Y+y, n.scale*(shapeList.MaxX-shapeList.MinX), n.scale*(shapeList.MaxY-shapeList.MinY)),
				rl.Vector2Zero(), 0, color)
		}
	}
	rl.EndTextureMode()
	n.rendered = true
}

func (s *DrawNodes) drawOnTexture(n *Node, pos *Position) {
	rec := s.nodeTextures.NodeTextureRec(n.idx)
	texture := s.nodeTextures.At(n.idx).Texture
	rec.Y = float32(texture.Height) - rec.Y - rec.Height // texture is upside down...
	rec.Height = -rec.Height
	rl.DrawTextureRec(texture, rec, rl.NewVector2(float32(pos.X), float32(pos.Y)), rl.White)
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
		color, err := StringToRGBA(s.Color)
		if err != nil {
			log.Fatal().Str("color", s.Color).Msg("unknown color")
		}
		col = HighlightableShapeColor{
			normal:      ShapeColor{border: color, fill: color},
			highlighted: ShapeColor{border: color, fill: color},
		}
		shapeColors[s.Color] = col
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
