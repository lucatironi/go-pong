package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
)

// SpriteRenderer renders a gameOject
type SpriteRenderer struct {
	shader  *Shader
	quadVao uint32
}

func newSpriteRenderer(shader *Shader) *SpriteRenderer {
	renderer := SpriteRenderer{
		shader: shader,
	}
	renderer.initRenderData()

	return &renderer
}

func (r *SpriteRenderer) initRenderData() {
	// Configure VAO/VBO
	var vertexBuffer uint32
	vertices := []float32{
		0.0, 1.0,
		1.0, 0.0,
		0.0, 0.0,

		0.0, 1.0,
		1.0, 1.0,
		1.0, 0.0,
	}

	gl.GenVertexArrays(1, &r.quadVao)
	gl.GenBuffers(1, &vertexBuffer)
	gl.BindVertexArray(r.quadVao)
	// Fill mesh buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)
	// Set mesh attributes
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

// Draw draws a gameObject
func (r *SpriteRenderer) Draw(position, size mgl.Vec2, rotation float32, color mgl.Vec3) {
	// Prepare transformations
	var model mgl.Mat4
	tMat := mgl.Translate2D(position.X(), position.Y())
	rMat := mgl.HomogRotate2D(rotation)
	sMat := mgl.Scale2D(size.X(), size.Y())

	model = tMat.Mul3(rMat.Mul3(sMat)).Mat4()

	r.shader.Use()
	r.shader.SetMatrix4("model", model, false)
	r.shader.SetVector3v("spriteColor", color, false)

	gl.BindVertexArray(r.quadVao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}
