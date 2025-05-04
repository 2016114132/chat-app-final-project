// Handles client connections
package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/2016114132/chat-app-final-project/shared"
)

var (
	clients     = make(map[string]*net.UDPAddr)
	names       = make(map[string]string)
	lastSeen    = make(map[string]time.Time)
	mu          sync.Mutex
	pingTimeout = 6 * time.Second
)

func handleConnection(conn *net.UDPConn) {
	buffer := make([]byte, 1024)

	// Background goroutine to remove inactive clients
	go func() {
		for {
			time.Sleep(2 * time.Second)
			now := time.Now()

			mu.Lock()
			for key, last := range lastSeen {
				if now.Sub(last) > pingTimeout {
					fmt.Printf("[-] %s disconnected\n", names[key])
					delete(clients, key)
					delete(names, key)
					delete(lastSeen, key)
				}
			}
			mu.Unlock()
		}
	}()

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue
		}

		message := string(buffer[:n])
		addrKey := addr.String()

		mu.Lock()
		lastSeen[addrKey] = time.Now()

		// First-time connection
		if _, ok := clients[addrKey]; !ok {
			clients[addrKey] = addr

			// Check if first message is nickname
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

		// Handle ping
		if shared.IsCommand(message, "ping") {
			mu.Unlock()
			continue
		}

		// Handle name update
		if shared.IsCommand(message, "name") {
			name := shared.FormatName(strings.TrimPrefix(message, "/name "))
			names[addrKey] = name
			mu.Unlock()
			continue
		}

		name := names[addrKey]
		mu.Unlock()

		broadcast := shared.FormatMessage(name, message)
		broadcastMessage(conn, addrKey, broadcast)
	}
}

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
