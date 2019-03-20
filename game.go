// Package game implements the core engine logic.
package game

import (
	"fmt"
	"math"
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
	//win.SetSmooth(true)

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

	// create player
	player, err := file.CreateSprite(file.Player)
	if err != nil {
		fmt.Printf("failed to create player: %s", err)
		return
	}
	var (
		playerPos         = pixel.ZV
		playerSpeed       = 500.0
		playerOrientation = 0.0
	)

	// main loop
	var (
		//camPos   = pixel.ZV
		//camSpeed = 1000.0
		camScale = 1.0
		last     = time.Now()
	)
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		// window camera
		cam := pixel.IM.Scaled(playerPos, camScale).Moved(win.Bounds().Center().Sub(playerPos))
		win.SetMatrix(cam)

		// handle keyboard input
		if win.Pressed(pixelgl.KeyA) {
			playerPos.X -= playerSpeed * dt
		}
		if win.Pressed(pixelgl.KeyD) {
			playerPos.X += playerSpeed * dt
		}
		if win.Pressed(pixelgl.KeyS) {
			playerPos.Y -= playerSpeed * dt
		}
		if win.Pressed(pixelgl.KeyW) {
			playerPos.Y += playerSpeed * dt
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
			mouse := cam.Unproject(win.MousePosition())
			// point player at mouse
			playerOrientation = math.Atan2(mouse.Y - playerPos.Y, mouse.X - playerPos.X)
		}

		win.Clear(colornames.Greenyellow)

		// draw tiles
		tileGrid.Draw(win)
		player.Draw(win, pixel.IM.Moved(playerPos).Rotated(playerPos, playerOrientation))

		win.Update()
	}
}
