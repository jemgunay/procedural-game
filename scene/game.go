package scene

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"github.com/jemgunay/procedural-game/client"
	"github.com/jemgunay/procedural-game/player"
	"github.com/jemgunay/procedural-game/scene/world"
	"github.com/jemgunay/procedural-game/server"
)

// Game is the main interactive game functionality layer.
type Game struct {
	gameType   GameType
	tileGrid   *world.TileGrid
	players    *player.Store
	mainPlayer *player.Player

	camPos        pixel.Vec
	camMatrix     pixel.Matrix
	seed          string
	camScale      float64
	prevMousePos  pixel.Vec
	locked        bool
	overlayResult chan LayerResult
	exitCh        chan struct{}
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
		Type:  "connect",
		Value: playerName,
	})

	// wait for register success
	var (
		seed, name string
		pos        pixel.Vec
		rot        float64
		health     uint64
	)
	// TODO: add a connect timeout
	for {
		msg, err := client.Poll()
		if err != nil {
			if err == client.ErrQueueClosed {
				return nil, fmt.Errorf("failed to handshake with server: %s", err)
			}
			continue
		}

		switch msg.Type {
		case "register_success", "connect_success":
			data, err := msg.Unpack()
			if err != nil {
				return nil, fmt.Errorf("failed to unpack register_success message: %s", err)
			}

			seed = data.GetString("seed")
			name = data.GetString("name")
			pos = data.Get("pos").(pixel.Vec)
			rot = data.GetFloat("rot")
			health = data.GetUInt("health")

			fmt.Printf("new user with username: %s\n", name)

		case "register_failure", "connect_failure":
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
		return nil, fmt.Errorf("failed to generate world: %s", err)
	}

	// create player store & new main player
	playerStore := player.NewStore()
	mainPlayer, err := playerStore.Add(playerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %s", err)
	}
	mainPlayer.SetPos(pos)
	mainPlayer.SetOrientation(rot)
	mainPlayer.SetHealth(health)

	player.InitArmoury()

	// create new game instance
	game = &Game{
		gameType:   gameType,
		seed:       seed,
		tileGrid:   tileGrid,
		players:    playerStore,
		mainPlayer: mainPlayer,
		camPos:     mainPlayer.Pos(),
		camScale:   0.5,
		exitCh:     make(chan struct{}, 1),
	}

	// receive and process incoming requests from the server
	go game.processServerUpdates()

	return
}

// processServerUpdates polls the client for incoming requests from the server and applies the corresponding client/
// player updates.
func (g *Game) processServerUpdates() {
	for {
		// exit from poll loop if game has disconnected
		select {
		case <-g.exitCh:
			return
		default:
		}

		// poll for updates from the server
		msg, err := client.Poll()
		if err != nil {
			if err == client.ErrQueueClosed {
				g.Disconnect()
			}
			continue
		}

		switch msg.Type {
		// update a player's position and orientation
		case "vitals":
			data, err := msg.Unpack()
			if err != nil {
				fmt.Printf("failed to split vitals request: %s", err)
				break
			}

			// find new player
			p, err := g.players.Find(data.GetString("name"))
			if err != nil {
				fmt.Printf("player doesn't exist: %s\n", err)
				break
			}
			p.SetPos(data.Get("pos").(pixel.Vec))
			p.SetOrientation(data.GetFloat("rot"))
			p.SetHealth(data.GetUInt("health"))

		// player has fired a projectile
		case "create_projectile":
			data, err := msg.Unpack()
			if err != nil {
				fmt.Printf("create_projectile message incorrectly formatted: %s\n", err)
			}

			startPos := pixel.V(data.GetFloat("startX"), data.GetFloat("startY"))
			vel := pixel.V(data.GetFloat("velX"), data.GetFloat("velY"))
			player.NewProjectile(startPos, vel, data.GetTime("spawnTime"), data.GetDuration("ttl"))

		// new player joined the game
		case "user_joined":
			if _, err := g.players.Add(msg.Value); err != nil {
				fmt.Printf("failed to create user: %s", err)
				break
			}
			fmt.Println(msg.Value + " joined the game!")

		// initialise world and already existing players after joining a new game
		case "init_world":
			// TODO: migrate into msg.Unpack()
			fmt.Printf("init world request: %s\n", msg.Value)
			items := strings.Split(msg.Value, "/")
			for _, item := range items {
				name, pos, rot, health, err := splitPosReq(item)
				if err != nil {
					fmt.Printf("failed to split vitals request: %s\n", err)
					continue
				}

				// add new player
				p, err := g.players.Add(name)
				if err != nil {
					fmt.Printf("failed to add player \"%s\": %s\n", name, err)
					continue
				}
				p.SetPos(pos)
				p.SetOrientation(rot)
				p.SetHealth(health)
			}

		// remove a player from the game
		case "disconnect":
			fmt.Println(msg.Value + " left the game!")
			g.players.Remove(msg.Value)

		// server has initiated shutdown
		case "server_shutdown":
			g.Disconnect()
		}
	}
}

