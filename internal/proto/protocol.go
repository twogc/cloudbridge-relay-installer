package proto

import (
	"bufio"
	"encoding/json"
	"fmt"
)

const (
	// Максимальный размер сообщения (1MB)
	MaxMessageSize = 1024 * 1024
)

// MessageType represents the type of message
type MessageType string

const (
	// Типы сообщений
	MessageTypeHello           MessageType = "hello"
	MessageTypeAuth            MessageType = "auth"
	MessageTypeAuthResponse    MessageType = "auth_response"
	MessageTypeTunnelInfo      MessageType = "tunnel_info"
	MessageTypeHeartbeat       MessageType = "heartbeat"
	MessageTypeHeartbeatResponse MessageType = "heartbeat_response"
	MessageTypeError           MessageType = "error"
)

// Message represents a protocol message
type Message struct {
	Type    MessageType     `json:"type"`
	Token   string         `json:"token,omitempty"`
	Version string         `json:"version,omitempty"`
	Features []string      `json:"features,omitempty"`
	Status  string         `json:"status,omitempty"`
	ClientID string        `json:"client_id,omitempty"`
	TunnelInfo *TunnelInfo `json:"tunnel_info,omitempty"`
	Error   *ErrorMessage  `json:"error,omitempty"`
}

// TunnelInfo represents tunnel configuration
type TunnelInfo struct {
	TunnelID    string            `json:"tunnel_id"`
	LocalPort   int               `json:"local_port"`
	RemoteHost  string            `json:"remote_host"`
	RemotePort  int               `json:"remote_port"`
	Options     map[string]string `json:"options,omitempty"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ReadMessage reads a message from the connection
func ReadMessage(reader *bufio.Reader) (*Message, error) {
	// Читаем строку до разделителя \n
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %v", err)
	}

	// Проверяем размер сообщения
	if len(line) > MaxMessageSize {
		return nil, fmt.Errorf("message too large: %d bytes", len(line))
	}

	// Декодируем JSON
	var msg Message
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	return &msg, nil
}

// WriteMessage writes a message to the connection
func WriteMessage(writer *bufio.Writer, msg *Message) error {
	// Кодируем сообщение в JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Проверяем размер сообщения
	if len(data) > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes", len(data))
	}

	// Добавляем разделитель \n
	data = append(data, '\n')

	// Записываем сообщение
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	// Сбрасываем буфер
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %v", err)
	}

	return nil
}

// WriteError writes an error message to the connection
func WriteError(writer *bufio.Writer, code string, message string) error {
	return WriteMessage(writer, &Message{
		Type: MessageTypeError,
		Error: &ErrorMessage{
			Code:    code,
			Message: message,
		},
	})
}

// WriteHello writes a hello message
func WriteHello(writer *bufio.Writer, version string, features []string) error {
	return WriteMessage(writer, &Message{
		Type:     MessageTypeHello,
		Version:  version,
		Features: features,
	})
}

// WriteAuthResponse writes an auth response message
func WriteAuthResponse(writer *bufio.Writer, status string, clientID string) error {
	return WriteMessage(writer, &Message{
		Type:     MessageTypeAuthResponse,
		Status:   status,
		ClientID: clientID,
	})
}

// WriteHeartbeatResponse writes a heartbeat response message
func WriteHeartbeatResponse(writer *bufio.Writer) error {
	return WriteMessage(writer, &Message{
		Type:   MessageTypeHeartbeatResponse,
		Status: "ok",
	})
}

// WriteTunnelInfo writes a tunnel info message
func WriteTunnelInfo(writer *bufio.Writer, tunnelInfo *TunnelInfo) error {
	return WriteMessage(writer, &Message{
		Type:       MessageTypeTunnelInfo,
		TunnelInfo: tunnelInfo,
	})
} 
