// Package scene manages the scene and executes different scene layers.
package scene

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"github.com/jemgunay/game/player"
	"github.com/jemgunay/game/server"
	"github.com/jemgunay/game/world"
)

// Layer is a drawable and updatable scene layer.
type Layer interface {
	Update(dt float64)
	Draw()
}

var (
	layerStack []LayerWrapper
	win        *pixelgl.Window
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

	// push a main menu layer to the scene
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

// LayerResult represents the state returned when a layer is popped from the layer stack.
type LayerResult string

// Layer state constants.
const (
	Default LayerResult = "default"
	Quit    LayerResult = "quit"
)

// LayerWrapper associates a LayerResult channel with a Layer.
type LayerWrapper struct {
	Layer
	resultCh chan LayerResult
}

// Push pushes a new layer to the layer stack (above the previous layer).
func Push(layer Layer) chan LayerResult {
	ch := make(chan LayerResult, 1)
	layerStack = append(layerStack, LayerWrapper{
		Layer:    layer,
		resultCh: ch,
	})
	return ch
}

// Pop pops the most recently added layer from the layer stack.
func Pop(result LayerResult) {
	if len(layerStack) > 0 {
		// send result to subscriber
		layerStack[len(layerStack)-1].resultCh <- result
		close(layerStack[len(layerStack)-1].resultCh)
		// pop layer from end of stack
		layerStack = layerStack[:len(layerStack)-1]
	}
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
	// create container sized half the window height
	container := NewUIContainer(NewPadding(5), func() pixel.Rect {
		b := win.Bounds()
		return b.Resized(b.Center(), pixel.V(b.Size().X, b.Size().Y*0.5))
	})

	menu := &MainMenu{
		uiContainer: container,
		createBtn:   NewButton("Create Game", colornames.Paleturquoise, colornames.White),
		joinBtn:     NewButton("Join Game", colornames.Palegreen, colornames.White),
		settingsBtn: NewButton("Settings", colornames.Palevioletred, colornames.White),
	}

	container.AddButton(menu.createBtn, menu.joinBtn, menu.settingsBtn)

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
		Pop(Default)
		Push(gameLayer)

	case m.joinBtn.Clicked():
		m.joinBtn.ToggleEnabled()
	case m.settingsBtn.Clicked():
		m.settingsBtn.ToggleEnabled()
	}

	if win.JustPressed(pixelgl.KeyEscape) {
		win.SetClosed(true)
	}
}

// Draw draws the main menu layer to the window.
func (m *MainMenu) Draw() {
	win.SetMatrix(pixel.IM)

	win.Clear(colornames.Whitesmoke)
	m.uiContainer.Draw()
}

// Game is the main interactive game functionality layer.
type Game struct {
	tileGrid     *world.TileGrid
	mainPlayer   *player.Player
	otherPlayers map[string]*player.Player

	cam           pixel.Matrix
	camScale      float64
	locked        bool
	overlayResult chan LayerResult
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
	if g.locked {
		// check for response from overlay menu layer
		select {
		case res := <-g.overlayResult:
			if res == Quit {
				win.SetClosed(true)
				return
			}
			g.locked = false
		default:
			return
		}
	}

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
			g.camScale += 0.02
		}
	}
	if win.Pressed(pixelgl.KeyF) {
		if g.camScale > 0 {
			g.camScale -= 0.02
		}
	}
	if win.JustPressed(pixelgl.KeyEscape) {
		g.locked = true
		g.overlayResult = Push(NewOverlayMenu())
	}
	// handle mouse movement
	if win.MousePosition() != win.MousePreviousPosition() {
		// point mainPlayer at mouse
		mousePos := g.cam.Unproject(win.MousePosition())
		g.mainPlayer.PointTo(mousePos)
	}
}

// Draw draws the game layer to the window.
func (g *Game) Draw() {
	// window camera
	g.cam = pixel.IM.Scaled(g.mainPlayer.Pos, g.camScale).Moved(win.Bounds().Center().Sub(g.mainPlayer.Pos))
	win.SetMatrix(g.cam)

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
type OverlayMenu struct {
	uiContainer *UIContainer
	resumeBtn   *Button
	serverBtn   *Button
	quitBtn     *Button
}

// NewOverlayMenu creates and initialises a new overlay menu layer.
func NewOverlayMenu() *OverlayMenu {
	// create container sized half the window height
	container := NewUIContainer(NewPadding(5), func() pixel.Rect {
		b := win.Bounds()
		return b.Resized(b.Center(), pixel.V(b.Size().X, b.Size().Y*0.5))
	})

	menu := &OverlayMenu{
		uiContainer: container,
		resumeBtn:   NewButton("Resume", colornames.Paleturquoise, colornames.White),
		serverBtn:   NewButton("Server Settings", colornames.Palegreen, colornames.White),
		quitBtn:     NewButton("Quit Game", colornames.Palevioletred, colornames.White),
	}
	container.AddButton(menu.resumeBtn, menu.serverBtn, menu.quitBtn)
	
	return menu
}

// Update updates the overlay menu layer logic.
func (m *OverlayMenu) Update(dt float64) {
	switch {
	case m.resumeBtn.Clicked() || win.JustPressed(pixelgl.KeyEscape):
		Pop(Default)
	case m.serverBtn.Clicked():
		m.serverBtn.ToggleEnabled()
	case m.quitBtn.Clicked():
		Pop(Quit)
	}
}

// Draw draws the overlay menu layer to the window.
func (m *OverlayMenu) Draw() {
	win.SetMatrix(pixel.IM)
	m.uiContainer.Draw()
}
