// Entry point
package main

import "github.com/2016114132/chat-app-final-project/tcp/server"

func main() {
	server.Start(":4000")
}