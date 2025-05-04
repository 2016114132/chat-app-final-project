// Handles client connections
package server

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

var (
	clients = make(map[net.Conn]string)
	mu      sync.Mutex
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()

	// Register client
	mu.Lock()
	clients[conn] = addr
	mu.Unlock()

	fmt.Printf("[+] %s connected\n", addr)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("[-] %s disconnected\n", addr)
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			return
		}

		broadcast := fmt.Sprintf("[%s]: %s", addr, message)
		broadcastMessage(conn, broadcast)
	}
}

func broadcastMessage(sender net.Conn, message string) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		if client != sender {
			_, err := client.Write([]byte(message))
			if err != nil {
				fmt.Println("Error broadcasting:", err)
			}
		}
	}
}
