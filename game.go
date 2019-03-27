// Package game implements the core engine logic.
package game

import (
	"fmt"
	"time"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/jemgunay/game/file"
	"github.com/jemgunay/game/player"
	"github.com/jemgunay/game/server"
	"github.com/jemgunay/game/world"
)

var (
	cfg = pixelgl.WindowConfig{
		Title:  "Test Game",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}

	players map[string]*player.Player
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

	// generate world
	tileGrid := world.NewTileGrid()
	if err := tileGrid.GenerateChunk(); err != nil {
		fmt.Printf("failed to generate world: %s\n", err)
		return
	}

	// start server
	go func() {
		if err := server.Start(9000); err != nil {
			fmt.Printf("server shut down: %s\n", err)
		}
	}()

	// create player
	mainPlayer, err := player.New("jemgunay")
	if err != nil {
		fmt.Printf("failed to create mainPlayer: %s\n", err)
		return
	}

	var (
		camScale = 1.0
		last     = time.Now()
	)
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		// window camera
		cam := pixel.IM.Scaled(mainPlayer.Pos(), camScale).Moved(win.Bounds().Center().Sub(mainPlayer.Pos()))
		win.SetMatrix(cam)

		// handle keyboard input
		if win.Pressed(pixelgl.KeyW) {
			mainPlayer.Up(dt)
		}
		if win.Pressed(pixelgl.KeyS) {
			mainPlayer.Down(dt)
		}
		if win.Pressed(pixelgl.KeyA) {
			mainPlayer.Left(dt)
		}
		if win.Pressed(pixelgl.KeyD) {
			mainPlayer.Right(dt)
		}
		if win.Pressed(pixelgl.KeyR) {
			camScale += 0.01
		}
		if win.Pressed(pixelgl.KeyF) {
			camScale -= 0.01
		}
		if win.Pressed(pixelgl.KeyEscape) {
			win.SetClosed(true)
		}
		// handle mouse movement
		if win.MousePosition() != win.MousePreviousPosition() {
			// point mainPlayer at mouse
			mousePos := cam.Unproject(win.MousePosition())
			mainPlayer.PointTo(mousePos)
		}

		win.Clear(colornames.Greenyellow)

		// draw tiles
		tileGrid.Draw(win)
		// draw player
		mainPlayer.Draw(win)

		win.Update()
	}
}
