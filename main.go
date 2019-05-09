// Package main is the game entry point.
package main

import (
	"fmt"

	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/file"
	"github.com/jemgunay/game/scene"
)

func main() {
	pixelgl.Run(func() {
		// load assets
		if err := file.LoadAllAssets(); err != nil {
			fmt.Printf("failed to process assets: %s\n", err)
			return
		}

		// start the scene
		scene.Start()
	})
}
