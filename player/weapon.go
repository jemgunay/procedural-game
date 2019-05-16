package player

import (
	"time"

	"github.com/faiface/pixel"
)

type ProjectileWeapon struct {
	name string
	fireRate time.Duration

	Ammo
}

type Ammo struct {
	variant string

}

// Bullet represents a single bullet.
type Bullet struct {
	pos       pixel.Vec
	velocity  pixel.Vec
	baseSpeed float32
}
