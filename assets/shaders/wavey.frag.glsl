#version 330 core

in vec2  vTexCoords;

out vec4 fragColor;

uniform sampler2D uTexture;
uniform vec4 uTexBounds;
// custom uniforms
uniform float uSpeed;
uniform float uTime;

void main() {
    vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
    vec3 influence = texture(uTexture, t).rgb;
    t.y += cos(t.x * 40.0 + (uTime * uSpeed))*0.005;
    t.x += cos(t.y * 40.0 + (uTime * uSpeed))*0.01;
    vec3 col = texture(uTexture, t).rgb;
    fragColor = vec4(col,1.0);
}