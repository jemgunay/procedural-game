package scene

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	basicFontAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
)

type UIContainer struct {
	buttons []*Button
	pixel.Vec
	pixel.Rect
}

func NewUIContainer(rect pixel.Rect) *UIContainer {
	return &UIContainer{
		Rect: rect,
	}
}

func (c *UIContainer) AddButton(btn *Button) {
	c.buttons = append(c.buttons, btn)
}

func (c *UIContainer) Draw() {
	btnWidth := c.W()
	btnHeight := c.H() / float64(len(c.buttons))
	offset := float64(5)

	for i, btn := range c.buttons {
		num := float64(i + 1)
		imd := imdraw.New(nil)
		imd.Color = btn.color
		// bottom left point
		imd.Push(pixel.V(c.X+offset, c.Y+offset))
		imd.Push(pixel.V(c.X+offset, c.Y+(num*btnHeight)-offset))
		imd.Push(pixel.V(c.X+(num*btnWidth)-offset, c.Y+(num*btnHeight)-offset))
		imd.Push(pixel.V(c.X+(num*btnWidth)-offset, c.Y+offset))
		imd.Polygon(0)
		imd.Draw(win)
	}
}

type Button struct {
	color pixel.RGBA
	//text *pixel.Text
}

func NewButton() *Button {
	return &Button{
		color: pixel.ToRGBA(colornames.Cadetblue),
		//text: text.New(pixel.V(100, 500), basicFontAtlas),
	}
}
