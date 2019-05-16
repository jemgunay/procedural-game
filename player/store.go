package player

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jemgunay/procedural-game/file"
)

// Store is a player store which can be concurrently accessed safely.
type Store struct {
	players map[string]*Player
	sync.RWMutex
}

// NewStore creates and initialises a new player store.
func NewStore() *Store {
	return &Store{
		players: make(map[string]*Player),
	}
}

// Find returns the player which corresponds with the specified username.
func (s *Store) Find(username string) (*Player, error) {
	s.RLock()
	p, ok := s.players[username]
	s.RUnlock()
	if !ok {
		return nil, errors.New("player does not exist with username: " + username)
	}
	return p, nil
}

// Add takes a username, creates a new corresponding user and adds it to the player store.
func (s *Store) Add(username string) (*Player, error) {
	// ensure player does not already exist in store
	s.RLock()
	_, ok := s.players[username]
	s.RUnlock()
	if ok {
		return nil, errors.New("player already exists with username: " + username)
	}

	// create sprite
	sprite, err := file.CreateSprite(file.Player)
	if err != nil {
		return nil, err
	}

	newPlayer := &Player{
		name:        username,
		pos:         pixel.ZV,
		baseSpeed:   300.0,
		orientation: 0.0,
		sprite:      sprite,
	}

	// add new player to players map
	s.Lock()
	s.players[username] = newPlayer
	s.Unlock()
	return newPlayer, nil
}

// Remove removes a player from the player store.
func (s *Store) Remove(username string) {
	s.Lock()
	delete(s.players, username)
	s.Unlock()
}

// Draw draws each of the players in the player store.
func (s *Store) Draw(win *pixelgl.Window) {
	s.RLock()
	for _, p := range s.players {
		p.Draw(win)
	}
	s.RUnlock()
}

// String produces a string containing descriptions of all players.
func (s *Store) String() string {
	b := strings.Builder{}
	s.RLock()
	for _, p := range s.players {
		p.RLock()
		b.WriteString("> " + p.name + "\n")
		b.WriteString("\tpos: " + p.pos.String() + "\n")
		b.WriteString("\trot: " + fmt.Sprint(p.orientation) + "\n")
		b.WriteString("\tspeed: " + fmt.Sprint(p.baseSpeed) + "\n")
		p.RUnlock()
	}
	s.RUnlock()
	return b.String()
}
