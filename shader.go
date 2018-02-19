package main

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
)

// Shader represents a shader
type Shader struct {
	ID uint32
}

// Use sets the current shader as active
func (s *Shader) Use() *Shader {
	gl.UseProgram(s.ID)
	return s
}

// Compile compiles the shader from given source code
func (s *Shader) Compile(vertexSource, fragmentSource string) {
	vertexShader, err := compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	s.ID = gl.CreateProgram()
	gl.AttachShader(s.ID, vertexShader)
	gl.AttachShader(s.ID, fragmentShader)
	gl.LinkProgram(s.ID)

	// Delete the shaders as they're linked into our program now and no longer necessery
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
}

// SetFloat utility function to pass a float to a shader
func (s *Shader) SetFloat(name string, value float32, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform1f(s.getUniformLocation(name), value)
}

// SetInteger utility function to pass an integer to a shader
func (s *Shader) SetInteger(name string, value int32, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform1i(s.getUniformLocation(name), value)
}

// SetVector2f utility function to pass a vec2 to a shader
func (s *Shader) SetVector2f(name string, x, y float32, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform2f(s.getUniformLocation(name), x, y)
}

// SetVector2v utility function to pass a vec2 to a shader
func (s *Shader) SetVector2v(name string, value mgl.Vec2, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform2f(s.getUniformLocation(name), value.X(), value.Y())
}

// SetVector3f utility function to pass a vec3 to a shader
func (s *Shader) SetVector3f(name string, x, y, z float32, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform3f(s.getUniformLocation(name), x, y, z)
}

// SetVector3v utility function to pass a vec3 to a shader
func (s *Shader) SetVector3v(name string, value mgl.Vec3, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform3f(s.getUniformLocation(name), value.X(), value.Y(), value.Z())
}

// SetVector4f utility function to pass a vec4 to a shader
func (s *Shader) SetVector4f(name string, x, y, z, w float32, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform4f(s.getUniformLocation(name), x, y, z, w)
}

// SetVector4v utility function to pass a vec4 to a shader
func (s *Shader) SetVector4v(name string, value mgl.Vec4, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform4f(s.getUniformLocation(name), value.X(), value.Y(), value.Z(), value.W())
}

// SetMatrix4 utility function to pass a mat4 to a shader
func (s *Shader) SetMatrix4(name string, matrix mgl.Mat4, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.UniformMatrix4fv(s.getUniformLocation(name), 1, false, &matrix[0])
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func (s *Shader) getUniformLocation(name string) int32 {
	return gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%v\x00", name)))
}
