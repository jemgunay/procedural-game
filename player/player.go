// Package player contains player creation and processing logic.
package player

import (
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/game/file"
)

// Player represents a drawable client player.
type Player struct {
	name        string
	pos         pixel.Vec
	speed       float64
	orientation float64
	sprite      *pixel.Sprite
}

// New creates and initialises a new player.
func New(name string) (*Player, error) {
	// create sprite
	sprite, err := file.CreateSprite(file.Player)
	if err != nil {
		return nil, err
	}

	return &Player{
		name:        name,
		pos:         pixel.ZV,
		speed:       500.0,
		orientation: 0.0,
		sprite:      sprite,
	}, nil
}

// Pos is a getter for the player's position.
func (p *Player) Pos() pixel.Vec {
	return p.pos
}

// Draw draws a player onto a window.
func (p *Player) Draw(win *pixelgl.Window) {
	p.sprite.Draw(win, pixel.IM.Moved(p.pos).Rotated(p.pos, p.orientation))
}

// Up moves the player upwards.
func (p *Player) Up(dt float64) {
	p.pos.Y += p.speed * dt
}

// Down moves the player downwards.
func (p *Player) Down(dt float64) {
	p.pos.Y -= p.speed * dt
}

// Left moves the player leftwards.
func (p *Player) Left(dt float64) {
	p.pos.X -= p.speed * dt
}

// Right moves the player rightwards.
func (p *Player) Right(dt float64) {
	p.pos.X += p.speed * dt
}

// PointTo rotates the player to face the specified target.
func (p *Player) PointTo(target pixel.Vec) {
	p.orientation = math.Atan2(target.Y-p.pos.Y, target.X-p.pos.X)
}
