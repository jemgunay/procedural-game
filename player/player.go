// Package player contains player creation and processing logic.
package player

import (
	"math"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// Player represents a drawable client player.
type Player struct {
	name            string
	pos             pixel.Vec
	prevPos         pixel.Vec
	baseSpeed       float64
	orientation     float64
	prevOrientation float64
	sprite          *pixel.Sprite

	sync.RWMutex
}

// Draw draws a player onto a window.
func (p *Player) Draw(win *pixelgl.Window) {
	p.RLock()
	p.sprite.Draw(win, pixel.IM.Moved(p.pos).Scaled(p.pos, 0.3).Rotated(p.pos, p.orientation))
	p.RUnlock()
}

// Up moves the player upwards.
func (p *Player) Up(dt float64) {
	p.Lock()
	p.pos.Y += p.baseSpeed * dt
	p.Unlock()
}

// Down moves the player downwards.
func (p *Player) Down(dt float64) {
	p.Lock()
	p.pos.Y -= p.baseSpeed * dt
	p.Unlock()
}

// Left moves the player leftwards.
func (p *Player) Left(dt float64) {
	p.Lock()
	p.pos.X -= p.baseSpeed * dt
	p.Unlock()
}

// Right moves the player rightwards.
func (p *Player) Right(dt float64) {
	p.Lock()
	p.pos.X += p.baseSpeed * dt
	p.Unlock()
}

// Pos retrieves the player position.
func (p *Player) Pos() pixel.Vec {
	p.RLock()
	pos := p.pos
	p.RUnlock()
	return pos
}

// Speed retrieves the player baseSpeed.
func (p *Player) Speed() float64 {
	p.Lock()
	speed := p.baseSpeed
	p.Unlock()
	return speed
}

// SetPos moves the player to the specified coordinates.
func (p *Player) SetPos(target pixel.Vec) {
	p.Lock()
	p.prevPos = p.pos
	p.pos = target
	p.Unlock()
}

// Orientation gets the player's orientation.
func (p *Player) Orientation() float64 {
	p.RLock()
	rot := p.orientation
	p.RUnlock()
	return rot
}

// SetOrientation sets the player's orientation.
func (p *Player) SetOrientation(target float64) {
	p.Lock()
	p.prevOrientation = p.orientation
	p.orientation = target
	p.Unlock()
}

// PointTo rotates the player to face the specified target.
func (p *Player) PointTo(target pixel.Vec) {
	p.Lock()
	p.prevOrientation = p.orientation
	p.orientation = math.Atan2(target.Y-p.pos.Y, target.X-p.pos.X)
	p.Unlock()
}

// HasMoved determines if the player has moved position or changed orientation.
func (p *Player) HasMoved() bool {
	p.RLock()
	moved := p.pos != p.prevPos && p.orientation != p.prevOrientation
	p.RUnlock()
	return moved
}