// Package game implements the core engine logic.
package game

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/jemgunay/game/file"
	"github.com/jemgunay/game/scene"
)

// StartClient is the main client entry point.
func StartClient() {
	// load assets
	if err := file.LoadAllAssets(); err != nil {
		fmt.Printf("failed to process assets: %s\n", err)
		return
	}

	cfg := pixelgl.WindowConfig{
		Title:     "Test Game",
		Bounds:    pixel.R(0, 0, 1024, 768),
		VSync:     true,
		Resizable: true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		fmt.Printf("failed create new window: %s\n", err)
		return
	}
	//win.SetSmooth(true)

	// start the scene
	scene.Start(win)
}

// StartServerOnly is the server only instance entry point.
func StartServerOnly() {}
