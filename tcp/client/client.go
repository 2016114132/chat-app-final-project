// Starts TCP client
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/2016114132/chat-app-final-project/shared"
)

var (
	messagesSent     int64
	messagesReceived int64
	startTime        time.Time
	totalLatency     int64
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
		conn.Close()   // Ensure connection is closed
		printMetrics() // Print metrics before exiting
		os.Exit(0)     // Exit the program
	}()

	startTime = time.Now() // Record the start time for metrics

	// Start a goroutine to continuously read and display messages from the server
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			atomic.AddInt64(&messagesReceived, 1) // Increment received message count
			receivedTime := time.Now()            // Record the time when the message is received

			//Extract and parse the timestamp from the message
			message := scanner.Text()

			parts := strings.SplitN(message, "|", 2)
			if len(parts) == 2 {
				sentTime, err := time.Parse(time.RFC3339Nano, parts[0])
				if err == nil {
					// Calculate the latency by subtracting the sent time from the received time
					latency := receivedTime.Sub(sentTime).Nanoseconds()
					atomic.AddInt64(&totalLatency, latency) // Add latency to the total

					// Just keep the message part and remove the time after it has been used to calculate latency
					message = parts[1]
				}
			}

			// Use carriage return (\r) to overwrite the current prompt line
			fmt.Printf("\r%s\n", message) // Display incoming message
			fmt.Print("You: ")            // Reprint input prompt
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

		// Format the message with a timestamp
		timestamp := time.Now().Format(time.RFC3339Nano)
		message := fmt.Sprintf("%s|%s", timestamp, text)

		// Send the message to the server, with newline to match protocol
		_, err = conn.Write([]byte(message + "\n"))
		if err != nil {
			// If connection is lost, notify user and exit
			fmt.Println("Server unavailable. Could not send message.")
			break
		}
		atomic.AddInt64(&messagesSent, 1) // Increment sent message count
	}
}
func printMetrics() {
	duration := time.Since(startTime).Seconds()
	received := atomic.LoadInt64(&messagesReceived)
	sent := atomic.LoadInt64(&messagesSent)
	latencySumValue := atomic.LoadInt64(&totalLatency)

	var averageLatency float64
	if received > 0 {
		averageLatency = float64(latencySumValue) / float64(received) / 1e6
	} else {
		averageLatency = 0 // No latency data available
	}

	fmt.Printf("\n--- Metrics ---\n")
	fmt.Printf("Messages Sent: %d\n", sent)
	fmt.Printf("Messages Received: %d\n", received)
	if sent > 0 {
		fmt.Printf("Packet Loss: %.2f%%\n", 100*(1-float64(received)/float64(sent)))
	} else {
		fmt.Printf("Packet Loss: N/A\n")
	}
	if duration > 0 {
		fmt.Printf("Throughput: %.2f messages/sec\n", float64(received)/duration)
	} else {
		fmt.Printf("Throughput: N/A\n")
	}
	fmt.Printf("Average Latency: %.2f ms\n", averageLatency)
}
