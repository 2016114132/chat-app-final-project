// Starts UDP server
package server

import (
	"fmt"
	"net"
)

// Start initializes the UDP server on the given address
// It creates a UDP listener and passes it to the handler function
func Start(address string) {
	// Resolve the address string (like "localhost:4001") into a UDP address structure
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		// If the address is invalid or cannot be resolved, panic and exit
		panic(err)
	}

	// Listen for incoming UDP packets on the resolved address
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		// Panic if the server fails to start
		panic(err)
	}
	// Ensure the UDP connection is closed when done
	defer conn.Close()

	fmt.Println("UDP Server listening on", address)

	// Start processing incoming UDP messages using the handler
	handleConnection(conn)
}
