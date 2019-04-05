package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/twinj/uuid"
)

// User represents a persistent user record.
type User struct {
	conn net.Conn
	name string
	// a uuid is returned to a new new client which is used as a validation token during future connections
	uuid    string
	blocked bool
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
	sync.RWMutex
}

// Get safely retrieves the user corresponding with the provided UUID. If the user doesn't exist, an empty User is
// returned.
func (d *UserDB) Get(uuid string) User {
	d.RLock()
	user := d.users[uuid]
	d.RUnlock()
	return user
}

// Update safely updates a user by its UUID.
func (d *UserDB) Update(user User) {
	d.Lock()
	d.users[user.uuid] = user
	d.Unlock()
}

// Broadcast broadcasts a message to all connected users except those in the specified list of exclusion UUIDs.
func (d *UserDB) Broadcast(msg Message, excludeUUIDs ...string) {
	d.RLock()
	for _, user := range d.users {
		for _, id := range excludeUUIDs {
			if user.uuid == id {
				continue
			}
			user.Send(msg)
		}
	}
	d.RUnlock()
}

// Create creates a new user in the user DB given a username and connection.
func (d *UserDB) Create(username string, conn net.Conn) (User, error) {
	// validate username
	switch {
	case len(username) < 6:
		return User{}, errors.New("username must have a minimum length of 6 characters")
	case len(username) > 12:
		return User{}, errors.New("username length must not exceed 12 characters")
	}
	// TODO: validate username to ensure it only contains letters/numbers

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
		return User{}, err
	}

	// insert new user into DB
	newUUID := uuid.NewV4().String()
	newUser := User{
		name: username,
		conn: conn,
		uuid: newUUID,
	}
	d.Lock()
	d.users[newUUID] = newUser
	d.Unlock()
	return newUser, nil
}

// Connect associates an existing user in the user DB with a new connection.
func (d *UserDB) Connect(uuid string, conn net.Conn) (User, error) {
	d.RLock()
	user, ok := d.users[uuid]
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
	d.users[uuid] = user
	d.Unlock()
	return user, nil
}

// Disconnect clears the reference to a user's connection.
func (d *UserDB) Disconnect(user User) {
	// inform client that disconnect has been acknowledged
	user.Send(Message{
		Type: "disconnected",
	})

	// clear connection reference before updating DB
	user.conn = nil
	d.Lock()
	// set connection to nil
	d.users[user.uuid] = user
	d.Unlock()

	// broadcast user leaving message to all remaining connected users
	d.Broadcast(Message{
		Type:  "disconnect",
		Value: user.name + " has left the game!",
	}, user.uuid)
}
