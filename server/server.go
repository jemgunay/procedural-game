// Package server handles launching a TCP game server and the processing of incoming player requests.
package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
)

var (
	tcpPort  string
	stopChan = make(chan struct{})
	userDB   = UserDB{
		users: make(map[string]User),
	}
)

// Start starts the TCP server and polls for incoming TCP connections.
func Start(port uint64) error {
	tcpPort = strconv.FormatUint(port, 10)

	// bind TCP listener
	listener, err := net.Listen("tcp", ":"+tcpPort)
	if err != nil {
		return fmt.Errorf("failed to bind TCP on port %s: %s\n", tcpPort, err)
	}
	defer listener.Close()

	fmt.Printf("TCP server listening on %s\n", listener.Addr())

	// main TCP server loop
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

	return nil
}

// Shutdown gracefully shuts down the TCP server.
func Shutdown() {
	stopChan <- struct{}{}
	// TODO: wg.Wait()
}

// Request represents an incoming request from a client or an outgoing request from the server.
type Request struct {
	Type string `json:"type"`
	Msg  string `json:"msg,omitempty"`
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// get client address
	addr := conn.RemoteAddr().String()
	fmt.Println("TCP client connection established: " + addr)

	var (
		user User
		err  error
	)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		// unmarshal raw request
		var req Request
		if err = json.Unmarshal(scanner.Bytes(), &req); err != nil {
			fmt.Printf("invalid request received from %s, %s:\n%s\n", addr, err, scanner.Text())
			continue
		}

		// require a successful register/connect before allowing access to other request instruction types
		if user.uuid == "" {
			user = establishUser(req, conn)
			continue
		}

		// execute operation on initialised user
		switch req.Type {
		case "disconnect":
			// close user connection
			userDB.Disconnect(user)
			// inform client that disconnect has been acknowledged
			user.Send(Request{
				Type: "disconnected",
			})
			user = User{}
			// broadcast user leaving message to all room users
			userDB.Broadcast(Request{
				Type: "disconnect",
				Msg:  user.name + " has left the game!",
			}, user.uuid)

		case "pos":
			fmt.Printf("new pos: %s\n", req.Msg)

		default:
			fmt.Printf("unsupported request type for connected stage: %s\n", req.Type)
		}

	}

	// client disconnecting
	fmt.Println("TCP client connection dropped: " + addr)
}

// handles registering (signing up) and reconnecting (logging in) users on an established connection, associating the
// connection with a user in the process
func establishUser(req Request, conn net.Conn) (user User) {
	var err error

	switch req.Type {
	case "register":
		// attempt to create new user given the provided username
		user, err = userDB.Create(req.Msg, conn)
		if err != nil {
			err = fmt.Errorf("failed to create user: %s", err)
			break
		}

		// respond with register success
		user.Send(Request{
			Type: "register_success",
			Msg:  user.uuid,
		})

	case "connect":
		// attempt to establish connection for existing user
		user, err = userDB.Connect(req.Msg, conn)
		if err != nil {
			err = fmt.Errorf("failed to connect existing user: %s", err)
			break
		}

		// respond with connect success
		user.Send(Request{
			Type: "connect_success",
		})

	default:
		err = errors.New("unsupported request type for init stage: " + req.Type)
	}

	// catch if any errors occurred
	if err != nil {
		fmt.Printf("failed to establish user connection: %s\n", err)
		// TODO: respond with error message
		return
	}

	// broadcast to all players that user successfully joined
	userDB.Broadcast(Request{
		Type: "announcement",
		Msg:  user.name + " has joined the game!",
	}, user.uuid)

	return
}
