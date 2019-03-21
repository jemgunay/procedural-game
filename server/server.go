// Package server handles launching a TCP game server and the processing of incoming player requests.
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

var (
	tcpPort     string
	requestPool = make(chan Request)
	stopChan    = make(chan struct{})
	userDB      = UserDB{
		users: make(map[string]User),
	}
)

// Start starts the TCP server and polls for incoming TCP connections.
func Start(port uint64) error {
	tcpPort = strconv.FormatUint(port, 10)

	// bind TCP listener
	listener, err := net.Listen("TCP", ":"+tcpPort)
	if err != nil {
		return fmt.Errorf("failed to bind TCP on port %s: %s", tcpPort, err)
	}
	defer listener.Close()

	fmt.Printf("TCP server listening on %s", listener.Addr())

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
				fmt.Printf("failed to accept connection: %s", err)
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
	Type string   `json:"type"`
	Msg  string   `json:"msg"`
	conn net.Conn `json:"-"`
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// get client address
	addr := conn.RemoteAddr().String()
	fmt.Println("TCP client connection established: " + addr)

	var user User

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		// unmarshal raw request
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			fmt.Printf("invalid request received from %s, %s:\n%s", addr, err, scanner.Text())
			continue
		}

		// execute operation
		switch req.Type {
		case "register":
			// extract and validate new username
			user, err := userDB.Create(req.Msg, req.conn)
			if err != nil {

			}

		case "connect":
			// validate uuid provided
			if err := userDB.Connect(req.Msg, req.conn); err != nil {

			}

		case "disconnect":

		case "pos":

		default:

		}

	}

	// client disconnecting
	fmt.Println("TCP client connection dropped: " + addr)
	// broadcast user leaving message to all room users
	requestPool <- Request{Type: "disconnect", conn: conn}
}
