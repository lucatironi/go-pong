package main

import (
	"bufio"
	"log"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// ResourceManager hosts several functions to load Textures and Shaders
type ResourceManager struct {
	shaders map[string]Shader
}

func newResourceManager() *ResourceManager {
	return &ResourceManager{
		shaders: make(map[string]Shader),
	}
}

// LoadShader loads (and generates) a shader program from file loading vertex, fragment (and geometry) shader's source code. If gShaderFile is not nullptr, it also loads a geometry shader
func (r *ResourceManager) LoadShader(vertexShaderFile, fragmentShaderFile, name string) Shader {
	r.shaders[name] = r.loadShaderFromFile(vertexShaderFile, fragmentShaderFile)
	return r.shaders[name]
}

// GetShader retrieves a stored shader
func (r *ResourceManager) GetShader(name string) *Shader {
	shader := r.shaders[name]
	return &shader
}

// Clear (Properly) delete all shaders
func (r *ResourceManager) Clear() {
	for _, shader := range r.shaders {
		gl.DeleteProgram(shader.ID)
	}
}

func (r *ResourceManager) loadShaderFromFile(vertexShaderFile, fragmentShaderFile string) Shader {
	shader := Shader{}
	shader.Compile(readShaderFile(vertexShaderFile), readShaderFile(fragmentShaderFile))
	return shader
}

func readShaderFile(filePath string) string {
	src := ""
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		src += "\n" + scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	src += "\x00"

	return src
}
