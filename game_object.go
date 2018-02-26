package main

import mgl "github.com/go-gl/mathgl/mgl32"

// GameObject holds the structure of a object in the game with a position and a size
type GameObject struct {
	position mgl.Vec2
	size     mgl.Vec2
	velocity mgl.Vec2
	color    mgl.Vec3
	rotation float32
}

func newGameObject(position, size mgl.Vec2) *GameObject {
	return &GameObject{
		position: position,
		size:     size,
		velocity: mgl.Vec2{0, 0},
		rotation: 0,
		color:    mgl.Vec3{1, 1, 1}}
}

// Draw renders a GameObject using the provided renderer
func (o *GameObject) Draw(renderer *SpriteRenderer) {
	renderer.Draw(o.position, o.size, o.rotation, o.color)
}

// Reset resets a GameObject
func (o *GameObject) Reset(position mgl.Vec2) {
	o.position = position
}

// CheckCollision checks collisions between two game objects using o - AABB
func (o *GameObject) CheckCollision(other *GameObject) bool {
	// Collision x-axis?
	collisionX := o.position.X()+o.size.X() >= other.position.X() &&
		other.position.X()+other.size.X() >= o.position.X()
	// Collision y-axis?
	collisionY := o.position.Y()+o.size.Y() >= other.position.Y() &&
		other.position.Y()+other.size.Y() >= o.position.Y()
	// Collision only if on both axes
	return collisionX && collisionY
}

// BallObject is a special game object to handle the ball
type BallObject struct {
	GameObject
	isStuck bool
	radius  float32
}

func newBallObject(position mgl.Vec2, radius float32, velocity mgl.Vec2) *BallObject {
	return &BallObject{
		isStuck: false,
		radius:  radius,
		GameObject: GameObject{
			position: position,
			size:     mgl.Vec2{radius * 2, radius * 2},
			velocity: velocity,
			rotation: 0,
			color:    mgl.Vec3{1, 1, 1}}}
}

// Move moves the ball
func (b *BallObject) Move(deltaTime float64, windowWidth, windowHeight int) mgl.Vec2 {
	// If not stuck to player board
	if !b.isStuck {
		// Move the ball
		b.position = b.position.Add(b.velocity.Mul(float32(deltaTime)))
		// Check if outside window bounds; if so, reverse velocity and restore at correct position
		if b.position.Y() <= 0.0 {
			b.velocity[1] = -b.velocity.Y()
			b.position[1] = 0.0
		} else if b.position.Y()+b.size.Y() >= float32(windowHeight) {
			b.velocity[1] = -b.velocity.Y()
			b.position[1] = float32(windowHeight) - b.size.Y()
		}
	}

	return b.position
}

// Reset resets the ball
func (b *BallObject) Reset(position, velocity mgl.Vec2) {
	b.position = position
	b.velocity = velocity
}
