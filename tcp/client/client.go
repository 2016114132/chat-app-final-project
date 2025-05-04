// Starts TCP client
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:4000")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to chat server.")

	// Graceful exit on Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nDisconnected from chat.")
		conn.Close()
		os.Exit(0)
	}()

	// Read messages from server
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			// Clear the current line and print incoming message
			fmt.Printf("\r%s\n", scanner.Text())
			fmt.Print("You: ")
		}
	}()

	// Read input from user and send to server
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("You: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		conn.Write([]byte(text + "\n"))
	}
}
