package scene

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

var (
	basicFontAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
)

// Padding represents a set of UI padding values.
type Padding struct {
	Top, Right, Bottom, Left float64
}

// NewPadding takes between 0 and 4 padding values and returns a corresponding Padding struct. The order of assignment
// is the same as the CSS standard.
func NewPadding(values ...float64) Padding {
	switch len(values) {
	case 1:
		// top/bottom/left/right = 1st
		return Padding{values[0], values[0], values[0], values[0]}
	case 2:
		// top/bottom = 1st, left/right = 2nd
		return Padding{values[0], values[1], values[0], values[1]}
	case 3:
		// top = 1st, left/right = 2nd, bottom = 3rd
		return Padding{values[0], values[1], values[2], values[1]}
	case 4:
		// top = 1st, right = 2nd, bottom = 3rd, left = 4th
		return Padding{values[0], values[1], values[2], values[3]}
	}
	return Padding{}
}

// UIContainer is a structure which draws child UI components
type UIContainer struct {
	buttons    []*Button
	padding    Padding
	boundsFunc func() pixel.Rect
}

// NewUIContainer creates and initialises a new UIContainer. The padding is applied to all children UI elements.
func NewUIContainer(padding Padding, boundsFunc func() pixel.Rect) *UIContainer {
	return &UIContainer{
		padding:    padding,
		boundsFunc: boundsFunc,
	}
}

// AddButton adds a button to the UIContainer button stack.
func (c *UIContainer) AddButton(btn ...*Button) {
	c.buttons = append(c.buttons, btn...)
}

// Draw draws the UIContainer and its contents.
func (c *UIContainer) Draw() {
	bounds := c.boundsFunc()

	// represents the maximum area a button can fill (i.e. before padding has been applied)
	btnWidth := bounds.W()
	btnHeight := bounds.H() / float64(len(c.buttons))

	for i, btn := range c.buttons {
		padding := c.padding
		yOffset := float64(i) * btnHeight

		btnBounds := pixel.Rect{
			Min: pixel.V(bounds.Min.X+padding.Left, bounds.Min.Y+yOffset+padding.Bottom),
			Max: pixel.V(bounds.Min.X+btnWidth-padding.Right, bounds.Min.Y+yOffset+btnHeight-padding.Top),
		}
		btn.Draw(btnBounds)
	}
}

// Button is a standard UI button.
type Button struct {
	// Enabled determines if a button can be clicked.
	Enabled bool

	bgColour      pixel.RGBA
	bgColourAlt   pixel.RGBA
	imd           *imdraw.IMDraw
	text          *text.Text
	textColour    pixel.RGBA
	textColourAlt pixel.RGBA
}

// Draw draws the button background and label text.
func (b *Button) Draw(bounds pixel.Rect) {
	// TODO: move this to an Update/Resize func and call only when window is resized
	// background
	b.imd = imdraw.New(nil)
	b.imd.Color = b.bgColour
	b.imd.Push(
		bounds.Min,
		pixel.V(bounds.Min.X, bounds.Max.Y),
		bounds.Max,
		pixel.V(bounds.Max.X, bounds.Min.Y),
	)
	b.imd.Polygon(0)
	b.imd.Draw(win)

	// text
	textScaleFactor := bounds.H() / b.text.Bounds().H()

	textYOffset := (bounds.H() * 0.5) - (b.text.Bounds().H() * textScaleFactor * 0.35)
	textXOffset := (bounds.W() * 0.5) - (b.text.Bounds().W() * textScaleFactor * 0.5)

	textPos := bounds.Min.Add(pixel.V(textXOffset, textYOffset))

	// scale width to fit background
	b.text.Draw(win, pixel.IM.Scaled(b.text.Orig, textScaleFactor).Moved(textPos))
}

// NewButton creates and initialises a new Button.
func NewButton(label string, bgColour, textColour color.Color) *Button {
	b := &Button{
		bgColour:      pixel.ToRGBA(bgColour),
		bgColourAlt:   fadeColour(bgColour, buttonFadeAlpha),
		text:          text.New(pixel.ZV, basicFontAtlas),
		textColour:    fadeColour(textColour, buttonFadeAlpha),
		textColourAlt: pixel.ToRGBA(textColour),
	}
	b.text.Color = b.textColour
	b.text.WriteString(label)
	return b
}

const buttonFadeAlpha = 0.8

func fadeColour(colour color.Color, alpha float64) pixel.RGBA {
	c := pixel.ToRGBA(colour)
	c.A = alpha
	return c
}
