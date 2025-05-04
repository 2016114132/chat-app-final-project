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
	conn, err := net.Dial("tcp", "localhost:4000")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	// Prompt for nickname
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your nickname: ")
	nickname, _ := reader.ReadString('\n')
	nickname = shared.SanitizeInput(nickname)
	if nickname == "" {
		nickname = "Anonymous"
	}
	// Send nickname to server as the first message
	conn.Write([]byte("/name " + nickname + "\n"))

	fmt.Println("Connected to chat server as", nickname)

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
	for {
		fmt.Print("You: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
		text = shared.SanitizeInput(text)
		if text == "" {
			continue
		}
		conn.Write([]byte(text + "\n"))
	}
}
