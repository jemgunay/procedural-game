// Package game implements the core engine logic.
package game

import (
	"fmt"

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

	if err := file.LoadAllAssets(); err != nil {
		fmt.Printf("failed to process assets: %s", err)
		return
	}

	sprite, err := file.ImageToSprite("road_nesw.png")
	if err != nil {

	}

	win.Clear(colornames.Greenyellow)

	sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

	for !win.Closed() {
		win.Update()
	}
}