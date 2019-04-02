// Package game implements the core engine logic.
package game

import (
	"fmt"

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

	// start the scene
	scene.Start()
}

// StartServerOnly is the server only instance entry point.
func StartServerOnly() {}
