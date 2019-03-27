// Package game implements the core engine logic.
package game

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/file"
	"github.com/jemgunay/game/scene"
)

var (
	cfg = pixelgl.WindowConfig{
		Title:  "Test Game",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
)

// Run is the client entry point.
func Run() {
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		fmt.Printf("failed create new window: %s\n", err)
		return
	}
	//win.SetSmooth(true)

	// load assets
	if err := file.LoadAllAssets(); err != nil {
		fmt.Printf("failed to process assets: %s\n", err)
		return
	}

	// push a new game layer to the scene
	scene.Push(scene.NewMainMenu())

	// main game loop
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		scene.Step(win, dt)
	}

}
