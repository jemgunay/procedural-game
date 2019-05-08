// Package scene manages the scene and executes different scene layers.
package scene

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"github.com/jemgunay/game/client"
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
		VSync:     false,
		Resizable: true,
	}

	// create window
	var err error
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		fmt.Printf("failed create new window: %s\n", err)
		return
	}

	// push a main menu layer to the scene
	Push(NewMainMenu())

	// limit update cycles to 120 FPS
	frameRateLimiter := time.Tick(time.Second / 120)
	prevTimestamp := time.Now()
	// main game loop
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
		<-frameRateLimiter
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

// Game is the main interactive game functionality layer.
type Game struct {
	gameType   GameType
	tileGrid   *world.TileGrid
	mainPlayer *player.Player
	players    *player.Store

	cam           pixel.Matrix
	seed          string
	camScale      float64
	prevMousePos  pixel.Vec
	locked        bool
	overlayResult chan LayerResult
}

// GameType is used to differentiate between a client and server game instance.
type GameType string

// Game type constants.
const (
	Client GameType = "client"
	Server GameType = "server"
)

// NewGame creates and initialises a new Game layer.
func NewGame(gameType GameType, addr string, playerName string) (game *Game, err error) {
	// connect to server
	if err = client.Start(addr); err != nil {
		return nil, fmt.Errorf("client failed to start: %s", err)
	}

	client.Send(server.Message{
		Type:  "register",
		Value: playerName,
	})

	// wait for register success
	var (
		seed, uuid string, pos,
		rot float64
	)
	for {
		switch msg := client.Poll(); msg.Type {
		case "register_success":
			components := strings.Split(msg.Value, "|")
			if len(components) != 3 {
				return nil, errors.New("malformed register_success response")
			}
			seed = components[1]
			uuid, pos, rot, err = splitPosReq(components[2])
			if err != nil {
				return nil, err
			}

			fmt.Printf("user UUID: %s\n", uuid)

		case "register_failure", "connect_failure":
			server.Shutdown()
			return nil, errors.New(msg.Value)

		default:
			continue
		}
		break
	}

	// parse seed into integer
	var seedNum int64
	for _, c := range seed {
		seedNum += int64(c)
	}

	// generate world
	fmt.Printf("generating new world with a seed of \"%s\" (%d)\n", seed, seedNum)
	tileGrid := world.NewTileGrid(seedNum)
	if err = tileGrid.GenerateChunk(); err != nil {
		server.Shutdown()
		return nil, fmt.Errorf("failed to generate world: %s", err)
	}

	// create new game instance
	game = &Game{
		gameType: gameType,
		seed:     seed,
		tileGrid: tileGrid,
		players:  player.NewStore(),
		camScale: 0.5,
	}

	// create main player
	game.mainPlayer, err = game.players.Add(playerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %s", err)
	}
	game.mainPlayer.SetPos(pos)
	game.mainPlayer.SetOrientation(rot)

	// receive and process incoming requests from the server
	go game.processServerUpdates()

	return
}

// processServerUpdates polls the client for incoming requests from the server and applies the corresponding client/
// player updates.
func (g *Game) processServerUpdates() {
	for {
		switch msg := client.Poll(); msg.Type {
		// new player joined the game
		case "user_joined":
			_, err := g.players.Add(msg.Value)
			if err != nil {
				break
			}
			fmt.Println(msg.Value + " joined the game!")

		// initialise world and already existing players after joining a new game
		case "init_world":
			fmt.Printf("init world request: %s\n", msg.Value)
			items := strings.Split(msg.Value, "/")
			for _, item := range items {
				name, pos, rot, err := splitPosReq(item)
				if err != nil {
					fmt.Printf("failed to split pos request: %s", err)
					break
				}

				// add new player
				p, err := g.players.Add(name)
				if err != nil {
					fmt.Printf("failed to add player \"%s\": %s\n", name, err)
					break
				}
				p.SetPos(pos)
				p.SetOrientation(rot)
			}

		// update a player's position and orientation
		case "pos":
			name, pos, rot, err := splitPosReq(msg.Value)
			if err != nil {
				fmt.Printf("failed to split pos request: %s", err)
				break
			}

			// find new player
			p, err := g.players.Find(name)
			if err != nil {
				fmt.Printf("player doesn't exist: %s\n", err)
				break
			}
			p.SetPos(pos)
			p.SetOrientation(rot)

		// remove a player from the game
		case "disconnect":
			fmt.Println(msg.Value + " left the game!")
			g.players.Remove(msg.Value)
		}
	}
}

// process a "pos" message from the server into its separate components
func splitPosReq(val string) (name string, pos pixel.Vec, rot float64, err error) {
	components := strings.Split(val, "|")
	if len(components) != 4 {
		err = fmt.Errorf("incorrect pos component count")
		return
	}
	x, err := strconv.ParseFloat(components[1], 64)
	if err != nil {
		err = fmt.Errorf("failed to parse X: %s\n", err)
		return
	}
	y, err := strconv.ParseFloat(components[2], 64)
	if err != nil {
		err = fmt.Errorf("failed to parse Y: %s\n", err)
		return
	}
	rot, err = strconv.ParseFloat(components[3], 64)
	if err != nil {
		err = fmt.Errorf("failed to parse rot: %s\n", err)
		return
	}
	name = components[0]
	pos = pixel.V(x, y)
	return
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
		if g.camScale < 1.2 {
			g.camScale += 0.02
		}
	}
	if win.Pressed(pixelgl.KeyF) {
		if g.camScale > 0.07 {
			g.camScale -= 0.02
		}
	}
	if win.JustPressed(pixelgl.KeyEscape) {
		g.locked = true
		g.overlayResult = Push(NewOverlayMenu())
	}

	// handle mouse movement
	if win.MousePosition() != g.prevMousePos {
		// point mainPlayer at mouse
		mousePos := g.cam.Unproject(win.MousePosition())
		g.mainPlayer.PointTo(mousePos)
	}
	g.prevMousePos = win.MousePosition()

	// send pos & orientation update to server
	if g.mainPlayer.HasMoved() {
		pos := g.mainPlayer.Pos()
		client.Send(server.Message{
			Type:  "pos",
			Value: fmt.Sprintf("%f|%f|%f", pos.X, pos.Y, g.mainPlayer.Orientation()),
		})
	}
}

// Draw draws the game layer to the window.
func (g *Game) Draw() {
	// window camera
	pos := g.mainPlayer.Pos()
	g.cam = pixel.IM.Scaled(pos, g.camScale).Moved(win.Bounds().Center().Sub(pos))
	win.SetMatrix(g.cam)

	win.Clear(colornames.Greenyellow)
	// draw tiles
	g.tileGrid.Draw(win)
	// draw players
	g.players.Draw(win)
}
