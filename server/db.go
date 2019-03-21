package server

import (
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
}

// UserDB is a database of users.
type UserDB struct {
	users map[string]User
	sync.RWMutex
}

func (d *UserDB) Get(uuid string) User {
	d.RLock()
	user := d.users[uuid]
	d.RUnlock()
	return user
}

func (d *UserDB) Update(uuid string, user User) {
	d.Lock()
	d.users[uuid] = user
	d.Unlock()
}

func (d *UserDB) Broadcast(message []byte, excludeUUIDs ...string) {
	d.Lock()
	for _, u := range excludeUUIDs {
		conn := d.users[u].conn
		if _, err := conn.Write(message); err != nil {
			fmt.Printf("failed to write \"%s\" to %s: %s", message, conn.RemoteAddr(), err)
		}
	}
	d.Unlock()
}

func (d *UserDB) Create(username string, conn net.Conn) (User, error) {
	// validate username
	switch {
	case len(username) > 5:
		return User{}, errors.New("username must have a minimum length of 6 characters")
	case len(username) < 13:
		return User{}, errors.New("username must have a maximum length of 12 characters")
	}
	// TODO: ensure username only contains letters/numbers

	d.RLock()
	for _, user := range d.users {
		if user.name == username {
			return User{}, errors.New("username already taken")
		}
	}
	d.RUnlock()

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

func (d *UserDB) Connect(uuid string, conn net.Conn) error {
	d.Lock()
	if _, ok := d.users[uuid]; !ok {
		return errors.New("user not found in DB")
	}
	// update connection
	d.users[uuid].conn = conn

	d.Unlock()
	return nil
}
