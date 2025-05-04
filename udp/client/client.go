// Starts UDP client
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/2016114132/chat-app-final-project/shared"
)

func main() {
	// Resolve the server address and port for UDP
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:4001")
	if err != nil {
		fmt.Println("Resolve error:", err)
		return
	}

	// Establish a UDP connection to the server
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	// Prompt the user to enter their nickname
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your nickname: ")
	nickname, _ := reader.ReadString('\n')
	nickname = shared.FormatName(nickname)

	// Send the nickname to the server using a custom "/name" command
	conn.Write([]byte("/name " + nickname + "\n"))

	fmt.Println("Connected to UDP chat as", nickname)

	// Handle graceful exit when Ctrl+C or kill signal is received
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nDisconnected.")
		conn.Close()
		os.Exit(0)
	}()

	// Start a goroutine that sends periodic "/ping" messages to the server
	// This acts as a heartbeat to let the server know the client is still active
	go func() {
		for {
			time.Sleep(2 * time.Second)
			conn.Write([]byte("/ping\n"))
		}
	}()

	// Start a goroutine to continuously read incoming messages from the server
	go func() {
		// Create a buffer for incoming data
		buffer := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("\rRead error: %s\n", err)
				os.Exit(1)
			}
			// Print the received message, trimming any trailing whitespace
			fmt.Printf("\r%s\n", strings.TrimSpace(string(buffer[:n])))
			fmt.Print("You: ")
		}
	}()

	// Main send loop: reads user input and sends it to the server
	for {
		fmt.Print("You: ")

		// Read a line from the user
		text, _ := reader.ReadString('\n')

		// Clean up the input
		text = shared.SanitizeInput(text)

		// Skip empty messages
		if text == "" {
			continue
		}

		// Send the message
		_, err := conn.Write([]byte(text + "\n"))

		// Display an error if the server is unreachable
		if err != nil {
			fmt.Println("Server unavailable. Could not send message.")
			continue
		}
	}
}