// process a vitals message from the server into its separate components
func splitPosReq(val string) (name string, pos pixel.Vec, rot float64, health uint64, err error) {
	components := strings.Split(val, "|")
	fmt.Println(components)
	if len(components) != 5 {
		err = fmt.Errorf("incorrect vitals component count")
		return
	}
	x, err := strconv.ParseFloat(components[1], 64)
	if err != nil {
		err = fmt.Errorf("failed to parse X: %s", err)
		return
	}
	y, err := strconv.ParseFloat(components[2], 64)
	if err != nil {
		err = fmt.Errorf("failed to parse Y: %s", err)
		return
	}
	rot, err = strconv.ParseFloat(components[3], 64)
	if err != nil {
		err = fmt.Errorf("failed to parse rot: %s", err)
		return
	}
	health, err = strconv.ParseUint(components[4], 10, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse health: %s", err)
		return
	}
	name = components[0]
	pos = pixel.V(x, y)
	return
}

// Update updates the game layer logic.
func (g *Game) Update(dt float64) {
	g.mainPlayer.Update(dt)

	// things that shouldn't update when the overview menu is up should occur here
	if g.locked {
		// check for response from overlay menu layer
		select {
		case res := <-g.overlayResult:
			if res == Quit {
				// quit button pressed
				win.SetClosed(true)
				return
			} else if res == Disconnect {
				// disconnect button pressed
				g.Disconnect()
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
	if win.Pressed(pixelgl.KeyUp) {
		if g.camScale < 1.2 {
			g.camScale += 0.02
		}
	}
	if win.Pressed(pixelgl.KeyDown) {
		if g.camScale > 0.07 {
			g.camScale -= 0.02
		}
	}
	if win.JustPressed(pixelgl.KeyEscape) {
		g.locked = true
		g.overlayResult = Push(NewOverlayMenu(g.gameType))
	}
	switch {
	case win.JustPressed(pixelgl.Key1):
		player.SwitchWeapon(1)
	case win.JustPressed(pixelgl.Key2):
		player.SwitchWeapon(2)
	case win.JustPressed(pixelgl.Key3):
		player.SwitchWeapon(3)
	case win.JustPressed(pixelgl.Key4):
		player.SwitchWeapon(4)
	case win.JustPressed(pixelgl.Key5):
		player.SwitchWeapon(5)
	case win.JustPressed(pixelgl.Key6):
		player.SwitchWeapon(6)
	case win.JustPressed(pixelgl.Key7):
		player.SwitchWeapon(7)
	case win.JustPressed(pixelgl.Key8):
		player.SwitchWeapon(8)
	case win.JustPressed(pixelgl.Key9):
		player.SwitchWeapon(9)
	}

	// handle mouse movement
	if win.MousePosition() != g.prevMousePos {
		// point mainPlayer at mouse
		mousePos := g.camMatrix.Unproject(win.MousePosition())
		g.mainPlayer.PointTo(mousePos)
	}
	g.prevMousePos = win.MousePosition()

	// send pos & orientation update to server
	if g.mainPlayer.HasMoved() {
		pos := g.mainPlayer.Pos()
		client.Send(server.Message{
			Type:  "vitals",
			Value: server.ConcatVitals(pos.X, pos.Y, g.mainPlayer.Orientation(), g.mainPlayer.Health()),
		})
	}

	switch {
	// determine whether to trigger a player attack action
	case win.JustPressed(pixelgl.MouseButton1):
		player.Attack()

	case win.JustReleased(pixelgl.MouseButton1):
		player.StopAttack()

	// determine whether to reload weapon
	case win.Pressed(pixelgl.KeyR):
		player.Reload()

	}

	// smooth player camera tracking
	lerp := g.mainPlayer.Speed() * 0.01 * dt
	camDelta := g.mainPlayer.Pos().Sub(g.camPos).Scaled(lerp)
	g.camPos = g.camPos.Add(camDelta)
}

// Draw draws the game layer to the window.
func (g *Game) Draw() {
	// window camera
	g.camMatrix = pixel.IM.Scaled(g.camPos, g.camScale).Moved(win.Bounds().Center().Sub(g.camPos))
	win.SetMatrix(g.camMatrix)

	win.Clear(colornames.Greenyellow)
	// draw tiles
	g.tileGrid.Draw(win)
	// draw players
	g.players.Draw(win)
	// draw projectiles
	player.DrawProjectiles(win)
}

// Disconnect triggers a client disconnect, followed by a server shutdown if a server is being hosted. The main menu is
// then displayed.
func (g *Game) Disconnect() {
	// disconnect local client before shutting down server
	client.Disconnect()
	g.exitCh <- struct{}{}

	if g.gameType == Server {
		server.Shutdown()
	}
	Pop(Default)
	Push(NewMainMenu())
}
