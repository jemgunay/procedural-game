package server

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/pixel"
)

// Message represents an incoming request from a client or an outgoing request from the server.
type Message struct {
	Type  string `json:"t"`
	Value string `json:"v"`
}

// Pack marshals a Message into a string for transmitting.
func (m Message) Pack() []byte {
	buf := bytes.Buffer{}
	buf.WriteString(`{"t":"` + m.Type + `","v":"` + m.Value + `"}`)
	// terminate with newline so clients can determine end of message
	buf.WriteByte('\n')
	return buf.Bytes()
}

type UnpackedMessage map[string]interface{}

func (m UnpackedMessage) Get(s string) interface{} {
	return m[s]
}

func (m UnpackedMessage) GetString(s string) string {
	return m[s].(string)
}

func (m UnpackedMessage) GetFloat(s string) float64 {
	return m[s].(float64)
}

func (m UnpackedMessage) GetInt(s string) int64 {
	return m[s].(int64)
}

func (m UnpackedMessage) GetUInt(s string) uint64 {
	return m[s].(uint64)
}

func (m UnpackedMessage) GetTime(s string) time.Time {
	return m[s].(time.Time)
}

func (m UnpackedMessage) GetDuration(s string) time.Duration {
	return m[s].(time.Duration)
}

// Unpack produces a map of validated and processed data based on a message's type. The map's values must be type
// asserted to extract the typed data.
func (m Message) Unpack() (UnpackedMessage, error) {
	components := strings.Split(m.Value, "|")

	switch m.Type {
	case "vitals":
		// validation
		return UnpackVitals(components)

	case "register_success", "connect_success":
		// validation
		if len(components) != 6 {
			return nil, errors.New("incorrect register_success component count")
		}

		unpacked, err := UnpackVitals(components[1:])
		if err != nil {
			return unpacked, err
		}

		unpacked["seed"] = components[0]
		return unpacked, nil

	case "create_projectile":
		if len(components) != 6 {
			return nil, errors.New("incorrect create_projectile component count")
		}
		spawnTime, err := strconv.ParseInt(components[0], 10, 64)
		if err != nil {
			return nil, errors.New("failed to parse health")
		}
		startX, err := strconv.ParseFloat(components[1], 64)
		if err != nil {
			return nil, errors.New("failed to parse startX")
		}
		startY, err := strconv.ParseFloat(components[2], 64)
		if err != nil {
			return nil, errors.New("failed to parse startY")
		}
		ttl, err := time.ParseDuration(components[3])
		if err != nil {
			return nil, errors.New("failed to parse ttl")
		}
		velX, err := strconv.ParseFloat(components[4], 64)
		if err != nil {
			return nil, errors.New("failed to parse velX")
		}
		velY, err := strconv.ParseFloat(components[5], 64)
		if err != nil {
			return nil, errors.New("failed to parse velY")
		}

		// unpacked response
		return UnpackedMessage{
			"spawnTime": time.Unix(0, spawnTime),
			"startX":    startX,
			"startY":    startY,
			"ttl":       ttl,
			"velX":      velX,
			"velY":      velY,
		}, nil
	}

	return nil, fmt.Errorf("unsupported message type supplied: %s", m.Type)
}

// UnpackVitals unpacks a message containing a player's vitals update.
func UnpackVitals(components []string) (UnpackedMessage, error) {
	if len(components) != 5 {
		return nil, errors.New("incorrect vitals component count")
	}
	x, err := strconv.ParseFloat(components[1], 64)
	if err != nil {
		return nil, errors.New("failed to parse X")
	}
	y, err := strconv.ParseFloat(components[2], 64)
	if err != nil {
		return nil, errors.New("failed to parse Y")
	}
	rot, err := strconv.ParseFloat(components[3], 64)
	if err != nil {
		return nil, errors.New("failed to parse rot")
	}
	health, err := strconv.ParseUint(components[4], 10, 64)
	if err != nil {
		return nil, errors.New("failed to parse health")
	}
	// unpacked response
	return UnpackedMessage{
		"name":   components[0],
		"pos":    pixel.V(x, y),
		"rot":    rot,
		"health": health,
	}, nil
}

// ConcatVitals packs a player's vitals update into a message.
func ConcatVitals(x, y, rot float64, health uint64) string {
	return fmt.Sprintf("%f|%f|%f|%d", x, y, rot, health)
}
