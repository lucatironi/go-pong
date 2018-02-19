package main

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	windowWidth  = 800
	windowHeight = 600
)

var (
	game                 *Game
	deltaTime, lastFrame float64
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func main() {
	window := initGlfw()
	defer glfw.Terminate()

	initOpenGL()

	// OpenGL configuration
	gl.Viewport(0, 0, windowWidth, windowHeight)
	gl.Enable(gl.CW)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	game = newGame(windowWidth, windowHeight)
	game.Init()

	for !window.ShouldClose() {
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame
		glfw.PollEvents()

		// Manage user input
		game.ProcessInput(deltaTime)
		// Update Game state
		game.Update(deltaTime)

		// Render
		gl.ClearColor(0.2, 0.2, 0.2, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		game.Draw()

		window.SwapBuffers()
	}
}

// KeyCallback defines the callback to handle keyboard events
func KeyCallback(window *glfw.Window, key glfw.Key, scanCode int, action glfw.Action, modifierKey glfw.ModifierKey) {
	// When a user presses the escape key, we set the WindowShouldClose property to true, closing the application
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
	if key >= 0 && key < 1024 {
		if action == glfw.Press {
			game.keys[key] = true
		} else if action == glfw.Release {
			game.keys[key] = false
		}
	}
}

// initGlfw initializes glfw and returns a glfw.Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Pong", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	window.SetKeyCallback(KeyCallback)

	return window
}

// initOpenGL initializes OpenGL.
func initOpenGL() {
	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
}
