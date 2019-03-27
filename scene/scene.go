package scene

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"github.com/jemgunay/game/player"
	"github.com/jemgunay/game/server"
	"github.com/jemgunay/game/world"
)

// Layer is a drawable and updatable scene layer.
type Layer interface {
	Update(win *pixelgl.Window, dt float64)
	Draw(win *pixelgl.Window)
}

var layerStack []Layer

// Push pushes a new layer to the layer stack (above the previous layer).
func Push(layer Layer) {
	layerStack = append(layerStack, layer)
}

// Pop pops the most recently added layer from the layer stack.
func Pop() {
	if len(layerStack) > 0 {
		layerStack = layerStack[:len(layerStack)-1]
	}
}

// Update calls the Update and Draw function for each of the layers in the layer stack.
func Step(win *pixelgl.Window, dt float64) {
	for _, layer := range layerStack {
		layer.Update(win, dt)
	}

	win.Clear(colornames.Greenyellow)
	for _, layer := range layerStack {
		layer.Draw(win)
	}
	win.Update()
}

// MainMenu is the main menu layer which is first displayed upon game startup.
type MainMenu struct{}

// NewMainMenu creates and initialises a new main menu layer.
func NewMainMenu() *MainMenu {
	return &MainMenu{}
}

// Update updates the main menu logic.
func (m *MainMenu) Update(win *pixelgl.Window, dt float64) {
	if win.Pressed(pixelgl.KeyEnter) {
		// kill main menu and launch game layer
		Pop()
		// push a new game layer to the scene
		gameLayer, err := NewGame()
		if err != nil {
			fmt.Printf("failed to create game scene: %s\n", err)
			// restore main menu
			Push(m)
			return
		}
		Push(gameLayer)
	}
}

// Draw draws the main menu to the window.
func (m *MainMenu) Draw(win *pixelgl.Window) {

}

// Game is the main interactive game functionality layer.
type Game struct {
	tileGrid     *world.TileGrid
	mainPlayer   *player.Player
	otherPlayers map[string]*player.Player

	camScale float64
}

// NewGame creates and initialises a new Game layer.
func NewGame() (*Game, error) {
	// generate world
	tileGrid := world.NewTileGrid()
	if err := tileGrid.GenerateChunk(); err != nil {
		return nil, fmt.Errorf("failed to generate world: %s", err)
	}

	// start server
	go func() {
		if err := server.Start(9000); err != nil {
			fmt.Printf("server shut down: %s\n", err)
		}
	}()

	// create main player
	mainPlayer, err := player.New("jemgunay")
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %s", err)
	}

	return &Game{
		tileGrid:     tileGrid,
		mainPlayer:   mainPlayer,
		otherPlayers: make(map[string]*player.Player),
		camScale:     1.0,
	}, nil
}

func (g *Game) Update(win *pixelgl.Window, dt float64) {
	// window camera
	cam := pixel.IM.Scaled(g.mainPlayer.Pos, g.camScale).Moved(win.Bounds().Center().Sub(g.mainPlayer.Pos))
	win.SetMatrix(cam)

	// handle keyboard input
	if win.Pressed(pixelgl.KeyW) {
		g.mainPlayer.Up(dt)
	}
	if win.Pressed(pixelgl.KeyS) {
		g.mainPlayer.Down(dt)
	}
	if win.Pressed(pixelgl.KeyA) {
		g.mainPlayer.Left(dt)
	}
	if win.Pressed(pixelgl.KeyD) {
		g.mainPlayer.Right(dt)
	}
	if win.Pressed(pixelgl.KeyR) {
		if g.camScale < 2 {
			g.camScale += 0.01
		}
	}
	if win.Pressed(pixelgl.KeyF) {
		if g.camScale > 0 {
			g.camScale -= 0.01
		}
	}
	if win.Pressed(pixelgl.KeyEscape) {
		win.SetClosed(true)
	}
	// handle mouse movement
	if win.MousePosition() != win.MousePreviousPosition() {
		// point mainPlayer at mouse
		mousePos := cam.Unproject(win.MousePosition())
		g.mainPlayer.PointTo(mousePos)
	}
}

func (g *Game) Draw(win *pixelgl.Window) {
	// draw tiles
	g.tileGrid.Draw(win)
	// draw players
	g.mainPlayer.Draw(win)
	for _, p := range g.otherPlayers {
		p.Draw(win)
	}
}

type OverlayMenu struct {
}

func NewOverlayMenu() *OverlayMenu {
	return &OverlayMenu{}
}

func (m *OverlayMenu) Update(win *pixelgl.Window, dt float64) {

}

func (m *OverlayMenu) Draw(win *pixelgl.Window) {

}
