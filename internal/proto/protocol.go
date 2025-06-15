package proto

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
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
	// Читаем длину сообщения (4 байта BigEndian)
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return nil, fmt.Errorf("failed to read message length: %v", err)
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// Читаем само сообщение
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, fmt.Errorf("failed to read message data: %v", err)
	}

	// Декодируем JSON
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	return &msg, nil
}

// WriteMessage writes a message to the connection
func WriteMessage(conn io.Writer, msg *Message) error {
	// Кодируем сообщение в JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Записываем длину сообщения (4 байта BigEndian)
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))
	if _, err := conn.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	// Записываем само сообщение
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %v", err)
	}

	return nil
}

// WriteError writes an error message to the connection
func WriteError(conn io.Writer, err string) error {
	return WriteMessage(conn, &Message{
		Type:  MessageTypeError,
		Error: err,
	})
} 
