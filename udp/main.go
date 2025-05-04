// Entry point for the UDP chat server
package main

// Import the custom UDP server package
import "github.com/2016114132/chat-app-final-project/udp/server"

// main is the starting point of the UDP server application.
// It launches the server on port 4001 and begins listening for messages.
func main() {
	// Start the UDP server on localhost at port 4001
	server.Start(":4001")
}
