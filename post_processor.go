package main

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Texture2D is able to store and configure a texture in OpenGL.
// It also hosts utility functions for easy management.
type Texture2D struct {
	// Holds the ID of the texture object, used for all
	// texture operations to reference to this particlar texture
	ID uint32
	// Texture image dimensions
	width, height int32 // Width and height of loaded image in pixels
	// Texture Format
	internalFormat int32  // Format of texture object
	imageFormat    uint32 // Format of loaded image
	// Texture configuration
	wrapS     int32 // Wrapping mode on S axis
	wrapT     int32 // Wrapping mode on T axis
	filterMin int32 // Filtering mode if texture pixels < screen pixels
	filterMax int32 // Filtering mode if texture pixels > screen pixels
}

func newTexture2D() *Texture2D {
	texture := Texture2D{
		internalFormat: gl.RGB,
		imageFormat:    gl.RGB,
		wrapS:          gl.REPEAT,
		wrapT:          gl.REPEAT,
		filterMin:      gl.LINEAR,
		filterMax:      gl.LINEAR,
	}
	gl.GenTextures(1, &texture.ID)

	return &texture
}

// Generate generates texture from image data
func (t *Texture2D) Generate(width, height int32, data []byte) {
	t.width = width
	t.height = height
	// Create Texture
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
	if data != nil {
		gl.TexImage2D(gl.TEXTURE_2D, 0, t.internalFormat, width, height, 0, t.imageFormat, gl.UNSIGNED_BYTE, gl.Ptr(&data[0]))
	} else {
		gl.TexImage2D(gl.TEXTURE_2D, 0, t.internalFormat, width, height, 0, t.imageFormat, gl.UNSIGNED_BYTE, nil)
	}
	// Set Texture wrap and filter modes
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, t.wrapS)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, t.wrapT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, t.filterMin)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, t.filterMax)
	// Unbind texture
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// Bind binds the texture as the current active GL_TEXTURE_2D texture object
func (t *Texture2D) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}

// PostProcessor hosts all PostProcessing effects for the game.
// It renders the game on a textured quad after which one can
// enable specific effects by enabling either the confuse, chaos or
// shake boolean.
// It is required to call BeginRender() before rendering the game
// and EndRender() after rendering the game for the class to work.
type PostProcessor struct {
	shader                     *Shader
	texture                    *Texture2D
	width, height              int32
	shake, chaos, confuse      bool
	msFrameBuffer, frameBuffer uint32
	rbo                        uint32
	quadVao                    uint32
}

func newPostProcessor(shader *Shader, width, height int32) *PostProcessor {
	postProcessor := PostProcessor{
		shader:  shader,
		width:   width,
		height:  height,
		shake:   false,
		chaos:   false,
		confuse: false}

	postProcessor.texture = newTexture2D()

	// Initialize renderbuffer/framebuffer object
	gl.GenFramebuffers(1, &postProcessor.msFrameBuffer)
	gl.GenFramebuffers(1, &postProcessor.frameBuffer)
	gl.GenRenderbuffers(1, &postProcessor.rbo)

	// Initialize renderbuffer storage with a multisampled color buffer (don't need a depth/stencil buffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, postProcessor.msFrameBuffer)
	gl.BindRenderbuffer(gl.RENDERBUFFER, postProcessor.rbo)
	gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, 8, gl.RGB, postProcessor.width, postProcessor.height) // Allocate storage for render buffer object
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, postProcessor.rbo)     // Attach MS render buffer object to framebuffer
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		fmt.Println("ERROR::POSTPROCESSOR: Failed to initialize MSFBO")
	}

	// Also initialize the FBO/texture to blit multisampled color-buffer to; used for shader operations (for postprocessing effects)
	gl.BindFramebuffer(gl.FRAMEBUFFER, postProcessor.frameBuffer)
	postProcessor.texture.Generate(postProcessor.width, postProcessor.height, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, postProcessor.texture.ID, 0) // Attach texture to framebuffer as its color attachment
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		fmt.Println("ERROR::POSTPROCESSOR: Failed to initialize FBO")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Initialize render data and uniforms
	postProcessor.initRenderData()
	postProcessor.shader.SetInteger("scene", 0, true)
	offset := float32(1.0 / 300.0)
	offsets := [][]float32{
		{-offset, offset},  // top-left
		{0.0, offset},      // top-center
		{offset, offset},   // top-right
		{-offset, 0.0},     // center-left
		{0.0, 0.0},         // center-center
		{offset, 0.0},      // center - right
		{-offset, -offset}, // bottom-left
		{0.0, -offset},     // bottom-center
		{offset, -offset},  // bottom-right
	}
	gl.Uniform2fv(postProcessor.shader.getUniformLocation("offsets"), 9, &offsets[0][0])
	edgeKernel := []int32{
		-1, -1, -1,
		-1, 8, -1,
		-1, -1, -1,
	}
	gl.Uniform1iv(postProcessor.shader.getUniformLocation("edge_kernel"), 9, &edgeKernel[0])
	blurKernel := []float32{
		1.0 / 16, 2.0 / 16, 1.0 / 16,
		2.0 / 16, 4.0 / 16, 2.0 / 16,
		1.0 / 16, 2.0 / 16, 1.0 / 16,
	}
	gl.Uniform1fv(postProcessor.shader.getUniformLocation("blur_kernel"), 9, &blurKernel[0])

	return &postProcessor
}

// BeginRender prepares the postprocessor's framebuffer operations before rendering the game
func (pp *PostProcessor) BeginRender() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, pp.msFrameBuffer)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

// EndRender should be called after rendering the game, so it stores all the rendered data into a texture object
func (pp *PostProcessor) EndRender() {
	// Now resolve multisampled color-buffer into intermediate FBO to store to texture
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, pp.msFrameBuffer)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, pp.frameBuffer)
	gl.BlitFramebuffer(0, 0, int32(pp.width), int32(pp.height), 0, 0, int32(pp.width), int32(pp.height), gl.COLOR_BUFFER_BIT, gl.NEAREST)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0) // Binds both READ and WRITE framebuffer to default framebuffer
}

// Render renders the PostProcessor texture quad (as a screen-encompassing large sprite)
func (pp *PostProcessor) Render(time float32) {
	// Set uniforms/options
	pp.shader.Use()
	pp.shader.SetFloat("time", time, false)
	pp.shader.SetInteger("confuse", boolToInt32(pp.confuse), false)
	pp.shader.SetInteger("chaos", boolToInt32(pp.chaos), false)
	pp.shader.SetInteger("shake", boolToInt32(pp.shake), false)
	// Render textured quad
	gl.ActiveTexture(gl.TEXTURE0)
	pp.texture.Bind()
	gl.BindVertexArray(pp.quadVao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

func (pp *PostProcessor) initRenderData() {
	// Configure VAO/VBO
	var vertexBuffer uint32
	vertices := []float32{
		// Pos      // Tex
		-1.0, -1.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
		-1.0, 1.0, 0.0, 1.0,

		-1.0, -1.0, 0.0, 0.0,
		1.0, -1.0, 1.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
	}

	gl.GenVertexArrays(1, &pp.quadVao)
	gl.GenBuffers(1, &vertexBuffer)
	gl.BindVertexArray(pp.quadVao)
	// Fill mesh buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)
	// Set mesh attributes
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 0, nil)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func boolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
