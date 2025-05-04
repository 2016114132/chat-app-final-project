// Starts UDP server
package server

import (
	"fmt"
	"net"
)

func Start(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("UDP Server listening on", address)

	handleConnection(conn)
}
