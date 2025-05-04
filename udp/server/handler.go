// Handles client connections for the UDP chat server
package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/2016114132/chat-app-final-project/shared"
)

// Global variables shared across all UDP clients
var (
	clients     = make(map[string]*net.UDPAddr)
	names       = make(map[string]string)
	lastSeen    = make(map[string]time.Time)
	mu          sync.Mutex
	pingTimeout = 6 * time.Second
)

// handleConnection listens for messages from clients, processes pings/names/messages,
// and launches a background cleanup routine for inactive clients.
func handleConnection(conn *net.UDPConn) {
	// Buffer to hold incoming UDP data
	buffer := make([]byte, 1024)

	// Background goroutine that runs every 2 seconds to clean up inactive clients
	go func() {
		for {
			// Check interval
			time.Sleep(2 * time.Second)
			now := time.Now()

			mu.Lock()
			for key, last := range lastSeen {
				if now.Sub(last) > pingTimeout {
					// Log disconnection
					fmt.Printf("[-] %s disconnected\n", names[key])
					delete(clients, key)
					delete(names, key)
					delete(lastSeen, key)
				}
			}
			mu.Unlock()
		}
	}()

	// Main server loop: wait for messages and respond
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue // Skip and wait for next message
		}

		// Convert bytes to string
		message := string(buffer[:n])

		// Use IP:port as the client's unique ID
		addrKey := addr.String()

		mu.Lock()

		// Update last seen timestamp for ping detection
		lastSeen[addrKey] = time.Now()

		// If this is a new client, register them
		if _, ok := clients[addrKey]; !ok {
			clients[addrKey] = addr

			// If the first message is a nickname, store it
			if shared.IsCommand(message, "name") {
				name := shared.FormatName(strings.TrimPrefix(message, "/name "))
				names[addrKey] = name
				fmt.Printf("[+] %s connected\n", name)
			} else {
				names[addrKey] = "Unknown"
				fmt.Printf("[+] Unknown client (%s) connected\n", addrKey)
			}
			mu.Unlock()
			continue
		}

		// Handle incoming ping messages to keep the client alive
		if shared.IsCommand(message, "ping") {
			mu.Unlock()
			continue
		}

		// Handle name update from an existing client
		if shared.IsCommand(message, "name") {
			name := shared.FormatName(strings.TrimPrefix(message, "/name "))
			names[addrKey] = name
			mu.Unlock()
			continue
		}

		// Retrieve the client's nickname for broadcasting
		name := names[addrKey]
		mu.Unlock()

		// Format and broadcast the chat message to all other clients
		broadcast := shared.FormatMessage(name, message)
		broadcastMessage(conn, addrKey, broadcast)
	}
}

// broadcastMessage sends the message to all clients except the one who sent it
func broadcastMessage(conn *net.UDPConn, sender string, message string) {
	mu.Lock()
	defer mu.Unlock()

	for key, client := range clients {
		if key != sender {
			_, err := conn.WriteToUDP([]byte(message), client)
			if err != nil {
				fmt.Println("Error broadcasting to", client.String(), ":", err)
			}
		}
	}
}
