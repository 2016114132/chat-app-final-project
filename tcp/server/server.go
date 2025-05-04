// Starts TCP server
package server

import (
	"fmt"
	"net"
)

// This function initializes the TCP server on the specified address
// It listens for incoming connections and spawns a goroutine for each client.
func Start(address string) {
	// Create a TCP listener on the given address
	listener, err := net.Listen("tcp", address)
	if err != nil {
		// If the server fails to start, panic and exit
		panic(err)
	}
	// Ensure listener is closed on exit
	defer listener.Close()

	fmt.Println("Server listening on", address)

	// Main server loop: continuously accept new client connections
	for {
		// Wait for and accept a new incoming connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue // If error occurs, skip and try again
		}

		// Launch a goroutine to handle this client separately (non-blocking)
		go handleConnection(conn)
	}
}
