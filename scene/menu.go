package scene

import (
	"fmt"
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/scene/ui"
	"github.com/jemgunay/game/server"
	"golang.org/x/image/colornames"
)

// MainMenu is the main menu layer which is first displayed upon game startup.
type MainMenu struct {
	uiContainer *ui.FixedContainer
	createBtn   *ui.Button
	joinBtn     *ui.Button
	settingsBtn *ui.Button
}

// NewMainMenu creates and initialises a new main menu layer.
func NewMainMenu() *MainMenu {
	// create container sized half the window height
	container := ui.NewFixedContainer(ui.NewPadding(5), func() pixel.Rect {
		b := win.Bounds()
		return b.Resized(b.Center(), pixel.V(b.Size().X, b.Size().Y*0.5))
	})

	menu := &MainMenu{
		uiContainer: container,
		createBtn:   ui.NewButton("Create Game", ui.Blue, colornames.White),
		joinBtn:     ui.NewButton("Join Game", ui.Green, colornames.White),
		settingsBtn: ui.NewButton("Settings", ui.Red, colornames.White),
	}

	container.AddElement(menu.createBtn, menu.joinBtn, menu.settingsBtn)

	return menu
}

// Update updates the main menu layer logic.
func (m *MainMenu) Update(dt float64) {
	switch {
	case m.createBtn.Clicked():
		// pop main menu and push create game layer
		Pop(Default)
		Push(NewCreateGameMenu())

	case m.joinBtn.Clicked():
		// pop main menu and push join game layer
		Pop(Default)
		Push(NewJoinGameMenu())

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

	win.Clear(colornames.White)
	m.uiContainer.Draw(win)
}

// CreateGameMenu is the menu layer for creating/hosting a game.
type CreateGameMenu struct {
	uiContainer         *ui.ScrollContainer
	backBtn             *ui.Button
	seedTextInput       *ui.TextBox
	portTextInput       *ui.TextBox
	playerNameTextInput *ui.TextBox
	startBtn            *ui.Button
}

// NewCreateGameMenu creates and initialises a new CreateGameMenu layer.
func NewCreateGameMenu() *CreateGameMenu {
	// create container sized half the window height
	container := ui.NewScrollContainer(ui.NewPadding(5), win.Bounds)

	menu := &CreateGameMenu{
		uiContainer:         container,
		backBtn:             ui.NewButton("Back", ui.Blue, colornames.White),
		seedTextInput:       ui.NewTextBox("World Seed", colornames.White, colornames.Black),
		portTextInput:       ui.NewTextBox("Port", colornames.White, colornames.Black),
		playerNameTextInput: ui.NewTextBox("Player Name", colornames.White, colornames.Black),
		startBtn:            ui.NewButton("Start", ui.Green, colornames.White),
	}
	menu.portTextInput.SetText("9000")

	container.AddElement(menu.backBtn, menu.seedTextInput, menu.portTextInput, menu.playerNameTextInput, menu.startBtn)
	return menu
}

// Update updates the game creation menu layer logic.
func (m *CreateGameMenu) Update(dt float64) {
	switch {
	case win.JustPressed(pixelgl.KeyEscape), m.backBtn.Clicked():
		Pop(Default)
		Push(NewMainMenu())

	case m.startBtn.Clicked():
		// parse seed into integer
		seedInput := m.seedTextInput.Text()
		portInput, err := strconv.ParseUint(m.portTextInput.Text(), 10, 64)
		if err != nil {
			fmt.Println("invalid port provided")
			return
		}

		// start server or client
		if err = server.Start(fmt.Sprintf(":%d", portInput), seedInput); err != nil {
			fmt.Printf("server failed to start: %s\n", err)
			return
		}

		// create a new game layer
		gameLayer, err := NewGame(Server, fmt.Sprintf(":%d", portInput), m.playerNameTextInput.Text())
		if err != nil {
			fmt.Printf("failed to create game layer: %s\n", err)
			server.Shutdown()
			return
		}

		// pop main menu and push game layer
		Pop(Default)
		Push(gameLayer)
	}
}

// Draw draws the game creation menu layer logic.
func (m *CreateGameMenu) Draw() {
	win.SetMatrix(pixel.IM)

	win.Clear(colornames.White)
	m.uiContainer.Draw(win)
}

// JoinGameMenu is the menu layer for joining an existing game.
type JoinGameMenu struct {
	uiContainer         *ui.ScrollContainer
	backBtn             *ui.Button
	hostAddrTextInput   *ui.TextBox
	playerNameTextInput *ui.TextBox
	joinBtn             *ui.Button
}

// NewJoinGameMenu creates and initialises a new JoinGameMenu layer.
func NewJoinGameMenu() *JoinGameMenu {
	// create container sized half the window height
	container := ui.NewScrollContainer(ui.NewPadding(5), win.Bounds)

	menu := &JoinGameMenu{
		uiContainer:         container,
		backBtn:             ui.NewButton("Back", ui.Blue, colornames.White),
		hostAddrTextInput:   ui.NewTextBox("Server Address", colornames.White, colornames.Black),
		playerNameTextInput: ui.NewTextBox("Player Name", colornames.White, colornames.Black),
		joinBtn:             ui.NewButton("Join", ui.Green, colornames.White),
	}
	menu.hostAddrTextInput.SetText("localhost:9000")

	container.AddElement(menu.backBtn, menu.hostAddrTextInput, menu.playerNameTextInput, menu.joinBtn)
	return menu
}

// Update updates the game join menu layer logic.
func (m *JoinGameMenu) Update(dt float64) {
	switch {
	case win.JustPressed(pixelgl.KeyEscape), m.backBtn.Clicked():
		Pop(Default)
		Push(NewMainMenu())

	case m.joinBtn.Clicked():
		// create a new game layer
		gameLayer, err := NewGame(Client, m.hostAddrTextInput.Text(), m.playerNameTextInput.Text())
		if err != nil {
			fmt.Printf("failed to create game layer: %s\n", err)
			return
		}

		// pop main menu and push game layer
		Pop(Default)
		Push(gameLayer)
	}
}

// Draw draws the game creation menu layer logic.
func (m *JoinGameMenu) Draw() {
	win.SetMatrix(pixel.IM)

	win.Clear(colornames.White)
	m.uiContainer.Draw(win)
}

// OverlayMenu is the overlay menu layer which is drawn over the main game layer.
type OverlayMenu struct {
	uiContainer *ui.FixedContainer
	resumeBtn   *ui.Button
	serverBtn   *ui.Button
	quitBtn     *ui.Button
}

// NewOverlayMenu creates and initialises a new overlay menu layer.
func NewOverlayMenu() *OverlayMenu {
	// create container sized half the window height
	container := ui.NewFixedContainer(ui.NewPadding(5), func() pixel.Rect {
		b := win.Bounds()
		return b.Resized(b.Center(), pixel.V(b.Size().X, b.Size().Y*0.5))
	})

	menu := &OverlayMenu{
		uiContainer: container,
		resumeBtn:   ui.NewButton("Resume", ui.Blue, colornames.White),
		serverBtn:   ui.NewButton("Server Settings", ui.Green, colornames.White),
		quitBtn:     ui.NewButton("Quit Game", ui.Red, colornames.White),
	}
	container.AddElement(menu.resumeBtn, menu.serverBtn, menu.quitBtn)

	return menu
}

// Update updates the overlay menu layer logic.
func (m *OverlayMenu) Update(dt float64) {
	switch {
	case m.resumeBtn.Clicked(), win.JustPressed(pixelgl.KeyEscape):
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
	m.uiContainer.Draw(win)
}
