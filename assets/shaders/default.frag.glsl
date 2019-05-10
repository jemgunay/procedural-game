#version 330 core

in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;
uniform sampler2D uTexture;

void main() {
    if (vIntensity == 0) {
        fragColor = uColorMask * vColor;
    } else {
        fragColor = vec4(0, 0, 0, 0);
        fragColor += (1 - vIntensity) * vColor;
        vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
        fragColor += vIntensity * vColor * texture(uTexture, t);
        fragColor *= uColorMask;
    }
}