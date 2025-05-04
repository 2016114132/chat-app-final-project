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
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:4001")
	if err != nil {
		fmt.Println("Resolve error:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	// Prompt for nickname
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your nickname: ")
	nickname, _ := reader.ReadString('\n')
	nickname = shared.FormatName(nickname)
	conn.Write([]byte("/name " + nickname + "\n"))

	fmt.Println("Connected to UDP chat as", nickname)

	// Graceful exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nDisconnected.")
		conn.Close()
		os.Exit(0)
	}()

	// Start ping loop
	go func() {
		for {
			time.Sleep(2 * time.Second)
			conn.Write([]byte("/ping\n"))
		}
	}()

	// Read from server
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				// fmt.Println("Read error:", err)
				fmt.Printf("\rRead error: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("\r%s\n", strings.TrimSpace(string(buffer[:n])))
			fmt.Print("You: ")
		}
	}()

	// Send loop
	for {
		fmt.Print("You: ")
		text, _ := reader.ReadString('\n')
		text = shared.SanitizeInput(text)
		if text == "" {
			continue
		}
		// conn.Write([]byte(text + "\n"))
		_, err := conn.Write([]byte(text + "\n"))
		if err != nil {
			fmt.Println("Server unavailable. Could not send message.")
			continue
		}
	}
}
