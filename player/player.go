// Package player contains player creation and processing logic.
package player

import (
	"fmt"
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
	orientation     float64
	prevOrientation float64
	health          uint64

	baseSpeed float64
	sprite    *pixel.Sprite

	sync.RWMutex
}

type MainPlayer struct {
	*Player
	AmmoStore     map[Ammo]int
	Weapons       []*ProjectileWeapon
	CurrentWeapon *ProjectileWeapon
}

func UpgradeToMain(p *Player) MainPlayer {
	m := MainPlayer{
		Player: p,
		AmmoStore: map[Ammo]int{
			PistolAmmo:  14,
			RifleAmmo:   60,
			ShotgunAmmo: 20,
		},
	}

	if err := m.AddWeapon(Deagle); err != nil {
		fmt.Printf("failed to add new weapon: %s\n", err)
	} else {
		m.CurrentWeapon = m.Weapons[0]
	}
	return m
}

func (p *MainPlayer) Update(dt float64) {
	//p.RLock()
	//p.RUnlock()
	if p.CurrentWeapon != nil {
		p.CurrentWeapon.Update(dt)
	}

}

func (p *MainPlayer) UpdateWeaponState(state WeaponState) {
	// TODO: if not in water etc
	if p.CurrentWeapon == nil {
		return
	}

	switch state {
	case Attacking:
		p.CurrentWeapon.Attack(p.pos)
	case Reloading:
		p.CurrentWeapon.Reload()
	}
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

// Health retrieves the player health.
func (p *Player) Health() uint64 {
	p.RLock()
	health := p.health
	p.RUnlock()
	return health
}

// SetHealth sets the player's health.
func (p *Player) SetHealth(health uint64) {
	p.Lock()
	p.health = health
	p.Unlock()
}

// Pos retrieves the player position.
func (p *Player) Pos() pixel.Vec {
	p.RLock()
	pos := p.pos
	p.RUnlock()
	return pos
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

// Speed retrieves the player baseSpeed.
func (p *Player) Speed() float64 {
	p.RLock()
	speed := p.baseSpeed
	p.RUnlock()
	return speed
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
