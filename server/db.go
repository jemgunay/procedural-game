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
}

func (u *User) Send(req Request) {
	if u.conn == nil {
		return
	}

	rawMsg, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("failed to process outbound request: %s:\n%v\n", err, req)
		return
	}

	if _, err := u.conn.Write(rawMsg); err != nil {
		fmt.Printf("failed to write \"%s\" to %s: %s\n", string(rawMsg), u.conn.RemoteAddr(), err)
	}
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

func (d *UserDB) Update(user User) {
	d.Lock()
	d.users[user.uuid] = user
	d.Unlock()
}

func (d *UserDB) Broadcast(resp Request, excludeUUIDs ...string) {
	d.RLock()
	for _, user := range d.users {
		for _, id := range excludeUUIDs {
			if user.uuid == id {
				continue
			}
			user.Send(resp)
		}
	}
	d.RUnlock()
}

func (d *UserDB) Create(username string, conn net.Conn) (User, error) {
	// validate username
	switch {
	case len(username) < 6:
		return User{}, errors.New("username must have a minimum length of 6 characters")
	case len(username) > 12:
		return User{}, errors.New("username length must not exceed 12 characters")
	}
	// TODO: ensure username only contains letters/numbers

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

func (d *UserDB) Connect(uuid string, conn net.Conn) (User, error) {
	d.RLock()
	user, ok := d.users[uuid]
	d.RUnlock()
	if !ok {
		return User{}, errors.New("user not found in DB")
	}

	// check user is not already connected to prevent kicking off different client
	if conn != nil {
		return User{}, errors.New("user already connected")
	}

	// update connection
	user.conn = conn
	d.Lock()
	d.users[uuid] = user
	d.Unlock()
	return user, nil
}

func (d *UserDB) Disconnect(user User) {
	d.Lock()
	// set connection to nil
	user.conn = nil
	d.users[user.uuid] = user
	d.Unlock()
}
