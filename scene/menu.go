package scene

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/scene/ui"
	"golang.org/x/image/colornames"
)

// MainMenu is the main menu layer which is first displayed upon game startup.
type MainMenu struct {
	uiContainer *ui.Container
	createBtn   *ui.Button
	joinBtn     *ui.Button
	settingsBtn *ui.Button
}

// NewMainMenu creates and initialises a new main menu layer.
func NewMainMenu() *MainMenu {
	// create container sized half the window height
	container := ui.NewContainer(ui.NewPadding(5), func() pixel.Rect {
		b := win.Bounds()
		return b.Resized(b.Center(), pixel.V(b.Size().X, b.Size().Y*0.5))
	})

	menu := &MainMenu{
		uiContainer: container,
		createBtn:   ui.NewButton("Create Game", colornames.Paleturquoise, colornames.White),
		joinBtn:     ui.NewButton("Join Game", colornames.Palegreen, colornames.White),
		settingsBtn: ui.NewButton("Settings", colornames.Palevioletred, colornames.White),
	}

	container.AddButton(menu.createBtn, menu.joinBtn, menu.settingsBtn)

	return menu
}

// Update updates the main menu layer logic.
func (m *MainMenu) Update(dt float64) {
	switch {
	case m.createBtn.Clicked():
		// create a new game layer
		gameLayer, err := NewGame(Server)
		if err != nil {
			fmt.Printf("failed to create game layer: %s\n", err)
			return
		}
		// pop main menu and push game layer
		Pop(Default)
		Push(gameLayer)

	case m.joinBtn.Clicked():
		// create a new game layer
		gameLayer, err := NewGame(Client)
		if err != nil {
			fmt.Printf("failed to create game layer: %s\n", err)
			return
		}
		// pop main menu and push game layer
		Pop(Default)
		Push(gameLayer)

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

// OverlayMenu is the overlay menu layer which is drawn over the main game layer.
type OverlayMenu struct {
	uiContainer *ui.Container
	resumeBtn   *ui.Button
	serverBtn   *ui.Button
	quitBtn     *ui.Button
}

// NewOverlayMenu creates and initialises a new overlay menu layer.
func NewOverlayMenu() *OverlayMenu {
	// create container sized half the window height
	container := ui.NewContainer(ui.NewPadding(5), func() pixel.Rect {
		b := win.Bounds()
		return b.Resized(b.Center(), pixel.V(b.Size().X, b.Size().Y*0.5))
	})

	menu := &OverlayMenu{
		uiContainer: container,
		resumeBtn:   ui.NewButton("Resume", colornames.Paleturquoise, colornames.White),
		serverBtn:   ui.NewButton("Server Settings", colornames.Palegreen, colornames.White),
		quitBtn:     ui.NewButton("Quit Game", colornames.Palevioletred, colornames.White),
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
	m.uiContainer.Draw(win)
}
