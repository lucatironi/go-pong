#version 330 core
layout (location = 0) in vec2 vertex; // <vec2 position>

uniform mat4 model;
uniform mat4 projection;

void main()
{
    gl_Position = projection * model * vec4(vertex.xy, 1.0, 1.0);
}