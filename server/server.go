// Package server handles launching a TCP game server and the processing of incoming player requests.
package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

var (
	listener net.Listener
	stopChan chan struct{}

	userDB UserDB
)

// Start starts the TCP server and polls for incoming TCP connections.
func Start(addr string) error {
	stopChan = make(chan struct{})
	userDB = UserDB{
		users: make(map[string]User),
	}

	// bind TCP listener
	var err error
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind TCP on port %s: %s", addr, err)
	}

	fmt.Printf("TCP server listening on %s\n", listener.Addr())

	// main TCP server loop
	go func() {
		defer listener.Close()

		for {
			select {
			case <-stopChan:
				// shutdown server
				break

			default:
				// listen for an incoming connection
				conn, err := listener.Accept()
				if err != nil {
					fmt.Printf("failed to accept connection: %s\n", err)
					break
				}
				// handle connection
				go handleConn(conn)
			}
		}

		// TODO: broadcast shutdown to all users
		// TODO: add waitgroup to complete all connections before killing server so that shutdown messages can be sent to all clients
		fmt.Println("TCP server shut down")
	}()

	return nil
}

// Shutdown gracefully shuts down the TCP server.
func Shutdown() {
	stopChan <- struct{}{}
	// TODO: wg.Wait()
}

// Message represents an incoming request from a client or an outgoing request from the server.
type Message struct {
	Type  string `json:"type"`
	Value string `json:"val,omitempty"`
}

// handles the processing and maintenance of a connection between the server and a single game client.
func handleConn(conn net.Conn) {
	defer conn.Close()

	// get client address
	addr := conn.RemoteAddr().String()
	fmt.Println("TCP client connection established on " + addr)

	var user User

	for {
		resp, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Printf("failed to read incoming TCP request: %s\n", err)
			break
		}

		// unmarshal raw request
		var msg Message
		if err := json.Unmarshal([]byte(resp), &msg); err != nil {
			fmt.Printf("invalid request received from %s, %s:\n%s\n", addr, err, []byte(resp))
			continue
		}

		fmt.Printf("> [Server] Incoming TCP request:\n%v\n", msg)

		// require a successful register/connect before allowing access to other request instruction types
		if user.conn == nil {
			user = establishUser(msg, conn)
			continue
		}

		// execute operation on initialised user
		switch msg.Type {
		case "disconnect":
			// clear user's connection reference in the user DB
			userDB.Disconnect(user)
			// destroy reference to local user
			user = User{}

		case "pos":
			fmt.Printf("new pos: %s\n", msg.Value)
			// write pos msg right back to other clients
			msg.Value = msg.Value + "|" + user.name
			userDB.Broadcast(msg, user.uuid)

		default:
			fmt.Printf("unsupported request type for connected stage: %s\n", msg.Type)
		}
	}

	// clean up on messy connection closure
	if user.conn != nil {
		userDB.Disconnect(user)
	}

	// client disconnecting
	fmt.Println("TCP client connection disconnected on " + addr)
}

// handles registering (signing up) and reconnecting (logging in) users on an established connection, associating the
// connection with a user in the process
func establishUser(msg Message, conn net.Conn) (user User) {
	var err error

	switch msg.Type {
	case "register":
		// attempt to create new user given the provided username
		user, err = userDB.Create(msg.Value, conn)
		if err != nil {
			err = fmt.Errorf("failed to create user: %s", err)
			user.Send(Message{
				Type:  "register_failure",
				Value: err.Error(),
			})
			break
		}

		// respond with register success
		user.Send(Message{
			Type:  "register_success",
			Value: user.uuid,
		})

	case "connect":
		// attempt to establish connection for existing user
		user, err = userDB.Connect(msg.Value, conn)
		if err != nil {
			err = fmt.Errorf("failed to connect existing user: %s", err)
			break
		}

		// respond with connect success
		user.Send(Message{
			Type: "connect_success",
		})

	default:
		err = errors.New("unsupported request type for init stage: " + msg.Type)
	}

	// catch if any errors occurred
	if err != nil {
		fmt.Printf("failed to establish user connection: %s\n", err)
		// TODO: respond with error message
		return
	}

	// broadcast to all players that user successfully joined
	userDB.Broadcast(Message{
		Type:  "user_joined",
		Value: user.name,
	}, user.uuid)

	return
}
