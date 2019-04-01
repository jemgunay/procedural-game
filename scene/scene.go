// Package scene manages the scene and executes different scene layers.
package scene

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/player"
	"github.com/jemgunay/game/server"
	"github.com/jemgunay/game/world"
	"golang.org/x/image/colornames"
)

// Layer is a drawable and updatable scene layer.
type Layer interface {
	Update(dt float64)
	Draw()
}

var (
	layerStack []Layer
	// keep an internal reference to the window
	win *pixelgl.Window
	// used to indicate a layer pop to a layer push caller
	popChan = make(chan struct{})
)

// Start initialises and starts up the scene.
func Start() {
	// create window config
	cfg := pixelgl.WindowConfig{
		Title:     "Test Game",
		Bounds:    pixel.R(0, 0, 1024, 768),
		VSync:     true,
		Resizable: true,
	}

	// create window
	var err error
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		fmt.Printf("failed create new window: %s\n", err)
		return
	}
	//win.SetSmooth(true)

	// push a new game layer to the scene
	Push(NewMainMenu())

	// main game loop
	prevTimestamp := time.Now()
	for !win.Closed() {
		dt := time.Since(prevTimestamp).Seconds()
		prevTimestamp = time.Now()

		for _, layer := range layerStack {
			layer.Update(dt)
		}

		for _, layer := range layerStack {
			layer.Draw()
		}
		win.Update()
	}
}

// Push pushes a new layer to the layer stack (above the previous layer).
func Push(layer Layer) {
	layerStack = append(layerStack, layer)
}

// Pop pops the most recently added layer from the layer stack.
func Pop() {
	if len(layerStack) > 0 {
		layerStack = layerStack[:len(layerStack)-1]
	}
	popChan <- struct{}{}
}

// WaitForPop is used to block a layer's update loop until it's child layer has been popped.
func WaitForPop() {
	<-popChan
}

// Count returns the number of layers in the layer stack.
func Count() int {
	return len(layerStack)
}

// MainMenu is the main menu layer which is first displayed upon game startup.
type MainMenu struct {
	uiContainer *UIContainer
	createBtn   *Button
	joinBtn     *Button
	settingsBtn *Button
}

// NewMainMenu creates and initialises a new main menu layer.
func NewMainMenu() *MainMenu {
	menu := &MainMenu{
		createBtn:   NewButton("Create Game", colornames.Paleturquoise, colornames.White),
		joinBtn:     NewButton("Join Game", colornames.Palegreen, colornames.White),
		settingsBtn: NewButton("Settings", colornames.Palevioletred, colornames.White),
	}

	// create container sized half the window height
	container := NewUIContainer(NewPadding(5), func() pixel.Rect {
		b := win.Bounds()
		return b.Resized(b.Center(), pixel.V(b.Size().X, b.Size().Y*0.5))
	})
	container.AddButton(menu.createBtn, menu.joinBtn, menu.settingsBtn)
	menu.uiContainer = container

	return menu
}

// Update updates the main menu layer logic.
func (m *MainMenu) Update(dt float64) {
	switch {
	case m.createBtn.Clicked():
		// create a new game layer
		gameLayer, err := NewGame()
		if err != nil {
			fmt.Printf("failed to create game scene: %s\n", err)
			return
		}
		// pop main menu and push game layer
		Pop()
		Push(gameLayer)

	case m.joinBtn.Clicked():
		m.joinBtn.ToggleEnabled()
	case m.settingsBtn.Clicked():
		m.settingsBtn.ToggleEnabled()
	}

	if win.Pressed(pixelgl.KeyEscape) {
		win.SetClosed(true)
	}
}

// Draw draws the main menu layer to the window.
func (m *MainMenu) Draw() {
	win.Clear(colornames.Whitesmoke)
	m.uiContainer.Draw()
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

// Update updates the game layer logic.
func (g *Game) Update(dt float64) {
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

// Draw draws the game layer to the window.
func (g *Game) Draw() {
	win.Clear(colornames.Greenyellow)
	// draw tiles
	g.tileGrid.Draw(win)
	// draw players
	g.mainPlayer.Draw(win)
	for _, p := range g.otherPlayers {
		p.Draw(win)
	}
}

// OverlayMenu is the overlay menu layer which is drawn over the main game layer.
type OverlayMenu struct{}

// NewOverlayMenu creates and initialises a new overlay menu layer.
func NewOverlayMenu() *OverlayMenu {
	return &OverlayMenu{}
}

// Update updates the overlay menu layer logic.
func (m *OverlayMenu) Update(dt float64) {}

// Draw draws the overlay menu layer to the window.
func (m *OverlayMenu) Draw() {}
