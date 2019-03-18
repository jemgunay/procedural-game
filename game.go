// Package game implements the core engine logic.
package game

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/file"
	"github.com/jemgunay/game/world"
	"golang.org/x/image/colornames"
)

// Run is the client entry point.
func Run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Test Game",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		fmt.Printf("failed create new window: %s", err)
		return
	}
	win.SetSmooth(true)

	// load assets
	if err := file.LoadAllAssets(); err != nil {
		fmt.Printf("failed to process assets: %s", err)
		return
	}

	// generate world
	tileGrid := world.NewTileGrid()
	if err := tileGrid.GenerateChunk(); err != nil {
		fmt.Printf("failed to generate world: %s", err)
		return
	}

	// main loop
	var (
		camPos   = pixel.ZV
		camSpeed = 1000.0
		last     = time.Now()
	)
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		// window camera
		cam := pixel.IM.Scaled(camPos, 0.2).Moved(camPos.Scaled(-1.0))
		win.SetMatrix(cam)

		// handle keyboard input
		if win.Pressed(pixelgl.KeyLeft) {
			camPos.X -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyRight) {
			camPos.X += camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyDown) {
			camPos.Y -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyUp) {
			camPos.Y += camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyEscape) {
			win.SetClosed(true)
		}
		win.Clear(colornames.Greenyellow)

		// draw tiles
		tileGrid.Draw(win)

		win.Update()
	}
}
