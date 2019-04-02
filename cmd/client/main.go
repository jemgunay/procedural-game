// Package main is the entry point for the game client.
package main

import (
	"github.com/faiface/pixel/pixelgl"

	"github.com/jemgunay/game"
)

func main() {
	pixelgl.Run(game.StartClient)
}
