package main

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
)

// GameState represents a state
type GameState int

const (
	gameActive GameState = iota
	gameMenu
	gameWin
)

var (
	paddleSize     = mgl.Vec2{20, 100}
	paddleVelocity = float32(500)

	initialBallVelocity = mgl.Vec2{450.0, 300.0}

	shakeTime = 0.0
)

// Game represents a game uber object
type Game struct {
	state           GameState
	keys            map[glfw.Key]bool
	width, height   int
	renderer        *SpriteRenderer
	resourceManager *ResourceManager
	particles       *ParticleGenerator
	effects         *PostProcessor
	paddle1         *GameObject
	paddle2         *GameObject
	ball            *BallObject
}

func newGame(width, height int) *Game {
	return &Game{
		state:  gameActive,
		keys:   make(map[glfw.Key]bool),
		width:  width,
		height: height,
	}
}

// Init initializes a game
func (g *Game) Init() {
	g.resourceManager = newResourceManager()
	// Load shaders
	g.resourceManager.LoadShader("./shaders/sprite.vs", "./shaders/sprite.frag", "sprite")
	g.resourceManager.LoadShader("./shaders/particle.vs", "./shaders/particle.frag", "particle")
	g.resourceManager.LoadShader("./shaders/post_processing.vs", "./shaders/post_processing.frag", "postprocessing")
	// Configure shaders
	projection := mgl.Ortho2D(0.0, float32(g.width), float32(g.height), 0.0)
	g.resourceManager.GetShader("sprite").Use().SetMatrix4("projection", projection, false)
	g.resourceManager.GetShader("particle").Use().SetMatrix4("projection", projection, false)
	// Set render-specific controls
	g.renderer = newSpriteRenderer(g.resourceManager.GetShader("sprite"))
	g.particles = newParticleGenerator(g.resourceManager.GetShader("particle"), 50)
	g.effects = newPostProcessor(g.resourceManager.GetShader("postprocessing"), int32(g.width), int32(g.height))
	// Configure game objects
	paddle1Position := mgl.Vec2{
		10,
		float32(g.height/2) - paddleSize.Y()/2}
	g.paddle1 = newGameObject(paddle1Position, paddleSize)
	paddle2Position := mgl.Vec2{
		float32(g.width) - paddleSize.X() - 10,
		float32(g.height/2) - paddleSize.Y()/2}
	g.paddle2 = newGameObject(paddle2Position, paddleSize)
	g.ball = newBallObject(mgl.Vec2{float32(g.width / 2), float32(g.height / 2)}, 10.0, initialBallVelocity)
}

// ProcessInput processes the input
func (g *Game) ProcessInput(deltaTime float64) {
	if g.state == gameActive {
		deltaSpace := paddleVelocity * float32(deltaTime)
		// Move paddle one
		if g.keys[glfw.KeyW] {
			if g.paddle1.position.Y() >= 0 {
				g.paddle1.position[1] -= deltaSpace
			}
		}
		if g.keys[glfw.KeyS] {
			if g.paddle1.position.Y() <= float32(g.height)-g.paddle1.size.Y() {
				g.paddle1.position[1] += deltaSpace
			}
		}
		// Move paddle two
		if g.keys[glfw.KeyUp] {
			if g.paddle2.position.Y() >= 0 {
				g.paddle2.position[1] -= deltaSpace
			}
		}
		if g.keys[glfw.KeyDown] {
			if g.paddle2.position.Y() <= float32(g.height)-g.paddle2.size.Y() {
				g.paddle2.position[1] += deltaSpace
			}
		}
	}
}

// Update updates the game
func (g *Game) Update(deltaTime float64) {
	// Update objects
	g.ball.Move(deltaTime, g.width, g.height)
	// Check for collisions
	g.DoCollisions()
	// Update particles
	g.particles.Update(deltaTime, &g.ball.GameObject, 1, mgl.Vec2{g.ball.radius, g.ball.radius})
	// Reduce shake time
	if shakeTime > 0.0 {
		shakeTime -= deltaTime
		if shakeTime <= 0.0 {
			g.effects.shake = false
		}
	}
	// Check loss condition
	if g.ball.position.X() <= 0.0 {
		// paddle2 scored
		g.ball.Reset(mgl.Vec2{float32(g.width / 2), float32(g.height / 2)}, initialBallVelocity.Mul(-1))
	} else if g.ball.position.X()+g.ball.size.X() >= float32(g.width) {
		// paddle1 scored
		g.ball.Reset(mgl.Vec2{float32(g.width / 2), float32(g.height / 2)}, initialBallVelocity)
	}
}

// Draw draws the game
func (g *Game) Draw() {
	if g.state == gameActive {
		// Begin rendering to postprocessing quad
		g.effects.BeginRender()
		// Draw paddles
		g.paddle1.Draw(g.renderer)
		g.paddle2.Draw(g.renderer)
		// Draw particles
		g.particles.Draw()
		// Draw ball
		g.ball.Draw(g.renderer)
		// End rendering to postprocessing quad
		g.effects.EndRender()
		// Render postprocessing quad
		g.effects.Render(float32(glfw.GetTime()))
	}
}

// DoCollisions checks if gameobjects collided
func (g *Game) DoCollisions() {
	if !g.ball.isStuck && (g.ball.CheckCollision(g.paddle1) || g.ball.CheckCollision(g.paddle2)) {
		shakeTime = 0.1
		g.effects.shake = true
		g.ball.velocity[0] = -g.ball.velocity.X()
	}
}
