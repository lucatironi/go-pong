#version 330 core
layout (location = 0) in vec2 vertex; // <vec2 position>

out vec4 ParticleColor;

uniform mat4 projection;
uniform vec2 offset;
uniform vec4 color;

void main()
{
    float scale = 10.0f;
    ParticleColor = color;
    gl_Position = projection * vec4((vertex.xy * scale) + offset, 1.0, 1.0);
}