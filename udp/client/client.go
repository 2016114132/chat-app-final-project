// Starts UDP client
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
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

	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	startTime = time.Now()

	// Handle graceful exit when Ctrl+C or kill signal is received
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nDisconnected.")
		close(stopChan) // Signal the read loop to stop
		conn.Close()    // Close the connection
		wg.Wait()       // Wait for read goroutine to exit
		printMetrics()  // Print your final metrics
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
	wg.Add(1)
	go func() {
		defer wg.Done()

		buffer := make([]byte, 1024)
		for {
			select {
			case <-stopChan:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(1 * time.Second)) // avoid blocking forever

				n, _, err := conn.ReadFromUDP(buffer)
				if err != nil {
					if os.IsTimeout(err) {
						continue
					}
					return // exit silently on any other error (like conn closed)
				}

				atomic.AddInt64(&messagesReceived, 1)
				receivedTime := time.Now()

				message := strings.TrimSpace(string(buffer[:n]))
				parts := strings.SplitN(message, "|", 2)
				if len(parts) == 2 {
					sentTime, err := time.Parse(time.RFC3339Nano, parts[0])
					if err == nil {
						latency := receivedTime.Sub(sentTime).Nanoseconds()
						atomic.AddInt64(&totalLatency, latency)
						message = parts[1]
					}
				}

				fmt.Printf("\r%s\n", message)
				fmt.Print("You: ")
			}
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

		// Command line to send spam messages
		if strings.HasPrefix(text, "/spam ") {
			countStr := strings.TrimPrefix(text, "/spam ")
			count, err := strconv.Atoi(countStr)
			if err != nil || count <= 0 {
				fmt.Println("Usage: /spam <positive number>")
				continue
			}

			for i := 0; i < count; i++ {
				msg := fmt.Sprintf("Spam %d", i+1)
				timestamp := time.Now().Format(time.RFC3339Nano)
				message := fmt.Sprintf("%s|%s", timestamp, msg)
				_, err := conn.Write([]byte(message + "\n"))
				if err != nil {
					fmt.Println("Failed to send spam message:", err)
					break
				}
				atomic.AddInt64(&messagesSent, 1)
				time.Sleep(20 * time.Millisecond) // optional: give some breathing room
			}
			continue
		}

		// Send the message
		timestamp := time.Now().Format(time.RFC3339Nano)
		message := fmt.Sprintf("%s|%s", timestamp, text)

		_, err := conn.Write([]byte(message + "\n"))
		atomic.AddInt64(&messagesSent, 1)

		// Display an error if the server is unreachable
		if err != nil {
			fmt.Println("Server unavailable. Could not send message.")
			continue
		}
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
		averageLatency = 0
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
