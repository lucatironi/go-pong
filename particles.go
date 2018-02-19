package main

import (
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
)

var lastUsedParticle = 0

// Particle handles a particle with a position, velocity, color and life
type Particle struct {
	position mgl.Vec2
	velocity mgl.Vec2
	color    mgl.Vec4
	life     float64
}

func newParticle(position, velocity mgl.Vec2, color mgl.Vec4, life float64) *Particle {
	return &Particle{
		position: position,
		velocity: velocity,
		color:    color,
		life:     life,
	}
}

// ParticleGenerator handles the generation and life cycle of particles
type ParticleGenerator struct {
	particles []*Particle
	amount    int
	shader    *Shader
	quadVao   uint32
}

func newParticleGenerator(shader *Shader, amount int) *ParticleGenerator {
	generator := &ParticleGenerator{
		amount: amount,
		shader: shader,
	}
	generator.Init()

	return generator
}

// Init initializes the generator
func (pg *ParticleGenerator) Init() {
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

	gl.GenVertexArrays(1, &pg.quadVao)
	gl.GenBuffers(1, &vertexBuffer)
	gl.BindVertexArray(pg.quadVao)
	// Fill mesh buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)
	// Set mesh attributes
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	// Create pg.amount default particle instances
	for i := 0; i < pg.amount; i++ {
		pg.particles = append(pg.particles, newParticle(mgl.Vec2{0, 0}, mgl.Vec2{0, 0}, mgl.Vec4{1, 1, 1, 1}, 0.0))
	}
}

// Update updates the particles managed by the generator
func (pg *ParticleGenerator) Update(deltaTime float64, object *GameObject, newParticles int, offset mgl.Vec2) {
	// Add new particles
	for i := 0; i < newParticles; i++ {
		unusedParticle := pg.firstUnusedParticle()
		pg.respawnParticle(pg.particles[unusedParticle], object, offset)
	}
	// Update all particles
	for i := 0; i < pg.amount; i++ {
		p := pg.particles[i]
		p.life -= deltaTime // reduce life
		if p.life > 0.0 {   // particle is alive, thus update
			p.position = p.position.Sub(p.velocity.Mul(float32(deltaTime)))
			p.color[3] -= float32(deltaTime) * 2.5
		}
	}
}

// Draw draws the particles managed by the generator
func (pg *ParticleGenerator) Draw() {
	// Use additive blending to give it a 'glow' effect
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	pg.shader.Use()
	for _, particle := range pg.particles {
		if particle.life > 0.0 {
			pg.shader.SetVector2v("offset", particle.position, false)
			pg.shader.SetVector4v("color", particle.color, false)
			gl.BindVertexArray(pg.quadVao)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
			gl.BindVertexArray(0)
		}
	}
	// Don't forget to reset to default blending mode
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (pg *ParticleGenerator) firstUnusedParticle() int {
	// First search from last used particle, this will usually return almost instantly
	for i := lastUsedParticle; i < pg.amount; i++ {
		if pg.particles[i].life <= 0.0 {
			lastUsedParticle = i
			return i
		}
	}
	// Otherwise, do a linear search
	for i := 0; i < lastUsedParticle; i++ {
		if pg.particles[i].life <= 0.0 {
			lastUsedParticle = i
			return i
		}
	}
	// All particles are taken, override the first one (note that if it repeatedly hits this case, more particles should be reserved)
	lastUsedParticle = 0

	return 0
}

func (pg *ParticleGenerator) respawnParticle(particle *Particle, object *GameObject, offset mgl.Vec2) {
	random := float32(rand.Int31n(50)) / 100.0 / 10.0
	randomColor := float32(rand.Int31n(50)) / 100.0
	particle.position = object.position.Add(mgl.Vec2{random, random}).Add(offset)
	particle.color = mgl.Vec4{randomColor, randomColor, randomColor, 1.0}
	particle.life = 1.0
	particle.velocity = object.velocity.Mul(0.1)
}
