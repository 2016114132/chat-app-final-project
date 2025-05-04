// Handles client connections
package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	clients = make(map[net.Conn]string)
	mu      sync.Mutex
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	name := addr // default name

	reader := bufio.NewReader(conn)

	// First line: expect nickname command
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Client disconnected before sending name.")
		return
	}
	if strings.HasPrefix(line, "/name ") {
		name = strings.TrimSpace(strings.TrimPrefix(line, "/name "))
	}

	// Register client
	mu.Lock()
	clients[conn] = name
	mu.Unlock()

	fmt.Printf("[+] %s connected\n", name)

	// reader = bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("[-] %s disconnected\n", name)
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			return
		}

		broadcast := fmt.Sprintf("[%s]: %s", name, strings.TrimSpace(message))
		broadcastMessage(conn, broadcast)
	}
}

func broadcastMessage(sender net.Conn, message string) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		if client != sender {
			// client.Write([]byte(message + "\n"))
			_, err := client.Write([]byte(message + "\n"))
			if err != nil {
				fmt.Println("Error broadcasting:", err)
			}
		}
	}
}
