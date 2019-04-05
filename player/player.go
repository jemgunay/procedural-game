// Package player contains player creation and processing logic.
package player

import (
	"math"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/jemgunay/game/file"
)

// Player represents a drawable client player.
type Player struct {
	Name            string
	pos             pixel.Vec
	prevPos         pixel.Vec
	speed           float64
	orientation     float64
	prevOrientation float64
	sprite          *pixel.Sprite

	sync.RWMutex
}

// New creates and initialises a new player.
func New(name string) (*Player, error) {
	// create sprite
	sprite, err := file.CreateSprite(file.Player)
	if err != nil {
		return nil, err
	}

	return &Player{
		Name:        name,
		pos:         pixel.ZV,
		speed:       500.0,
		orientation: 0.0,
		sprite:      sprite,
	}, nil
}

// Draw draws a player onto a window.
func (p *Player) Draw(win *pixelgl.Window) {
	p.RLock()
	p.sprite.Draw(win, pixel.IM.Moved(p.pos).Rotated(p.pos, p.orientation))
	p.RUnlock()
}

// Up moves the player upwards.
func (p *Player) Up(dt float64) {
	p.Lock()
	p.pos.Y += p.speed * dt
	p.Unlock()
}

// Down moves the player downwards.
func (p *Player) Down(dt float64) {
	p.Lock()
	p.pos.Y -= p.speed * dt
	p.Unlock()
}

// Left moves the player leftwards.
func (p *Player) Left(dt float64) {
	p.Lock()
	p.pos.X -= p.speed * dt
	p.Unlock()
}

// Right moves the player rightwards.
func (p *Player) Right(dt float64) {
	p.Lock()
	p.pos.X += p.speed * dt
	p.Unlock()
}

// Pos retrieves the player position.
func (p *Player) Pos() pixel.Vec {
	p.Lock()
	pos := p.pos
	p.Unlock()
	return pos
}

// SetPos moves the player to the specified coordinates.
func (p *Player) SetPos(target pixel.Vec) {
	p.Lock()
	p.pos = target
	p.Unlock()
}

// Orientation gets the player's orientation.
func (p *Player) Orientation() float64 {
	p.Lock()
	rot := p.orientation
	p.Unlock()
	return rot
}

// SetOrientation sets the player's orientation.
func (p *Player) SetOrientation(target float64) {
	p.Lock()
	p.orientation = target
	p.Unlock()
}

// PointTo rotates the player to face the specified target.
func (p *Player) PointTo(target pixel.Vec) {
	p.Lock()
	p.orientation = math.Atan2(target.Y-p.pos.Y, target.X-p.pos.X)
	p.Unlock()
}

func (p *Player) HasMoved() bool {
	p.RLock()
	moved := p.pos != p.prevPos && p.orientation != p.prevOrientation
	p.RUnlock()
	return moved
}