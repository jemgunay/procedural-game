// Package server handles launching a TCP game server and the processing of incoming player requests.
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

var (
	listener net.Listener
	stopChan chan struct{}

	userDB       UserDB
	projectileDB ProjectileDB
	worldSeed    string
)

// Start starts the TCP server and polls for incoming TCP connections.
func Start(addr, seed string) error {
	worldSeed = seed
	stopChan = make(chan struct{}, 1)
	userDB = UserDB{
		users: make(map[string]User),
		rand:  rand.New(rand.NewSource(int64(time.Now().UTC().Nanosecond()))),
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
			// listen for an incoming connection
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-stopChan:
					// shutdown server
					return

				default:
					fmt.Printf("failed to accept connection: %s\n", err)
				}
			}
			// handle connection
			go handleConn(conn)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the TCP server.
func Shutdown() {
	fmt.Println("TCP server shutting down")
	userDB.Broadcast(Message{
		Type: "server_shutdown",
	})
	time.Sleep(time.Millisecond * 500)
	stopChan <- struct{}{}
	listener.Close()
	// TODO: wg.Wait()
}

// handles the processing and maintenance of a connection between the server and a single game client.
func handleConn(conn net.Conn) {
	defer conn.Close()

	// get client address
	addr := conn.RemoteAddr().String()
	fmt.Println("TCP client connection established on " + addr)

	var user User
	defer func() {
		// clean up on messy connection closure
		if user.conn != nil {
			userDB.Disconnect(user)
		}

		// client disconnecting
		fmt.Println("TCP client connection disconnected on " + addr)
	}()

	for {
		select {
		case <-user.exitCh:
			return
		default:
		}

		resp, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Printf("failed to read incoming TCP request: %s\n", err)
			return
		}

		// unmarshal raw request
		var msg Message
		if err := json.Unmarshal([]byte(resp), &msg); err != nil {
			fmt.Printf("invalid request received from %s, %s:\n%s\n", addr, err, []byte(resp))
			continue
		}

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

		case "vitals":
			// write pos msg right back to other clients
			user.vitals = msg.Value
			userDB.Update(user)
			msg.Value = user.name + "|" + msg.Value
			userDB.Broadcast(msg, user.name)

		case "create_projectile":
			data, err := msg.Unpack()
			if err != nil {
				fmt.Printf("create_projectile message incorrectly formatted: %s\n", err)
			}
			newProjectile := Projectile{
				owner: user.name,
				spawnTime: data.GetTime("spawnTime"),
				ttl: data.GetDuration("ttl"),
				startX: data.GetFloat("startX"),
				startY: data.GetFloat("startY"),
				velX: data.GetFloat("velX"),
				velY: data.GetFloat("velY"),
			}

			projectileDB.Create(newProjectile)
			userDB.Broadcast(msg, user.name)

		default:
			fmt.Printf("unsupported request type for connected stage: %s\n", msg.Type)
		}
	}
}

// handles registering (signing up) and reconnecting (logging in) users on an established connection, associating the
// connection with a user in the process
func establishUser(msg Message, conn net.Conn) (user User) {
	if msg.Type != "connect" {
		fmt.Println("unsupported request type for init stage: " + msg.Type)
		return user
	}

	var ok bool
	var err error

	user, ok = userDB.Get(msg.Value)
	// user does not exist yet - attempt to create new user given the provided username
	if !ok {
		user, err = userDB.Create(msg.Value, conn)
		if err != nil {
			user.Send(Message{
				Type:  "register_failure",
				Value: "failed to create user: " + err.Error(),
			})
			return
		}

		// respond with register success
		user.Send(Message{
			Type:  "register_success",
			Value: worldSeed + "|" + user.name + "|" + user.vitals,
		})
	} else {
		// attempt to establish connection for existing user
		user, err = userDB.Connect(msg.Value, conn)
		if err != nil {
			user.Send(Message{
				Type:  "connect_failure",
				Value: "failed to connect existing user: " + err.Error(),
			})
			return
		}

		// respond with connect success
		user.Send(Message{
			Type:  "connect_success",
			Value: worldSeed + "|" + user.name + "|" + user.vitals,
		})
	}

	// broadcast to all players that user successfully joined
	userDB.Broadcast(Message{
		Type:  "user_joined",
		Value: user.name,
	}, user.name)

	// send world update to player
	var data strings.Builder
	for _, u := range userDB.users {
		if u.name == user.name {
			continue
		}
		if data.String() != "" {
			data.WriteString("/")
		}
		data.WriteString(u.name + "|" + u.vitals)
	}
	if data.String() != "" {
		user.Send(Message{
			Type:  "init_world",
			Value: data.String(),
		})
	}

	return
}
