// Handles client connections
package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/2016114132/chat-app-final-project/shared"
)

// clients is a concurrent-safe map of all connected clients and their nicknames.
// The key is the TCP connection, and the value is the client's name.
var (
	clients = make(map[net.Conn]string)
	mu      sync.Mutex // Used to prevent race conditions when accessing the clients map
)

// handleConnection is called whenever a new client connects.
// It reads the client's nickname, listens for messages, and broadcasts them to other clients.
func handleConnection(conn net.Conn) {
	// Ensure connection is closed when done
	defer conn.Close()

	// Get the client's network address as a default identifier
	addr := conn.RemoteAddr().String()

	// Default name is the client's address
	name := addr

	// Create a buffered reader to read incoming data
	reader := bufio.NewReader(conn)

	// First message should be the nickname using the "/name" command
	// Read the first line sent by the client
	line, err := reader.ReadString('\n')
	if err != nil {
		// Exit if client drops before naming
		fmt.Println("Client disconnected before sending name.")
		return
	}

	// If the line starts with "/name"
	if shared.IsCommand(line, "name") {
		// Extract and clean the nickname
		name = shared.FormatName(strings.TrimPrefix(line, "/name "))
	}

	// Store the new client connection and nickname
	mu.Lock()
	clients[conn] = name
	mu.Unlock()

	// Log the connection to the server console
	fmt.Printf("[+] %s connected\n", name)

	// Listen for messages from this client
	for {
		// Read full message until newline
		message, err := reader.ReadString('\n')
		if err != nil {
			// If an error occurs (likely disconnect), remove the client
			fmt.Printf("[-] %s disconnected\n", name)
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			return
		}

		// Format the message and send to all other connected clients
		broadcast := shared.FormatMessage(name, message)
		broadcastMessage(conn, broadcast)
	}
}

// broadcastMessage sends a given message to all connected clients except the sender.
func broadcastMessage(sender net.Conn, message string) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		if client != sender {
			// Send the message
			_, err := client.Write([]byte(message + "\n"))

			// Log any issues
			if err != nil {
				fmt.Println("Error broadcasting:", err)
			}
		}
	}
}
