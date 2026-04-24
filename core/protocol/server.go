package protocol

import (
	"fmt"
	"net"
)

// Server handles listening for incoming connections and creating protocol Clients.
type Server struct {
	port     string
	handlers map[string]CommandHandler // The "Phonebook" for commands
}

type CommandHandler func(*Client, Message)

// NewServer creates a new instance of a protocol Server.
func NewServer(port string) *Server {
	return &Server{
		port:     port,
		handlers: make(map[string]CommandHandler),
	}
}

func (s *Server) RegisterHandler(command string, handler CommandHandler) {
	s.handlers[command] = handler
}

// Start begins listening for incoming TCP connections.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	defer ln.Close()

	fmt.Println("Server started, listening on", s.port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// When a new raw connection arrives, we wrap it in our protocol's Client!
		client := NewClient(conn)

		// Ask a goroutine to handle this client
		go s.handleClient(client)
	}
}

// handleClient is where rules for the stream of incoming messages are enforced.
func (s *Server) handleClient(client *Client) {

	defer client.Close()
	fmt.Println("New client connected!")

	for msg := range client.Incoming {
		handler, exists := s.handlers[msg.Command]
		if exists {
			handler(client, msg)
		} else {
			fmt.Printf("Unknown command received: %s\n", msg.Command)
		}
	}

}
