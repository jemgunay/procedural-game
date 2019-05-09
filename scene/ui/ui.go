// Package ui implements graphical user interface components.
package ui

import (
	"image/color"
	"time"

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

// Drawer is the interface which defines a UI component that can be drawn to a window given its container's bounds.
type Drawer interface {
	Draw(win *pixelgl.Window, bounds pixel.Rect)
}

// FixedContainer is a fixed size container which draws child UI components.
type FixedContainer struct {
	elements   []Drawer
	padding    Padding
	boundsFunc func() pixel.Rect
}

// NewFixedContainer creates and initialises a new FixedContainer. The padding is applied to all children UI elements.
func NewFixedContainer(padding Padding, boundsFunc func() pixel.Rect) *FixedContainer {
	return &FixedContainer{
		padding:    padding,
		boundsFunc: boundsFunc,
	}
}

// AddElement adds an element to the FixedContainer elements stack.
func (c *FixedContainer) AddElement(element ...Drawer) {
	c.elements = append(element, c.elements...)
}

// Draw draws the FixedContainer and its contents.
func (c *FixedContainer) Draw(win *pixelgl.Window) {
	bounds := c.boundsFunc()

	// represents the maximum area a button can fill before padding has been applied)
	elementWidth := bounds.W()
	elementHeight := bounds.H() / float64(len(c.elements))

	for i, element := range c.elements {
		padding := c.padding
		yOffset := float64(i) * elementHeight

		elementBounds := pixel.Rect{
			Min: pixel.V(bounds.Min.X+padding.Left, bounds.Max.Y-yOffset-elementHeight+padding.Top),
			Max: pixel.V(bounds.Min.X+elementWidth-padding.Right, bounds.Max.Y-yOffset-padding.Bottom),
		}
		element.Draw(win, elementBounds)
	}
}

// ScrollContainer is a scrollable UI container which draws child UI components.
type ScrollContainer struct {
	elements   []Drawer
	padding    Padding
	boundsFunc func() pixel.Rect

	scrollBarWidth       float64
	scrollBarColour      pixel.RGBA
	scrollBtnColour      pixel.RGBA
	scrollBtnColourAlt   pixel.RGBA
	scrollBarPressed     bool
	scrollBtnBounds      pixel.Rect
	scrollBtnClickDeltaY float64
}

// NewScrollContainer creates and initialises a new ScrollContainer. The padding is applied to all children UI elements.
func NewScrollContainer(padding Padding, boundsFunc func() pixel.Rect) *ScrollContainer {
	s := &ScrollContainer{
		padding:         padding,
		boundsFunc:      boundsFunc,
		scrollBarWidth:  25,
		scrollBarColour: pixel.ToRGBA(LightBlue),
		scrollBtnColour: pixel.ToRGBA(Orange),
	}
	s.scrollBtnColourAlt = fadeColour(s.scrollBtnColour, 0.9)
	return s
}

// AddElement adds an element to the ScrollContainer elements stack.
func (c *ScrollContainer) AddElement(element ...Drawer) {
	c.elements = append(element, c.elements...)
}

// Draw draws the ScrollContainer and its contents.
func (c *ScrollContainer) Draw(win *pixelgl.Window) {
	bounds := c.boundsFunc()

	// draw elements at fixed size
	elementWidth := bounds.W()
	elementHeight := 170.0

	contentHeight := float64(len(c.elements)) * elementHeight
	contentToBoundsRatio := bounds.H() / contentHeight
	scrollBtnHeight := bounds.H() * contentToBoundsRatio
	// cap min scroll button height to prevent it getting too small on smaller bounds heights
	if scrollBtnHeight < bounds.H()/2 {
		scrollBtnHeight = bounds.H() / 2
	}

	// draw scroll bar background
	scrollBarBG := imdraw.New(nil)
	scrollBarBG.Color = c.scrollBarColour
	scrollBarBG.Push(
		pixel.V(bounds.Max.X-c.scrollBarWidth, bounds.Max.Y),
		pixel.V(bounds.Max.X, bounds.Max.Y),
		pixel.V(bounds.Max.X, bounds.Min.Y),
		pixel.V(bounds.Max.X-c.scrollBarWidth, bounds.Min.Y),
	)
	scrollBarBG.Polygon(0)
	scrollBarBG.Draw(win)

	// init scroll button Y to top of bar if not yet set
	if c.scrollBtnBounds.Size().Len() == 0 {
		c.scrollBtnBounds.Min.Y = bounds.Max.Y - scrollBtnHeight
		c.scrollBtnBounds.Max.Y = bounds.Max.Y
	}

	var contentScrollOffset float64
	// only draw scroll bar button and offset content Y if the content height is greater than the bounds height
	if contentHeight > bounds.H() {
		// update scroll X pos in case bounds are horizontally resized
		c.scrollBtnBounds.Min.X = bounds.Max.X - c.scrollBarWidth
		c.scrollBtnBounds.Max.X = bounds.Max.X

		// on scroll bar button mouse click/release
		if win.JustPressed(pixelgl.MouseButton1) && c.scrollBtnBounds.Contains(win.MousePosition()) {
			c.scrollBarPressed = true
			c.scrollBtnClickDeltaY = c.scrollBtnBounds.Max.Y - win.MousePosition().Y
		}
		if win.JustReleased(pixelgl.MouseButton1) {
			c.scrollBarPressed = false
			c.scrollBtnClickDeltaY = 0
		}

		// move scroll bar button to mouse Y
		if c.scrollBarPressed {
			c.scrollBtnBounds.Max.Y = win.MousePosition().Y + c.scrollBtnClickDeltaY
		} else {
			// handle mouse scroll button - scroll distance is the height of one element
			if win.MouseScroll().Y != 0 {
				c.scrollBtnBounds.Max.Y += win.MouseScroll().Y * (bounds.H() / float64(len(c.elements)))
			}
		}
		c.scrollBtnBounds.Min.Y = c.scrollBtnBounds.Max.Y - scrollBtnHeight

		// prevent scrolling above permitted scroll area
		if c.scrollBtnBounds.Max.Y > bounds.Max.Y {
			c.scrollBtnBounds.Max.Y = bounds.Max.Y
			c.scrollBtnBounds.Min.Y = c.scrollBtnBounds.Max.Y - scrollBtnHeight
		} else if c.scrollBtnBounds.Min.Y < bounds.Min.Y {
			c.scrollBtnBounds.Min.Y = bounds.Min.Y
			c.scrollBtnBounds.Max.Y = c.scrollBtnBounds.Min.Y + scrollBtnHeight
		}

		// draw scroll bar button
		scrollBtn := imdraw.New(nil)
		scrollBtn.Color = c.scrollBtnColour
		if c.scrollBarPressed {
			scrollBtn.Color = c.scrollBtnColourAlt
		}
		scrollBtn.Push(
			pixel.V(bounds.Max.X-c.scrollBarWidth, c.scrollBtnBounds.Max.Y),
			pixel.V(bounds.Max.X, c.scrollBtnBounds.Max.Y),
			pixel.V(bounds.Max.X, c.scrollBtnBounds.Max.Y-scrollBtnHeight),
			pixel.V(bounds.Max.X-c.scrollBarWidth, c.scrollBtnBounds.Max.Y-scrollBtnHeight),
		)
		scrollBtn.Polygon(0)
		scrollBtn.Draw(win)

		// determine how far to offset content Y based on scroll position
		contentScrollDeltaMax := contentHeight - bounds.H()
		scrollBarMaxScrolledDist := bounds.H() - scrollBtnHeight
		currentScrolledPercent := (bounds.Max.Y - c.scrollBtnBounds.Max.Y) / scrollBarMaxScrolledDist
		contentScrollOffset = contentScrollDeltaMax * currentScrolledPercent
	}

	// draw elements
	for i, element := range c.elements {
		padding := c.padding
		padding.Right += c.scrollBarWidth
		yOffset := float64(i)*elementHeight - contentScrollOffset

		elementBounds := pixel.Rect{
			Min: pixel.V(bounds.Min.X+padding.Left, bounds.Max.Y-yOffset-elementHeight+padding.Top),
			Max: pixel.V(bounds.Min.X+elementWidth-padding.Right, bounds.Max.Y-yOffset-padding.Bottom),
		}
		element.Draw(win, elementBounds)
	}
}

// styling presets/modifiers
const (
	btnFadeAlpha = 0.85
)

// Colour presets.
var (
	Blue      = color.RGBA{175, 238, 238, 243}
	Green     = color.RGBA{152, 251, 152, 243}
	Red       = color.RGBA{219, 112, 147, 243}
	Orange    = color.RGBA{255, 222, 173, 255}
	LightBlue = color.RGBA{240, 248, 255, 255}

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

// NewButton creates and initialises a new Button.
func NewButton(label string, bgColour, labelColour color.Color) *Button {
	return &Button{
		enabled:        true,
		bgColour:       pixel.ToRGBA(bgColour),
		bgColourAlt:    fadeColour(bgColour, btnFadeAlpha),
		label:          label,
		labelColour:    pixel.ToRGBA(labelColour),
		labelColourAlt: fadeColour(labelColour, btnFadeAlpha),
	}
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
			bg.Color = b.bgColourAlt
			label.Color = b.labelColourAlt
		} else {
			bg.Color = btnColourDisabled
			label.Color = btnColourDisabledAlt
		}
	} else {
		if b.enabled {
			bg.Color = b.bgColour
			label.Color = b.labelColour
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

	// label
	label.WriteString(b.label)
	labelScaleFactor := bounds.H() / label.Bounds().H()
	labelYOffset := (bounds.H() * 0.5) - (label.Bounds().H() * labelScaleFactor * 0.35)
	labelXOffset := (bounds.W() * 0.5) - (label.Bounds().W() * labelScaleFactor * 0.5)
	labelPos := bounds.Min.Add(pixel.V(labelXOffset, labelYOffset))

	// scale width to fit background
	label.Draw(win, pixel.IM.Scaled(label.Orig, labelScaleFactor).Moved(labelPos))
}

const textBoxTickerDuration = time.Millisecond * 600

// TextBox is a text input box UI element.
type TextBox struct {
	hasFocus bool

	bgColour      pixel.RGBA
	label         string
	text          string
	textColour    pixel.RGBA
	textColourAlt pixel.RGBA

	cursorLastTick time.Time
	cursorChar     string
}

// NewTextBox creates and initialises a UI text input box.
func NewTextBox(label string, bgColour, textColour color.Color) *TextBox {
	return &TextBox{
		bgColour:      pixel.ToRGBA(bgColour),
		label:         label,
		textColour:    pixel.ToRGBA(textColour),
		textColourAlt: fadeColour(textColour, btnFadeAlpha),
	}
}

// Draw draws a TextBox onto a window within the specified bounds. It also processes keyboard input and click focus.
func (t *TextBox) Draw(win *pixelgl.Window, bounds pixel.Rect) {
	// background
	bg := imdraw.New(nil)
	bg.Push(
		bounds.Min,
		pixel.V(bounds.Min.X, bounds.Max.Y),
		bounds.Max,
		pixel.V(bounds.Max.X, bounds.Min.Y),
	)
	bg.Polygon(0)
	bg.Draw(win)

	// give input focus if clicked & lose focus if mouse clicked outside of input
	if win.JustPressed(pixelgl.MouseButton1) {
		t.hasFocus = bounds.Contains(win.MousePosition())
		if t.hasFocus {
			t.cursorLastTick = time.Now()
			t.cursorChar = "|"
		} else {
			t.cursorChar = ""
		}
	}

	// store new text input if text box has focus
	if t.hasFocus {
		if time.Since(t.cursorLastTick) >= textBoxTickerDuration {
			t.cursorLastTick = time.Now()
			if t.cursorChar == "|" {
				t.cursorChar = " "
			} else {
				t.cursorChar = "|"
			}
		}
		newText := win.Typed()
		if newText != "" {
			t.cursorChar = "|"
			t.cursorLastTick = time.Now()
		}
		t.text += newText

		switch {
		// lose text input focus on enter key press
		case win.JustPressed(pixelgl.KeyEnter), win.Repeated(pixelgl.KeyEnter):
			t.hasFocus = false
			t.cursorChar = ""

		// delete character on backspace key press
		case win.JustPressed(pixelgl.KeyBackspace), win.Repeated(pixelgl.KeyBackspace):
			if len(t.text) > 0 {
				t.text = t.text[:len(t.text)-1]
				t.cursorChar = "|"
				t.cursorLastTick = time.Now()
			}
		}
	}

	// label
	label := text.New(pixel.ZV, basicFontAtlas)
	label.Color = t.textColour
	label.WriteString(t.label)

	labelHeight := bounds.H() * 0.3
	labelScaleFactor := labelHeight / label.Bounds().H()
	labelXOffset := (bounds.W() * 0.5) - (label.Bounds().W() * labelScaleFactor * 0.5)
	labelPos := bounds.Min.Add(pixel.V(0, bounds.H()-(label.Bounds().H()*labelScaleFactor*0.7)))

	label.Draw(win, pixel.IM.Scaled(label.Orig, labelScaleFactor).Moved(labelPos))

	// text
	inputText := text.New(pixel.ZV, basicFontAtlas)
	inputText.Color = t.textColour
	inputText.WriteString(t.text)
	inputText.WriteString(t.cursorChar)

	labelHeight = bounds.H() * 0.7
	labelScaleFactor = labelHeight / inputText.Bounds().H()
	labelYOffset := (labelHeight * 0.5) - (inputText.Bounds().H() * labelScaleFactor * 0.35)
	labelXOffset = (bounds.W() * 0.5) - (inputText.Bounds().W() * labelScaleFactor * 0.5)
	labelPos = bounds.Min.Add(pixel.V(labelXOffset, labelYOffset))

	inputText.Draw(win, pixel.IM.Scaled(inputText.Orig, labelScaleFactor).Moved(labelPos))
}

// SetText sets the text box input value.
func (t *TextBox) SetText(text string) {
	t.text = text
}

// Text returns the text box input value.
func (t *TextBox) Text() string {
	return t.text
}

// applies an alpha value to the provided colour
func fadeColour(colour color.Color, alpha float64) pixel.RGBA {
	c := pixel.ToRGBA(colour)
	c.A = alpha
	return c
}
