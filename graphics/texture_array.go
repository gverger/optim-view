package graphics

import rl "github.com/gen2brain/raylib-go/raylib"

type TextureArray struct {
	textureSize         int
	nodesPerTextureLine int

	Textures []rl.RenderTexture2D
}

const (
	MaxTextureSize = 8192
)

func NewTextureArray(textures int, textureSize int) *TextureArray {
	array := TextureArray{
		textureSize:         textureSize,
		nodesPerTextureLine: MaxTextureSize / textureSize,
	}

	array.Textures = make([]rl.RenderTexture2D, 0)
	nbTextureLines := (textures-1)/array.nodesPerTextureLine + 1
	nbTextures := (nbTextureLines-1)/array.nodesPerTextureLine + 1
	for i := 0; i < nbTextures; i++ {
		array.Textures = append(array.Textures,
			rl.LoadRenderTexture(
				int32(textureSize*array.nodesPerTextureLine),
				int32(min(array.nodesPerTextureLine, nbTextureLines)*textureSize)),
		)
		rl.BeginTextureMode(array.Textures[i])
		rl.ClearBackground(rl.Fade(rl.White, 0))
		rl.EndTextureMode()
		nbTextureLines -= array.nodesPerTextureLine
	}

	return &array
}

func (array TextureArray) nodeTextureIdx(node int) int {
	return (node - 1) / (array.nodesPerTextureLine * array.nodesPerTextureLine)
}

func (array TextureArray) At(idx int) rl.RenderTexture2D {
	return array.Textures[array.nodeTextureIdx(idx)]
}

func (array TextureArray) NodeTextureRec(node int) rl.Rectangle {
	n := (node - 1) % (array.nodesPerTextureLine * array.nodesPerTextureLine)
	x := n % array.nodesPerTextureLine
	y := n / array.nodesPerTextureLine
	return rl.NewRectangle(
		float32(x*array.textureSize),
		float32(y*array.textureSize),
		float32(array.textureSize),
		float32(array.textureSize),
	)
}

func (array TextureArray) Unload() {
	for _, t := range array.Textures {
		rl.UnloadRenderTexture(t)
	}
}
