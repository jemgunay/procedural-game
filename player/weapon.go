package player

import (
	"time"

	"github.com/faiface/pixel"
)

type Armoury struct {
	AmmoStore map[Ammo]int
	Weapons   []ProjectileWeapon
}

func NewArmoury() *Armoury {
	return &Armoury{
		AmmoStore: map[Ammo]int{
			PistolAmmo:  14,
			RifleAmmo:   60,
			ShotgunAmmo: 20,
		},
		Weapons: []ProjectileWeapon{
			weapons["deagle"],
		},
	}
}

type Ammo string

const (
	PistolAmmo  Ammo = "pistol"
	RifleAmmo   Ammo = "rifle"
	ShotgunAmmo Ammo = "shotgun"
)

var weapons = map[string]ProjectileWeapon{
	"deagle": {
		ammoType:            PistolAmmo,
		automatic:           false,
		maxAmmoCapacity:     7,
		currentAmmoCapacity: 7,
		fireDelay:           time.Millisecond * 500,
		reloadDelay:         time.Second * 3,

		barrelLength:   10,
		maxSpreadAngle: 3,

		state: Ready,
	},
}

type WeaponState string

const (
	Ready     WeaponState = "ready"
	Reloading WeaponState = "reloading"
)

type ProjectileWeapon struct {
	ammoType            Ammo
	automatic           bool
	maxAmmoCapacity     uint
	currentAmmoCapacity uint
	fireDelay           time.Duration
	reloadDelay         time.Duration

	barrelLength   float32
	maxSpreadAngle float32

	state WeaponState
}

// Projectile represents a single projectile.
type Projectile struct {
	pos      pixel.Vec
	velocity pixel.Vec
	speed    float32
}
