package player

import (
	"time"

	"github.com/faiface/pixel"
	"github.com/pkg/errors"
)

type Ammo string

const (
	PistolAmmo  Ammo = "pistol"
	RifleAmmo   Ammo = "rifle"
	ShotgunAmmo Ammo = "shotgun"
)

func (p *MainPlayer) AddWeapon(name WeaponName) error {
	w, ok := weapons[name]
	if !ok {
		return errors.New("weapon \"" + string(name) + "\" does not exist")
	}
	w.currentAmmoCapacity = w.maxAmmoCapacity
	w.state = Ready

	p.Lock()
	p.Weapons = append(p.Weapons, &w)
	p.Unlock()
	return nil
}

type WeaponName string

const (
	Deagle WeaponName = "Deagle"
	M4A1   WeaponName = "M4A1"
)

var weapons = map[WeaponName]ProjectileWeapon{
	Deagle: {
		ammoType:        PistolAmmo,
		automatic:       false,
		maxAmmoCapacity: 7,
		barrelLength:    10,
		maxSpreadAngle:  3.2,

		fireDelay:   time.Millisecond * 500,
		reloadDelay: time.Second * 2,
	},
	M4A1: {
		ammoType:        RifleAmmo,
		automatic:       true,
		maxAmmoCapacity: 30,
		barrelLength:    20,
		maxSpreadAngle:  3,

		fireDelay:   time.Millisecond * 150,
		reloadDelay: time.Second * 3,
	},
}

type WeaponState string

const (
	Ready     WeaponState = "ready"
	Attacking WeaponState = "attacking"
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

	state           WeaponState
	stateChangeTime time.Time
}

func (w *ProjectileWeapon) Attack(playerPos pixel.Vec) {
	if w.state != Ready {
		return
	}
	w.state = Attacking
	w.stateChangeTime = time.Now()
	if w.automatic {

	}


}

func (w *ProjectileWeapon) Reload() {
	if w.state != Ready {
		return
	}
	w.state = Reloading
	w.stateChangeTime = time.Now()
}

func (w *ProjectileWeapon) Update(dt float64) {
	switch w.state {
	case Attacking:
		if time.Now().Sub(w.stateChangeTime) >= w.fireDelay {
			w.state = Ready
		}
		return
	case Reloading:
		if time.Now().Sub(w.stateChangeTime) >= w.reloadDelay {
			w.state = Ready
		}
		return
	}

}

// Projectile represents a single projectile.
type Projectile struct {
	pos       pixel.Vec
	velocity  pixel.Vec
	spawnTime time.Time
	ttl       time.Duration
}

