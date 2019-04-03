// Package ui implements graphical user interface components.
package ui

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
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

// Container is a structure which draws child UI components
type Container struct {
	buttons    []*Button
	padding    Padding
	boundsFunc func() pixel.Rect
}

// NewContainer creates and initialises a new Container. The padding is applied to all children UI elements.
func NewContainer(padding Padding, boundsFunc func() pixel.Rect) *Container {
	return &Container{
		padding:    padding,
		boundsFunc: boundsFunc,
	}
}

// AddButton adds a button to the Container button stack.
func (c *Container) AddButton(btn ...*Button) {
	// reverse button order before appending
	for i := len(btn)/2 - 1; i >= 0; i-- {
		opp := len(btn) - 1 - i
		btn[i], btn[opp] = btn[opp], btn[i]
	}
	c.buttons = append(btn, c.buttons...)
}

// Draw draws the Container and its contents.
func (c *Container) Draw(win *pixelgl.Window) {
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
		btn.Draw(win, btnBounds)
	}
}

// styling presets/modifiers
const btnFadeAlpha = 0.8

var (
	btnColourDisabled    = pixel.RGB(0.9, 0.9, 0.9)
	btnColourDisabledAlt = pixel.RGB(0.8, 0.8, 0.8)
)

// Button is a standard UI button.
type Button struct {
	enabled bool
	clicked bool

	bgColour       pixel.RGBA
	bgColourAlt    pixel.RGBA
	label          string
	labelColour    pixel.RGBA
	labelColourAlt pixel.RGBA
}

// Enabled indicates if the button should be styled with normal or disabled colours.
func (b *Button) Enabled() bool {
	return b.enabled
}

// ToggleEnabled toggles the button's enabled state.
func (b *Button) ToggleEnabled() {
	b.enabled = !b.enabled
}

// Clicked can be used to poll a button to determine if it has been clicked since the last check.
func (b *Button) Clicked() bool {
	if !b.enabled {
		return false
	}
	if b.clicked {
		// reset clicked value once polled
		b.clicked = false
		return true
	}
	return false
}

// Draw draws the button background and label label.
func (b *Button) Draw(win *pixelgl.Window, bounds pixel.Rect) {
	bg := imdraw.New(nil)
	label := text.New(pixel.ZV, basicFontAtlas)

	// colourise
	if bounds.Contains(win.MousePosition()) {
		// check if button has been clicked
		if win.JustPressed(pixelgl.MouseButton1) {
			b.clicked = true
		}
		if b.enabled {
			bg.Color = b.bgColour
			label.Color = b.labelColour
		} else {
			bg.Color = btnColourDisabled
			label.Color = btnColourDisabledAlt
		}
	} else {
		if b.enabled {
			bg.Color = b.bgColourAlt
			label.Color = b.labelColourAlt
		} else {
			bg.Color = btnColourDisabledAlt
			label.Color = btnColourDisabled
		}
	}

	// background
	bg.Push(
		bounds.Min,
		pixel.V(bounds.Min.X, bounds.Max.Y),
		bounds.Max,
		pixel.V(bounds.Max.X, bounds.Min.Y),
	)
	bg.Polygon(0)
	bg.Draw(win)

	// label text
	label.WriteString(b.label)
	labelScaleFactor := bounds.H() / label.Bounds().H()
	labelYOffset := (bounds.H() * 0.5) - (label.Bounds().H() * labelScaleFactor * 0.35)
	labelXOffset := (bounds.W() * 0.5) - (label.Bounds().W() * labelScaleFactor * 0.5)
	labelPos := bounds.Min.Add(pixel.V(labelXOffset, labelYOffset))

	// scale width to fit background
	label.Draw(win, pixel.IM.Scaled(label.Orig, labelScaleFactor).Moved(labelPos))
}

// NewButton creates and initialises a new Button.
func NewButton(label string, bgColour, labelColour color.Color) *Button {
	b := &Button{
		enabled:        true,
		bgColour:       pixel.ToRGBA(bgColour),
		bgColourAlt:    fadeColour(bgColour, btnFadeAlpha),
		label:          label,
		labelColour:    fadeColour(labelColour, btnFadeAlpha),
		labelColourAlt: pixel.ToRGBA(labelColour),
	}
	return b
}

func fadeColour(colour color.Color, alpha float64) pixel.RGBA {
	c := pixel.ToRGBA(colour)
	c.A = alpha
	return c
}
