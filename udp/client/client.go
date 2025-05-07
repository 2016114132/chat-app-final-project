// Starts UDP client
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
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
	latencySum	  	int64 
	latencyCount	  int64 
	latencyMutex	  sync.Mutex 
	pingTimestamps	  =make(map[string]time.Time)
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

	//Channel to signal the reading goroutine to stop
	stopChan := make(chan struct{})
	// Start a goroutine to handle the signal
	var wg sync.WaitGroup

	// Record the start time for metrics
	startTime = time.Now() 

	// Start a goroutine that sends periodic "/ping" messages to the server
	// This acts as a heartbeat to let the server know the client is still active
	go func() {
		for {
			select {
				case <-stopChan:
					return
				default:
					// Sleep for 2 seconds before sending the next ping
					time.Sleep(2 * time.Second)
					timestamp := time.Now()
					messageID := fmt.Sprintf("%d", timestamp.UnixNano()) // Unique message ID
					pingMessage := fmt.Sprintf("/ping %s\n", messageID)

					latencyMutex.Lock()
					pingTimestamps[messageID] = timestamp // Store the timestamp for this ping
					latencyMutex.Unlock()

					conn.Write([]byte(pingMessage))
				
			}
		}
	}()

	// Start a goroutine to continuously read incoming messages from the server
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Create a buffer for incoming data
		buffer := make([]byte, 1024)
		for {
			select {
			case <-stopChan:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(1 * time.Second)) // Prevent blocking forever
				
				n, _, err := conn.ReadFromUDP(buffer)
				if err != nil {
					if os.IsTimeout(err){ 
						continue
					}
					fmt.Printf("\rRead error: %s\n", err)
					return
				}
				atomic.AddInt64(&messagesReceived, 1) // Increment received message count

				message := strings.TrimSpace(string(buffer[:n]))
				// Check if the message is a pong response
				if strings.HasPrefix(message, "/pong") {
					// Extract the message ID from the ping response
					parts := strings.Split(message, " ")
					if len(parts) == 3 {
						messageID := parts[1]

						latencyMutex.Lock()
						if sentTime, exists := pingTimestamps[messageID]; exists {
							// Calculate the latency
							latency := time.Since(sentTime).Milliseconds()
							atomic.AddInt64(&latencySum, latency)
							atomic.AddInt64(&latencyCount, 1)
							delete(pingTimestamps, messageID) // Remove the ping timestamp after processing
						}
						latencyMutex.Unlock()
					}
				}else{
					// Print the received message, trimming any trailing whitespace
					fmt.Printf("\r%s\n", message)
					fmt.Print("You: ")
				}
			}
			

			
		}
	}()

	go func() {
		<-sig
		fmt.Println("\nDisconnected.")
		close(stopChan) //Signal goroutines to stop
		conn.Close()
		wg.Wait() // Wait for all goroutines to finish
		printMetrics() // Print metrics before exiting
		os.Exit(0)
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
		atomic.AddInt64(&messagesSent, 1) // Increment sent message count
	}
}

// printMetrics prints the metrics of the chat session
func printMetrics() {
	duration := time.Since(startTime).Seconds()
	received := atomic.LoadInt64(&messagesReceived)
    sent := atomic.LoadInt64(&messagesSent)
    latencySumValue := atomic.LoadInt64(&latencySum)
    latencyCountValue := atomic.LoadInt64(&latencyCount)

    var averageLatency float64
    if latencyCountValue > 0 {
        averageLatency = float64(latencySumValue) / float64(latencyCountValue)
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