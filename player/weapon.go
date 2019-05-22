package player

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/pkg/errors"
)

type Ammo string

const (
	PistolAmmo  Ammo = "pistol"
	RifleAmmo   Ammo = "rifle"
	ShotgunAmmo Ammo = "shotgun"
)

func CollectWeapon(name WeaponName) error {
	w, ok := weapons[name]
	if !ok {
		return errors.New("weapon \"" + string(name) + "\" does not exist")
	}
	w.currentAmmoCapacity = w.maxAmmoCapacity
	w.state = Ready

	Armoury = append(Armoury, &w)
	SwitchWeapon(len(Armoury))
	return nil
}

func SwitchWeapon(inventorySlot int) {
	if len(weapons) < inventorySlot {
		return
	}
	if ActiveWeapon != nil {
		if ActiveWeapon.state == Attacking {
			return
		}
		// cancel any reloads
		if ActiveWeapon.state == Reloading {
			ActiveWeapon.state = Ready
		}
	}
	ActiveWeapon = Armoury[inventorySlot-1]
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
		barrelLength:    40,
		maxSpreadAngle:  3.2,

		fireDelay:   time.Millisecond * 500,
		reloadDelay: time.Second * 3,
	},
	M4A1: {
		ammoType:        RifleAmmo,
		automatic:       true,
		maxAmmoCapacity: 30,
		barrelLength:    70,
		maxSpreadAngle:  4,

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

var isWeaponTriggered bool

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
	sync.RWMutex
}

// Projectile represents a single projectile.
type Projectile struct {
	pos       pixel.Vec
	velocity  pixel.Vec
	spawnTime time.Time
	ttl       time.Duration
}

func Reload() {
	if ActiveWeapon == nil {
		return
	}
	if ActiveWeapon.state != Ready {
		return
	}
	if ActiveWeapon.currentAmmoCapacity == ActiveWeapon.maxAmmoCapacity {
		return
	}
	if AmmoStore[ActiveWeapon.ammoType] <= 0 {
		fmt.Println("not enough ammo to reload")
		return
	}

	ActiveWeapon.state = Reloading
	ActiveWeapon.stateChangeTime = time.Now().UTC()
}

func Attack() {
	isWeaponTriggered = true
}

func StopAttack() {
	isWeaponTriggered = false
}

func (p *Player) Shoot() {
	if ActiveWeapon == nil {
		return
	}
	if ActiveWeapon.state != Ready {
		fmt.Printf("weapon not ready to shoot: %s", ActiveWeapon.stateChangeTime)
		return
	}
	if ActiveWeapon.currentAmmoCapacity <= 0 {
		fmt.Println("out of ammo")
		return
	}

	bulletPos := pixel.V(
		p.Pos().X+float64(ActiveWeapon.barrelLength)*math.Cos(p.Orientation()),
		p.Pos().Y+float64(ActiveWeapon.barrelLength)*math.Sin(p.Orientation()),
	)

	projectile := Projectile{
		pos:       bulletPos,
		velocity:  pixel.V(ProjectileSpeed*math.Cos(p.Orientation()), ProjectileSpeed*math.Sin(p.Orientation())),
		spawnTime: time.Now().UTC(),
		ttl:       time.Second * 5,
	}
	Projectiles = append(Projectiles, projectile)

	// consume ammo round
	ActiveWeapon.currentAmmoCapacity--
	fmt.Printf("shot: ammo in weapon: %d/%d, ammo in armoury: %d\n", ActiveWeapon.currentAmmoCapacity, ActiveWeapon.maxAmmoCapacity, AmmoStore[ActiveWeapon.ammoType])
	ActiveWeapon.state = Attacking
	ActiveWeapon.stateChangeTime = time.Now().UTC()

	if !ActiveWeapon.automatic {
		isWeaponTriggered = false
	}
}

func (p *Player) Update(dt float64) {
	if ActiveWeapon != nil {
		ActiveWeapon.Lock()

		switch ActiveWeapon.state {
		case Attacking:
			if ActiveWeapon.stateChangeTime.Add(ActiveWeapon.fireDelay).After(time.Now().UTC()) {
				ActiveWeapon.state = Ready
			}

		case Reloading:
			if ActiveWeapon.stateChangeTime.Add(ActiveWeapon.reloadDelay).After(time.Now().UTC()) {
				requiredAmmo := ActiveWeapon.maxAmmoCapacity - ActiveWeapon.currentAmmoCapacity
				availableAmmo := AmmoStore[ActiveWeapon.ammoType]

				if availableAmmo >= int(requiredAmmo) {
					AmmoStore[ActiveWeapon.ammoType] = availableAmmo - int(requiredAmmo)
					ActiveWeapon.currentAmmoCapacity += requiredAmmo
				} else {
					AmmoStore[ActiveWeapon.ammoType] = 0
					ActiveWeapon.currentAmmoCapacity += uint(availableAmmo)
				}
				ActiveWeapon.state = Ready
				fmt.Printf("reloaded: ammo in weapon: %d/%d, ammo in armoury: %d\n", ActiveWeapon.currentAmmoCapacity, ActiveWeapon.maxAmmoCapacity, AmmoStore[ActiveWeapon.ammoType])
			}
		case Ready:
			if isWeaponTriggered {
				p.Shoot()
			}
		}
		ActiveWeapon.Unlock()
	}

	var aliveProjectiles []Projectile
	for _, p := range Projectiles {
		// only retain projectiles with unexpired TTLs
		if !time.Now().UTC().After(p.spawnTime.Add(p.ttl)) {
			p.pos = p.pos.Add(p.velocity)
			aliveProjectiles = append(aliveProjectiles, p)
		}
	}
	Projectiles = aliveProjectiles
}

func DrawProjectiles(win *pixelgl.Window) {
	for _, p := range Projectiles {
		circle := imdraw.New(nil)
		circle.Color = pixel.RGB(0.2, 0.2, 0.2)
		circle.Push(p.pos)
		circle.Circle(3, 0)
		circle.Draw(win)
	}
}
