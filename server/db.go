package server

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
	"unicode"
)

const (
	// MinUsernameLength is the minimum username length.
	MinUsernameLength = 5
	// MaxUsernameLength is the maximum username length.
	MaxUsernameLength = 12
)

// User represents a persistent user record.
type User struct {
	name string
	// vitals is a stringified composition of the core player data such as position, health, etc
	vitals string
	x, y   float64
	rot    float64
	health uint64

	conn   net.Conn
	exitCh chan struct{}
}

// Send marshals and writes a message to a user's client.
func (u *User) Send(msg Message) {
	if u.conn == nil {
		return
	}

	rawMsg := msg.Pack()
	if _, err := u.conn.Write(rawMsg); err != nil {
		fmt.Printf("failed to write \"%s\" to %s: %s\n", string(rawMsg), u.conn.RemoteAddr(), err)
	}
}

// UserDB is a database of users.
type UserDB struct {
	users map[string]User
	rand  *rand.Rand

	sync.RWMutex
}

// Get safely retrieves the user corresponding with the provided username. If the user doesn't exist, an empty User is
// returned.
func (d *UserDB) Get(username string) (User, bool) {
	d.RLock()
	user, ok := d.users[username]
	d.RUnlock()
	return user, ok
}

// Update safely updates a user by its username.
func (d *UserDB) Update(user User) {
	d.Lock()
	d.users[user.name] = user
	d.Unlock()
}

// Broadcast broadcasts a message to all connected users except those in the specified list of exclusion usernames.
func (d *UserDB) Broadcast(msg Message, excludeUsernames ...string) {
	d.RLock()
	for _, user := range d.users {
		skipUser := false
		// if current user is in exclusion list, then skip sending message to that user
		for _, name := range excludeUsernames {
			if user.name == name {
				skipUser = true
				break
			}
		}
		if skipUser {
			continue
		}
		user.Send(msg)
	}
	d.RUnlock()
}

// Create creates a new user in the user DB given a username and connection.
func (d *UserDB) Create(username string, conn net.Conn) (User, error) {
	// create new user at the top of this func so that the conn can be consumed on error
	newUser := User{
		name:   username,
		x:      float64(d.rand.Intn(8000)),
		y:      float64(d.rand.Intn(8000)),
		rot:    0.0,
		health: 100,
		conn:   conn,
		exitCh: make(chan struct{}, 1),
	}

	// validate username
	switch {
	case len(username) < MinUsernameLength:
		return newUser, fmt.Errorf("username must have a minimum length of %d characters", MinUsernameLength)
	case len(username) > MaxUsernameLength:
		return newUser, fmt.Errorf("username length must not exceed %d characters", MaxUsernameLength)
	}

	// allow letters, numbers, underscore and hyphen
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' {
			return newUser, errors.New("username can only contain letters, numbers, underscores and hyphens")
		}
	}

	var err error
	d.RLock()
	for _, user := range d.users {
		if user.name == username {
			err = errors.New("username already taken")
			break
		}
	}
	d.RUnlock()
	if err != nil {
		return newUser, err
	}

	// set initial random position
	newUser.vitals = ConcatVitals(
		newUser.x,
		newUser.y,
		newUser.rot,
		newUser.health,
	)

	// insert new user into DB
	d.Lock()
	d.users[newUser.name] = newUser
	d.Unlock()
	return newUser, nil
}

// Connect associates an existing user in the user DB with a new connection.
func (d *UserDB) Connect(username string, conn net.Conn) (User, error) {
	d.RLock()
	user, ok := d.users[username]
	d.RUnlock()
	if !ok {
		return User{}, errors.New("user not found in DB")
	}

	// check user is not already connected to prevent kicking off different client
	if user.conn != nil {
		return User{}, errors.New("user already connected")
	}

	// update connection
	user.conn = conn
	d.Lock()
	d.users[username] = user
	d.Unlock()
	return user, nil
}

// Disconnect clears the reference to a user's connection.
func (d *UserDB) Disconnect(user User) {
	// clear connection reference before updating DB
	user.conn = nil
	d.Lock()
	// set connection to nil
	d.users[user.name] = user
	d.Unlock()

	// broadcast user leaving message to all remaining connected users
	d.Broadcast(Message{
		Type:  "disconnect",
		Value: user.name,
	}, user.name)
}

// Projectile represents a server projectile instance.
type Projectile struct {
	owner          string
	x, y           float64
	startX, startY float64
	velX, velY     float64
	spawnTime      time.Time
	ttl            time.Duration
}

// ProjectileDB is a database of projectiles.
type ProjectileDB struct {
	projectiles []Projectile

	sync.RWMutex
}

func (d *ProjectileDB) Create(projectile Projectile) {
	d.Lock()
	d.projectiles = append(d.projectiles, projectile)
	d.Unlock()
}

func (d *ProjectileDB) Update() {
	d.Lock()
	var aliveProjectiles []Projectile
	for _, p := range d.projectiles {
		// only retain projectiles with unexpired TTLs
		if !time.Now().UTC().After(p.spawnTime.Add(p.ttl)) {
			timeAlive := float64(time.Now().UTC().Sub(p.spawnTime)/time.Millisecond) / 100
			p.x = p.startX + p.velX*timeAlive
			p.y = p.startY + p.velY*timeAlive
			aliveProjectiles = append(aliveProjectiles, p)
		}
	}
	d.projectiles = aliveProjectiles
	d.Unlock()
}

