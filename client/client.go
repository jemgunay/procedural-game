// Package client handles retrieving and sending messages from and to a game server instance.
package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/jemgunay/procedural-game/server"
	"github.com/pkg/errors"
)

const (
	maxSendFails           = uint(10)
	messageQueueBufferSize = 2048
)

var (
	conn         net.Conn
	messageQueue chan server.Message
	connected    bool

	sendFailCounter uint
	stopChan        chan struct{}
)

// Start initialises a connection with a TCP game server.
func Start(addr string) error {
	stopChan = make(chan struct{}, 1)
	messageQueue = make(chan server.Message, messageQueueBufferSize)
	sendFailCounter = 0

	var err error
	conn, err = net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind TCP on port %s: %s", addr, err)
	}

	connected = true
	fmt.Println("TCP server connection established on " + addr)

	go func() {
		defer conn.Close()
		// listen for reply
		r := bufio.NewReader(conn)
		for {
			select {
			case <-stopChan:
				fmt.Println("TCP client connection disconnected on " + addr)
				close(messageQueue)
				return

			default:
				resp, err := r.ReadString('\n')
				if err != nil {
					fmt.Printf("failed to read incoming TCP request: %s\n", err)
					Disconnect()
					continue
				}

				// unmarshal raw request
				var msg server.Message
				if err := json.Unmarshal([]byte(resp), &msg); err != nil {
					fmt.Printf("invalid request received from %s, %s:\n%s\n", addr, err, []byte(resp))
					continue
				}

				messageQueue <- msg
			}
		}
	}()

	return nil
}

var (
	// ErrQueueClosed indicates that a message queue has been closed and that no more messages will be provided by it.
	ErrQueueClosed = errors.New("message queue closed")
	// ErrQueueEmpty indicates that there are currently no messages in the message queue.
	ErrQueueEmpty = errors.New("message queue is empty")
)

// Poll pulls a message from the queue and returns it to be processed by the scene. If there are no messages in
// the queue, an empty Message is returned.
func Poll() (server.Message, error) {
	select {
	case msg, ok := <-messageQueue:
		if ok {
			return msg, nil
		}
		return server.Message{}, ErrQueueClosed
	default:
		return server.Message{}, ErrQueueEmpty
	}
}

// Send marshals and writes a message to a server.
func Send(msg server.Message) {
	if _, err := conn.Write(msg.Pack()); err != nil {
		fmt.Printf("failed to the following to %s: %s:\n%+v\n", conn.RemoteAddr(), err, msg)

		// if too many write fails occur in a row, then disconnect from server
		sendFailCounter++
		if sendFailCounter >= maxSendFails {
			fmt.Println("too many failed sends occurred - triggering a client disconnect")
			Disconnect()
		}
		return
	}
	sendFailCounter = 0
}

// Disconnect disconnects the client from the server.
func Disconnect() {
	if !connected {
		return
	}
	fmt.Println("disconnecting client")
	stopChan <- struct{}{}
}
