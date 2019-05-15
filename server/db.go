package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
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
	conn      net.Conn
	name      string
	blocked   bool
	posRotStr string
}

// Send marshals and writes a message to a user's client.
func (u *User) Send(msg Message) {
	if u.conn == nil {
		return
	}

	rawMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("failed to process outbound request: %s:\n%v\n", err, msg)
		return
	}

	if _, err := u.conn.Write(append(rawMsg, '\n')); err != nil {
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

// Update safely updates a user by its UUID.
func (d *UserDB) Update(user User) {
	d.Lock()
	d.users[user.name] = user
	d.Unlock()
}

// Broadcast broadcasts a message to all connected users except those in the specified list of exclusion UUIDs.
func (d *UserDB) Broadcast(msg Message, excludeUserNames ...string) {
	d.RLock()
	for _, user := range d.users {
		for _, name := range excludeUserNames {
			if user.name == name {
				continue
			}
			user.Send(msg)
		}
	}
	d.RUnlock()
}

// Create creates a new user in the user DB given a username and connection.
func (d *UserDB) Create(username string, conn net.Conn) (User, error) {
	// create new user at the top of this func so that the conn can be consumed on error
	newUser := User{
		name: username,
		conn: conn,
	}

	// validate username
	switch {
	case len(username) < MinUsernameLength:
		return newUser, errors.New("username must have a minimum length of 5 characters")
	case len(username) > MaxUsernameLength:
		return newUser, errors.New("username length must not exceed 12 characters")
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
	newUser.posRotStr = fmt.Sprintf("%f|%f|%f", float64(d.rand.Intn(8000)), float64(d.rand.Intn(8000)), 0.0)

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
