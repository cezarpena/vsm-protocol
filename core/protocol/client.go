package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

// Client represents a connected peer that can send and receive protocol Messages.
type Client struct {
	conn     net.Conn
	scanner  *bufio.Scanner
	Incoming chan Message // Mailbox
	handlers map[string]CommandHandler
}

// NewClient establishes a wrapper around an existing connection.
func NewClient(conn net.Conn) *Client {
	c := &Client{
		conn:     conn,
		scanner:  bufio.NewScanner(conn),
		Incoming: make(chan Message, 10), // Mailbox that holds up to 10 messages
		handlers: make(map[string]CommandHandler),
	}
	// Start background mail listener
	go c.listen()

	return c
}

// SendMessage is responsible for framing and sending a Message over the connection.
func (c *Client) SendMessage(msg Message) error {

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Glue the JSON and the Newline together
	data = append(data, '\n')

	_, err = c.conn.Write(data)
	return err

}

// ReceiveMessage reads data from the connection, un-frames it, and reconstructs a Message.
func (c *Client) ReceiveMessage() (Message, error) {

	if !c.scanner.Scan() {
		// Check if error or EOF / EOC
		err := c.scanner.Err()
		if err == nil {
			return Message{}, fmt.Errorf("connection closed")
		}
		return Message{}, err
	}

	text := c.scanner.Text()

	var msg Message
	err := json.Unmarshal([]byte(text), &msg)
	if err != nil {
		return Message{}, err
	}

	return msg, nil
}

// Close gracefully closes the underlying connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) listen() {
	// When loop is finished (socket gets closed), close mailbox - channel too
	defer close(c.Incoming)

	for {
		msg, err := c.ReceiveMessage()
		if err != nil {
			return
		}
		// Success. Putting message in the mailbox.
		handler, exists := c.handlers[msg.Command]

		if exists {
			handler(c, msg)
		} else {
			c.Incoming <- msg
		}
	}
}

func (c *Client) RegisterHandler(command string, handler CommandHandler) {
	c.handlers[command] = handler
}
