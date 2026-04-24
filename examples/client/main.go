package main

import (
	"fmt"
	"net"
	"vsm-protocol/core/protocol"
)

func main() {

	conn, err := net.Dial("tcp", "localhost:9999")

	if err != nil {
		fmt.Println(err)
		return
	}

	client := protocol.NewClient(conn)

	client.RegisterHandler("SERVER_ANNOUNCEMENT", func(c *protocol.Client, m protocol.Message) {
		fmt.Printf(" [AUTO-HANDLER] Caught Server Announcement: %s\n", m.Payload)
	})

	msg := protocol.Message{
		Command: "AUTH",
		Payload: "cezar-admin-token",
	}
	client.SendMessage(msg)

	echo := <-client.Incoming

	fmt.Printf(" [SUCCESS] Got response from channel: %s\n", echo.Payload)

	client.SendMessage(protocol.Message{Command: "PING", Payload: "1"})
	client.SendMessage(protocol.Message{Command: "PING", Payload: "2"})
	client.SendMessage(protocol.Message{Command: "PING", Payload: "3"})

	for i := 0; i < 3; i++ {
		response := <-client.Incoming
		fmt.Printf("Collected response %d: %s\n", i+1, response.Payload)
	}

	fmt.Println("Client example running. Implement your connection logic here.")
}
