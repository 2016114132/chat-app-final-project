// Starts TCP client
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/2016114132/chat-app-final-project/shared"
)

func main() {
	// Attempt to connect to the TCP chat server running on localhost at port 4000
	conn, err := net.Dial("tcp", "localhost:4000")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}

	// Ensure connection is closed properly when the client exits
	defer conn.Close()

	// Prompt the user for their nickname to display in the chat
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your nickname: ")
	nickname, _ := reader.ReadString('\n')
	nickname = shared.FormatName(nickname)

	// Send nickname to server using a custom "/name" command protocol
	conn.Write([]byte("/name " + nickname + "\n"))

	// Inform the user that they are successfully connected
	fmt.Println("Connected to chat server as", nickname)

	// Set up a signal listener to gracefully handle Ctrl+C (SIGINT) or SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig // Wait until signal is received
		fmt.Println("\nDisconnected from chat.")
		conn.Close() // Ensure connection is closed
		os.Exit(0)   // Exit the program
	}()

	// Start a goroutine to continuously read and display messages from the server
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			// Use carriage return (\r) to overwrite the current prompt line
			fmt.Printf("\r%s\n", scanner.Text()) // Display incoming message
			fmt.Print("You: ")                   // Reprint input prompt
		}
	}()

	// Main loop: read user input from terminal and send it to the server
	for {
		// Show prompt
		fmt.Print("You: ")

		// Read message from user
		text, err := reader.ReadString('\n')

		// Handle input errors
		if err != nil {
			fmt.Println("Error:", err)
			break
		}

		// Clean the input by triming whitespace
		text = shared.SanitizeInput(text)

		// Skip sending if the message is empty
		if text == "" {
			continue
		}

		// Send the message to the server, with newline to match protocol
		_, err = conn.Write([]byte(text + "\n"))
		if err != nil {
			// If connection is lost, notify user and exit
			fmt.Println("Server unavailable. Could not send message.")
			break
		}
	}
}
