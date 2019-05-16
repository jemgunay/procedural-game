package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/faiface/pixel"
)

// Message represents an incoming request from a client or an outgoing request from the server.
type Message struct {
	Type  string `json:"type"`
	Value string `json:"val,omitempty"`
}

// Unpack produces a map of validated and processed data based on a message's type. The map's values must be type
// asserted to extract the typed data.
func (m *Message) Unpack() (map[string]interface{}, error) {
	components := strings.Split(m.Value, "|")

	switch m.Type {
	case "pos":
		// validation
		if len(components) != 4 {
			return nil, errors.New("incorrect pos component count")
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
		// unpacked response
		return map[string]interface{}{
			"name": components[0],
			"pos":  pixel.V(x, y),
			"rot":  rot,
		}, nil

	case "register_success", "connect_success":
		// validation
		if len(components) != 5 {
			return nil, errors.New("incorrect register_success component count")
		}

		x, err := strconv.ParseFloat(components[2], 64)
		if err != nil {
			return nil, errors.New("failed to parse X")
		}
		y, err := strconv.ParseFloat(components[3], 64)
		if err != nil {
			return nil, errors.New("failed to parse Y")
		}
		rot, err := strconv.ParseFloat(components[4], 64)
		if err != nil {
			return nil, errors.New("failed to parse rot")
		}

		// unpacked response
		return map[string]interface{}{
			"username": components[0],
			"seed":     components[1],
			"pos":      pixel.V(x, y),
			"rot":      rot,
		}, nil
	}

	return nil, fmt.Errorf("unsupported message type supplied: %s", m.Type)
}
