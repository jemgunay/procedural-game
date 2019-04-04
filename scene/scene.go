// Package scene manages the scene and executes different scene layers.
package scene

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/client"
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

// Game is the main interactive game functionality layer.
type Game struct {
	gameType   GameType
	tileGrid   *world.TileGrid
	playerName string
	players PlayerStore

	cam           pixel.Matrix
	camScale      float64
	prevMousePos  pixel.Vec
	locked        bool
	overlayResult chan LayerResult
}

type GameType string

const (
	Client GameType = "client"
	Server GameType = "server"
)

type PlayerStore struct {
	players map[string]*player.Player
	sync.RWMutex
}

func (s *PlayerStore) GetCopy(username string) player.Player {
	s.RLock()
	defer s.RUnlock()
	p, ok := s.players[username]
	if !ok {
		return player.Player{}
	}
	return *p
}

func (s *PlayerStore) Add(username string) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.players[username]; ok {
		return errors.New("player with that username already exists")
	}

	newPlayer, err := player.New(username)
	if err != nil {
		return fmt.Errorf("failed to create new player: %s", err)
	}

	s.players[username] = newPlayer
	return nil
}

func (s *PlayerStore) Draw() {
	s.RLock()
	for _, p := range s.players {
		p.Draw(win)
	}
	s.RUnlock()
}

func (s *PlayerStore) Execute(username string, f func(p *player.Player)) error {
	s.Lock()
	defer s.Unlock()
	pl, ok := s.players[username]
	if ok {
		return errors.New("player not found")
	}
	f(pl)
	return nil
}

// NewGame creates and initialises a new Game layer.
func NewGame(gameType GameType) (*Game, error) {
	// generate world
	tileGrid := world.NewTileGrid()
	if err := tileGrid.GenerateChunk(); err != nil {
		return nil, fmt.Errorf("failed to generate world: %s", err)
	}

	// temp player name
	var userName string

	// start server if
	if gameType == Server {
		if err := server.Start(":9000"); err != nil {
			return nil, fmt.Errorf("server failed to start: %s", err)
		}
		userName = "jemgunay"
	} else {
		// TODO: remove this test username
		userName = "willyG"
	}

	// connect to server
	if err := client.Start("localhost:9000"); err != nil {
		return nil, fmt.Errorf("client failed to start: %s", err)
	}

	client.Send(server.Message{
		Type:  "register",
		Value: userName,
	})

	// wait for register success
	for {
		msg := client.Poll()
		switch msg.Type {
		case "register_success":
			fmt.Printf("user UUID: %s\n", msg.Value)
		case "register_failure":
			return nil, errors.New(msg.Value)
		default:
			continue
		}
		break
	}

	g := &Game{
		gameType:   gameType,
		tileGrid:   tileGrid,
		playerName: userName,
		players: PlayerStore{
			players: make(map[string]*player.Player),
		},
		camScale: 1.0,
	}

	// start main
	go func() {
		for {
			msg := client.Poll()
			switch msg.Type {
			case "user_joined":
				if err := g.players.Add(msg.Value); err != nil {
					break
				}

			case "pos":
				fmt.Printf("pos request: %s\n", msg.Value)
				components := strings.Split(msg.Value, "|")
				if len(components) != 4 {
					fmt.Println("incorrect pos component count")
					break
				}
				x, err := strconv.ParseFloat(components[0], 64)
				if err != nil {
					fmt.Printf("failed to parse X: %s\n", err)
					break
				}
				y, err := strconv.ParseFloat(components[1], 64)
				if err != nil {
					fmt.Printf("failed to parse Y: %s\n", err)
					break
				}
				rot, err := strconv.ParseFloat(components[2], 64)
				if err != nil {
					fmt.Printf("failed to parse rot: %s\n", err)
					break
				}

				f := func(p *player.Player) {
					p.Pos = pixel.V(x, y)
					p.Orientation = rot
				}
				if err = g.players.Execute(components[4], f); err != nil {
					fmt.Printf("failed to update player pos: %s\n", err)
					break
				}
			}
		}
	}()

	return g, nil
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
		//g.mainPlayer.Up(dt)
		g.players.Execute(g.playerName, func(p *player.Player) {
			p.Up(dt)
		})
	}
	if win.Pressed(pixelgl.KeyS) {
		//g.mainPlayer.Down(dt)
		g.players.Execute(g.playerName, func(p *player.Player) {
			p.Down(dt)
		})
	}
	if win.Pressed(pixelgl.KeyA) {
		//g.mainPlayer.Left(dt)
		g.players.Execute(g.playerName, func(p *player.Player) {
			p.Left(dt)
		})
	}
	if win.Pressed(pixelgl.KeyD) {
		//g.mainPlayer.Right(dt)
		g.players.Execute(g.playerName, func(p *player.Player) {
			p.Right(dt)
		})
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
	if win.MousePosition() != g.prevMousePos {
		// point mainPlayer at mouse
		mousePos := g.cam.Unproject(win.MousePosition())
		//g.mainPlayer.PointTo(mousePos)
		g.players.Execute(g.playerName, func(p *player.Player) {
			p.PointTo(mousePos)
		})
	}
	g.prevMousePos = win.MousePosition()

	// send pos & orientation update to server
	mainPlayer := g.players.GetCopy(g.playerName)
	if mainPlayer.Pos != mainPlayer.PrevPos && mainPlayer.Orientation != mainPlayer.PrevOrientation {
		client.Send(server.Message{
			Type:  "pos",
			Value: fmt.Sprintf("%f|%f|%f", mainPlayer.Pos.X, mainPlayer.Pos.Y, mainPlayer.Orientation),
		})
	}
}

// Draw draws the game layer to the window.
func (g *Game) Draw() {
	// window camera
	mainPlayer := g.players.GetCopy(g.playerName)
	g.cam = pixel.IM.Scaled(mainPlayer.Pos, g.camScale).Moved(win.Bounds().Center().Sub(mainPlayer.Pos))
	win.SetMatrix(g.cam)

	win.Clear(colornames.Greenyellow)
	// draw tiles
	g.tileGrid.Draw(win)
	// draw players
	g.players.Draw()
}
