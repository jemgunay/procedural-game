package player

import (
	"fmt"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/procedural-game/client"
	"github.com/jemgunay/procedural-game/server"
	"github.com/pkg/errors"
)

var (
	AmmoStore    map[Ammo]int
	Armoury      []*ProjectileWeapon
	ActiveWeapon *ProjectileWeapon
	Projectiles  []Projectile

	projectileSpeed   = 100.0
	isWeaponTriggered bool
)

// InitArmoury initialises the main player's armoury, setting the starting ammo and weapons.
func InitArmoury() {
	// give initial ammo stock
	AmmoStore = map[Ammo]int{
		PistolAmmo:  14,
		RifleAmmo:   60,
		ShotgunAmmo: 20,
	}

	// give pistol
	if err := CollectWeapon(Deagle); err != nil {
		fmt.Printf("failed to add new weapon: %s\n", err)
	}

	// give rifle
	if err := CollectWeapon(M4A1); err != nil {
		fmt.Printf("failed to add new weapon: %s\n", err)
	}
}

// Ammo represents an ammunition type.
type Ammo string

// Ammunition type constants.
const (
	PistolAmmo  Ammo = "pistol"
	RifleAmmo   Ammo = "rifle"
	ShotgunAmmo Ammo = "shotgun"
)

// CollectWeapon adds the specified weapon to the player's weapon inventory.
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

// SwitchWeapon switches the player's active weapon to one in their inventory.
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
	fmt.Printf("switched to %s\n", ActiveWeapon)
}

// WeaponName represents a weapon name.
type WeaponName string

// Weapon type constants.
const (
	Deagle WeaponName = "Deagle"
	M4A1   WeaponName = "M4A1"
)

// weapons is the list of supported weapons
var weapons = map[WeaponName]ProjectileWeapon{
	Deagle: {
		weaponName:      Deagle,
		ammoType:        PistolAmmo,
		automatic:       false,
		maxAmmoCapacity: 7,
		barrelLength:    40,
		maxSpreadAngle:  3.2,

		fireDelay:   time.Millisecond * 500,
		reloadDelay: time.Second * 3,
	},
	M4A1: {
		weaponName:      M4A1,
		ammoType:        RifleAmmo,
		automatic:       true,
		maxAmmoCapacity: 30,
		barrelLength:    70,
		maxSpreadAngle:  4,

		fireDelay:   time.Millisecond * 150,
		reloadDelay: time.Second * 3,
	},
}

// WeaponState represents the current state of a weapon.
type WeaponState string

const (
	// Ready indicates that the weapon is ready to be fired or reloaded.
	Ready WeaponState = "ready"
	// Attacking indicates that the weapon is attacking and cannot be reloaded.
	Attacking WeaponState = "attacking"
	// Reloading indicates that the weapon is reloading and cannot be fired.
	Reloading WeaponState = "reloading"
)

// ProjectileWeapon is a weapon that produces projectiles.
type ProjectileWeapon struct {
	weaponName          WeaponName
	ammoType            Ammo
	automatic           bool
	maxAmmoCapacity     uint
	currentAmmoCapacity uint
	fireDelay           time.Duration
	reloadDelay         time.Duration

	barrelLength   float64
	maxSpreadAngle float64

	state           WeaponState
	stateChangeTime time.Time
	sync.RWMutex
}

// String returns the weapon's name as a string.
func (p *ProjectileWeapon) String() string {
	return string(p.weaponName)
}

// Projectile represents a single projectile. It contains time information to determine when it should be destroyed.
type Projectile struct {
	startPos, pos pixel.Vec
	velocity      pixel.Vec
	spawnTime     time.Time
	ttl           time.Duration
}

// Reload sets the currently active weapon's state to reloading assuming the weapon is in a reloadable state.
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

// Attack causes the weapon to be fireable assuming the weapon is in a state to be fired.
func Attack() {
	isWeaponTriggered = true
}

// StopAttack prevents a weapon from attacking.
func StopAttack() {
	isWeaponTriggered = false
}

// Shoot causes a projectile to be fired from a weapon. It determines the initial position and direction the projectile
// should have, and alters the active weapon's state.
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

	projectileUnit := pixel.Unit(p.Orientation())
	startPos := p.Pos().Add(projectileUnit.Scaled(ActiveWeapon.barrelLength))
	// determine offset to position bullet to the right on the tip of the weapon barrel
	startPos = startPos.Add(pixel.Unit(p.Orientation() - 90).Scaled(19))

	projectile := Projectile{
		startPos:  startPos,
		pos:       startPos,
		velocity:  projectileUnit.Scaled(projectileSpeed),
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

	client.Send(server.Message{
		"projectile",
		"",
	})
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
			timeAlive := float64(time.Now().UTC().Sub(p.spawnTime)/time.Millisecond) / 100
			p.pos = p.startPos.Add(p.velocity.Scaled(timeAlive))
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
