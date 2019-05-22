package server

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

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

// Unpack produces a map of validated and processed data based on a message's type. The map's values must be type
// asserted to extract the typed data.
func (m Message) Unpack() (map[string]interface{}, error) {
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
	}

	return nil, fmt.Errorf("unsupported message type supplied: %s", m.Type)
}

// UnpackVitals unpacks a message containing a player's vitals update.
func UnpackVitals(components []string) (map[string]interface{}, error) {
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
	return map[string]interface{}{
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
