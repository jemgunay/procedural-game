package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/jemgunay/game/server"
)

var (
	tcpPort string
	conn    net.Conn
)

// Start initialises a connection with a TCP game server.
func Start(port uint64) error {
	tcpPort = strconv.FormatUint(port, 10)

	var err error
	conn, err = net.Dial("tcp", "localhost:"+tcpPort)
	if err != nil {
		return fmt.Errorf("failed to bind TCP on port %s: %s", tcpPort, err)
	}

	fmt.Printf("connecting to TCP server on %s\n", conn.RemoteAddr())

	/*go func() {
		for {
			resp, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Printf("error reading from connection: %s\n", err)
			}

			fmt.Println("> resp:\n" + resp)
		}
	}()*/

	go func() {
		defer conn.Close()
		// listen for reply
		for {
			resp, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Printf("faield to read incoming TCP request: %s\n", err)
				break
			}
			fmt.Println("> incoming: " + resp)
		}

		fmt.Println("connection to server closed")
	}()

	return nil
}

func Send(msg server.Message) {
	fmt.Println("starting send operation")
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("failed to process outbound request: %s:\n%v\n", err, msg)
		return
	}

	if _, err := conn.Write(append(rawMsg, '\n')); err != nil {
		fmt.Printf("failed to write \"%s\" to %s: %s\n", string(rawMsg), conn.RemoteAddr(), err)
	}
	fmt.Println("ending send operation")
}
