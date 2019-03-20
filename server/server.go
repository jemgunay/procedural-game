// Package server handles launching a TCP game server and the processing of incoming player requests.
package server

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

type Request struct {
	requestPool chan Request
}

var (
	tcpPort     string
	requestChan chan Request
	stopChan    chan struct{}
)

// Start starts
func Start(port uint64) error {
	tcpPort = strconv.FormatUint(port, 10)

	// bind TCP listener
	listener, err := net.Listen("TCP", ":"+tcpPort)
	if err != nil {
		return fmt.Errorf("failed to listen for TCP on tcpPort %s %s", err)
	}
	defer listener.Close()

	fmt.Println("TCP server listening on port " + tcpPort)

	// main TCP server loop
	for {
		// listen for an incoming connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		// handle connection
		go handleRequest(conn)
	}
	return nil
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	input := bufio.NewScanner(conn)
	for input.Scan() {
		// unmarshal client string request into Message object
		/*msg := Message{}
		msg.unmarshalRequest(input.Text())
		clientUUID = msg.TargetUUID

		// produce response based on request
		requestChan <- Request{&msg, ch}*/
	}

	// client disconnecting
	/*fmt.Println(clientAddress + " TCP client connection dropped")
	// broadcast user leaving message to all room users
	exitMsg := Message{TargetUUID: clientUUID, Type: "exit"}
	requestPool <- Request{msg: &exitMsg}*/
}
