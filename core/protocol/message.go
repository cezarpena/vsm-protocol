package protocol

// Message represents the base structure of a protocol message.
type Message struct {
	Command string `json:"command"` // e.g. "AUTH", "DATA", "QUIT"
	Payload string `json:"payload"` // The actual data
}

// TODO: Define specific message types (like AuthMessage, DataMessage, etc.) or verbs
