package proto

import (
	"encoding/json"
	"io"
)

// MessageType represents the type of message
type MessageType int

const (
	MessageTypeRegister MessageType = iota + 1
	MessageTypeConnect
	MessageTypePing
	MessageTypePong
	MessageTypeConfig
	MessageTypeError
)

// Message represents a protocol message
type Message struct {
	Type    MessageType     `json:"type"`
	Token   string         `json:"token,omitempty"`
	Config  *ConfigMessage `json:"config,omitempty"`
	Pong    *PongMessage   `json:"pong,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// ConfigMessage represents a configuration message
type ConfigMessage struct {
	Token string `json:"token"`
}

// PongMessage represents a pong message
type PongMessage struct {
	Token string `json:"token"`
}

// ReadMessage reads a message from the connection
func ReadMessage(conn io.Reader) (*Message, error) {
	var msg Message
	if err := json.NewDecoder(conn).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// WriteMessage writes a message to the connection
func WriteMessage(conn io.Writer, msg *Message) error {
	return json.NewEncoder(conn).Encode(msg)
}

// WriteError writes an error message to the connection
func WriteError(conn io.Writer, err string) error {
	return WriteMessage(conn, &Message{
		Type:  MessageTypeError,
		Error: err,
	})
} 
