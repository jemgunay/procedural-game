package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/jemgunay/procedural-game/server"
)

var (
	conn         net.Conn
	messageQueue chan server.Message
)

// Start initialises a connection with a TCP game server.
func Start(addr string) error {
	messageQueue = make(chan server.Message, 1024)

	var err error
	conn, err = net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind TCP on port %s: %s", addr, err)
	}

	fmt.Println("TCP server connection established on " + addr)

	go func() {
		defer conn.Close()
		// listen for reply
		r := bufio.NewReader(conn)
		for {
			resp, err := r.ReadString('\n')
			if err != nil {
				fmt.Printf("failed to read incoming TCP request: %s\n", err)
				break
			}

			// unmarshal raw request
			var msg server.Message
			if err := json.Unmarshal([]byte(resp), &msg); err != nil {
				fmt.Printf("invalid request received from %s, %s:\n%s\n", addr, err, []byte(resp))
				continue
			}

			messageQueue <- msg
		}

		fmt.Println("TCP server connection disconnected on " + addr)
	}()

	return nil
}

// Poll pulls a message from the queue and returns it to be processed by the scene. If there are no messages in
// the queue, an empty Message is returned.
func Poll() server.Message {
	select {
	case msg := <-messageQueue:
		return msg
	default:
		return server.Message{}
	}
}

// Send marshals and writes a message to a server.
func Send(msg server.Message) {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("failed to process outbound request: %s:\n%v\n", err, msg)
		return
	}

	if _, err := conn.Write(append(rawMsg, '\n')); err != nil {
		fmt.Printf("failed to write \"%s\" to %s: %s\n", string(rawMsg), conn.RemoteAddr(), err)
	}
}
