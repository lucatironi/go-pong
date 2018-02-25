package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"image"
	"image/draw"

	"github.com/go-gl/gl/v4.1-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Character holds all state information relevant to a character as loaded using FreeType
type Character struct {
	textureID uint32 // ID handle of the glyph texture
	width     int    // glyph width
	height    int    // glyph height
	advance   int    // glyph advance
	bearingH  int    // glyph bearing horizontal
	bearingV  int    // glyph bearing vertical
}

// TextRenderer renders text displayed by a font loaded using the FreeType library.
// A single font is loaded, processed into a list of Character items for later rendering.
type TextRenderer struct {
	chars  []*Character // Holds a list of pre-compiled Characters
	shader *Shader      // Shader used for text rendering
	vao    uint32       // Render state
	vbo    uint32       // Render state
}

func newTextRenderer(shader *Shader) *TextRenderer {
	renderer := TextRenderer{
		shader: shader,
		chars:  make([]*Character, 0, 96),
	}
	renderer.shader.SetInteger("text", 0, false)

	return &renderer
}

func (t *TextRenderer) initRenderData() {
	// Configure VAO/VBO
	gl.GenVertexArrays(1, &t.vao)
	gl.GenBuffers(1, &t.vbo)
	gl.BindVertexArray(t.vao)
	// Fill mesh buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 6*4*4, nil, gl.DYNAMIC_DRAW)
	// Set mesh attributes
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

// LoadFont pre-compiles a list of characters from the given font
func (t *TextRenderer) LoadFont(fontFile string, fontSize float64) {
	fd, err := os.Open(fontFile)
	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR::TEXTRENDERER: %v", err))
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR::TEXTRENDERER: %v", err))
	}

	// Read the truetype font.
	ttf, err := truetype.Parse(data)
	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR::TEXTRENDERER: %v", err))
	}

	// Make each gylph
	for ch := rune(32); ch <= rune(127); ch++ {
		char := new(Character)

		// Create new face to measure glyph dimensions
		ttfFace := truetype.NewFace(ttf, &truetype.Options{
			Size:    fontSize,
			DPI:     72,
			Hinting: font.HintingFull,
		})

		gBnd, gAdv, ok := ttfFace.GlyphBounds(ch)
		if ok != true {
			fmt.Println(fmt.Sprintf("ERROR::TEXTRENDERER: ttf face glyphBounds error"))
		}

		gh := int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)
		gw := int32((gBnd.Max.X - gBnd.Min.X) >> 6)

		// If gylph has no dimensions set to a max value
		if gw == 0 || gh == 0 {
			gBnd = ttf.Bounds(fixed.Int26_6(fontSize))
			gw = int32((gBnd.Max.X - gBnd.Min.X) >> 6)
			gh = int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)
		}

		// The glyph's ascent and descent equal -bounds.Min.Y and +bounds.Max.Y.
		gAscent := int(-gBnd.Min.Y) >> 6
		gdescent := int(gBnd.Max.Y) >> 6

		// Set w,h and adv, bearing V and bearing H in char
		char.width = int(gw)
		char.height = int(gh)
		char.advance = int(gAdv)
		char.bearingV = gdescent
		char.bearingH = (int(gBnd.Min.X) >> 6)

		// Create image to draw glyph
		fg, bg := image.White, image.Black
		rect := image.Rect(0, 0, int(gw), int(gh))
		rgba := image.NewRGBA(rect)
		draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

		// Create a freetype context for drawing
		c := freetype.NewContext()
		c.SetDPI(72)
		c.SetFont(ttf)
		c.SetFontSize(fontSize)
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)
		c.SetSrc(fg)
		c.SetHinting(font.HintingFull)

		// Set the glyph dot
		px := 0 - (int(gBnd.Min.X) >> 6)
		py := (gAscent)
		pt := freetype.Pt(px, py)

		// Draw the text from mask to image
		_, err = c.DrawString(string(ch), pt)
		if err != nil {
			fmt.Println(fmt.Sprintf("ERROR::TEXTRENDERER: %v", err))
		}

		// Generate texture
		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Dx()), int32(rgba.Rect.Dy()), 0,
			gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

		char.textureID = texture

		// Add char to chars list
		t.chars = append(t.chars, char)
	}

	gl.BindTexture(gl.TEXTURE_2D, 0)

	t.initRenderData()
}

// RenderText renders a string of text using the precompiled list of characters
func (t *TextRenderer) RenderText(x, y, scale float32, color mgl.Vec3, text string, argv ...interface{}) {
	t.shader.Use()
	t.shader.SetVector3v("textColor", color, false)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(t.vao)

	lowChar := rune(32)
	indices := []rune(fmt.Sprintf(text, argv...))

	for i := range indices {
		char := indices[i]
		// Find rune in chars list
		charRune := t.chars[char-lowChar]

		// Calculate position and size for current rune
		xPos := x + float32(charRune.bearingH)*scale
		yPos := y - float32(charRune.height-charRune.bearingV)*scale
		w := float32(charRune.width) * scale
		h := float32(charRune.height) * scale

		// Update VBO for each character
		var vertices = []float32{
			// X, Y, U, V
			xPos, yPos, 0.0, 0.0,
			xPos + w, yPos, 1.0, 0.0,
			xPos, yPos + h, 0.0, 1.0,
			xPos, yPos + h, 0.0, 1.0,
			xPos + w, yPos, 1.0, 0.0,
			xPos + w, yPos + h, 1.0, 1.0}

		// Render glyph texture over quad
		gl.BindTexture(gl.TEXTURE_2D, charRune.textureID)
		// Update content of VBO memory
		gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)
		// Be sure to use glBufferSubData and not glBufferData
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		// Render quad
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		// Now advance cursors for next glyph (note that advance is number of 1/64 pixels)
		x += float32((charRune.advance >> 6)) * scale // Bitshift by 6 to get value in pixels (2^6 = 64 (divide amount of 1/64th pixels by 64 to get amount of pixels))
	}
	// clear opengl textures and programs
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}
