package main

import (
	"fmt"
	"vsm-protocol/core/protocol"
)

func main() {

	srv := protocol.NewServer(":9999")

	srv.RegisterHandler("AUTH", func(c *protocol.Client, m protocol.Message) {
		fmt.Printf(" [AUTH HANDLER] Processing token: %s\n", m.Payload)

		// 1. Normal response
		c.SendMessage(protocol.Message{Command: "AUTH_RESPONSE", Payload: "SUCCESS"})

		// 2. Spontaneous Announcement
		c.SendMessage(protocol.Message{Command: "SERVER_ANNOUNCEMENT", Payload: "Welcome to the VSM Mesh!"})
	})

	srv.RegisterHandler("PING", func(c *protocol.Client, m protocol.Message) {
		fmt.Printf(" [PING HANDLER] Got ping #%s\n", m.Payload)

		// Echo it back so the client mailbox gets filled
		c.SendMessage(protocol.Message{
			Command: "PONG",
			Payload: m.Payload,
		})
	})

	if err := srv.Start(); err != nil {
		fmt.Println("Server error:", err)
	}

	fmt.Println("Server example running. Implement your listening logic here.")
}
