// Package game implements the core engine logic.
package game

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/file"
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

	if err := file.LoadAllAssets(); err != nil {
		fmt.Printf("failed to process assets: %s", err)
		return
	}

	sprite, err := file.CreateSprite("road_nesw.png")
	if err != nil {
		fmt.Printf("failed to create sprite: %s", err)
		return
	}

	angle := 0.0

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		win.Clear(colornames.Greenyellow)

		angle += 3 * dt

		mat := pixel.IM.Rotated(pixel.ZV, angle).Moved(win.Bounds().Center())
		sprite.Draw(win, mat)

		win.Update()
	}
}
