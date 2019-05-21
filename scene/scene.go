// Package scene manages the scene and executes different scene layers.
package scene

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/jemgunay/procedural-game/file"
	"github.com/jemgunay/procedural-game/scene/world"
)

// Layer is a drawable and updatable scene layer.
type Layer interface {
	Update(dt float64)
	Draw()
}

var (
	layerStack []LayerWrapper
	win        *pixelgl.Window
)

// Start initialises and starts up the scene.
func Start() {
	// create window config
	cfg := pixelgl.WindowConfig{
		Title:     "Procedural Game Demo",
		Bounds:    pixel.R(0, 0, 1024, 768),
		VSync:     false,
		Resizable: true,
	}

	// create window
	var err error
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		fmt.Printf("failed create new window: %s\n", err)
		return
	}

	// create shaders
	world.DefaultShader, err = file.NewDefaultFragShader()
	if err != nil {
		fmt.Printf("failed create new window: %s\n", err)
		return
	}
	world.WavyShader, err = file.NewWavyFragShader(5)
	if err != nil {
		fmt.Printf("failed create new window: %s\n", err)
		return
	}

	// push a main menu layer to the scene
	Push(NewMainMenu())

	// limit update cycles to 120 FPS
	frameRateLimiter := time.Tick(time.Second / 120)
	prevTimestamp := time.Now().UTC()
	// main game loop
	for !win.Closed() {

		dt := time.Since(prevTimestamp).Seconds()
		prevTimestamp = time.Now().UTC()

		for _, layer := range layerStack {
			layer.Update(dt)
		}

		for _, layer := range layerStack {
			layer.Draw()
		}
		win.Update()
		<-frameRateLimiter
	}
}

// LayerResult represents the state returned when a layer is popped from the layer stack.
type LayerResult string

// Layer state constants.
const (
	Default    LayerResult = "default"
	Disconnect LayerResult = "disconnect"
	Quit       LayerResult = "quit"
)

// LayerWrapper associates a LayerResult channel with a Layer.
type LayerWrapper struct {
	Layer
	resultCh chan LayerResult
}

// Push pushes a new layer to the layer stack (above the previous layer).
func Push(layer Layer) chan LayerResult {
	ch := make(chan LayerResult, 1)
	layerStack = append(layerStack, LayerWrapper{
		Layer:    layer,
		resultCh: ch,
	})
	return ch
}

// Pop pops the most recently added layer from the layer stack.
func Pop(result LayerResult) {
	if len(layerStack) > 0 {
		// send result to subscriber
		layerStack[len(layerStack)-1].resultCh <- result
		close(layerStack[len(layerStack)-1].resultCh)
		// pop layer from end of stack
		layerStack = layerStack[:len(layerStack)-1]
	}
}

// Count returns the number of layers in the layer stack.
func Count() int {
	return len(layerStack)
}
