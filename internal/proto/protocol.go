package proto

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// Constants for message framing
const (
	// MaxMessageSize is the maximum allowed size of a message in bytes
	MaxMessageSize = 1024 * 1024 // 1MB

	// LengthPrefixSize is the size of the length prefix in bytes
	LengthPrefixSize = 4
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeRegister MessageType = "register"
	MessageTypeConfig   MessageType = "config"
	MessageTypeConnect  MessageType = "connect"
	MessageTypePing     MessageType = "ping"
	MessageTypePong     MessageType = "pong"
	MessageTypeError    MessageType = "error"
)

// RelayMessage is the interface that all message types must implement
type RelayMessage interface {
	GetType() MessageType
	GetToken() string
	GetVersion() string
}

// Message is the base message type
type Message struct {
	Type    MessageType `json:"type"`
	Token   string     `json:"token,omitempty"`
	Version string     `json:"version,omitempty"`
}

// GetType returns the message type
func (m Message) GetType() MessageType {
	return m.Type
}

// GetToken returns the message token
func (m Message) GetToken() string {
	return m.Token
}

// GetVersion returns the message version
func (m Message) GetVersion() string {
	return m.Version
}

// RegisterMessage is sent by the client to register a tunnel
type RegisterMessage struct {
	Message
}

// ConfigMessage is sent by the server in response to registration
type ConfigMessage struct {
	Message
	PublicURL string `json:"public_url"`
	Status    string `json:"status"`
}

// ConnectMessage is sent by the client to connect to a tunnel
type ConnectMessage struct {
	Message
}

// PingMessage is sent to keep the connection alive
type PingMessage struct {
	Message
}

// PongMessage is sent in response to a ping
type PongMessage struct {
	Message
}

// ErrorMessage is sent when an error occurs
type ErrorMessage struct {
	Message
	Error string `json:"error"`
}

// NewMessage creates a new message of the specified type
func NewMessage(msgType MessageType) RelayMessage {
	switch msgType {
	case MessageTypeRegister:
		return &RegisterMessage{Message: Message{Type: msgType}}
	case MessageTypeConfig:
		return &ConfigMessage{Message: Message{Type: msgType}}
	case MessageTypeConnect:
		return &ConnectMessage{Message: Message{Type: msgType}}
	case MessageTypePing:
		return &PingMessage{Message: Message{Type: msgType}}
	case MessageTypePong:
		return &PongMessage{Message: Message{Type: msgType}}
	case MessageTypeError:
		return &ErrorMessage{Message: Message{Type: msgType}}
	default:
		return nil
	}
}

// ReadMessage reads a length-prefixed message from the connection
func ReadMessage(conn io.Reader) (RelayMessage, error) {
	// Read message length (4 bytes, big endian)
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read message length: %v", err)
	}

	// Check message size
	if length > MaxMessageSize {
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, fmt.Errorf("failed to read message data: %v", err)
	}

	// Unmarshal base message to determine type
	var base Message
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("failed to unmarshal base message: %v", err)
	}

	// Create appropriate message type
	msg := NewMessage(base.Type)
	if msg == nil {
		return nil, fmt.Errorf("unknown message type: %s", base.Type)
	}

	// Unmarshal full message
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	return msg, nil
}

// WriteMessage writes a message to the connection
func WriteMessage(conn io.Writer, msg RelayMessage) error {
	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Check message size
	if len(data) > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes", len(data))
	}

	// Write message length (4 bytes, big endian)
	if err := binary.Write(conn, binary.BigEndian, uint32(len(data))); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	// Write message data
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %v", err)
	}

	return nil
}

// WriteError writes an error message to the connection
func WriteError(conn io.Writer, errMsg string) error {
	msg := &ErrorMessage{
		Message: Message{
			Type: MessageTypeError,
		},
		Error: errMsg,
	}
	return WriteMessage(conn, msg)
}

// WritePong writes a pong message to the connection
func WritePong(conn io.Writer, token string) error {
	msg := &PongMessage{
		Message: Message{
			Type:  MessageTypePong,
			Token: token,
		},
	}
	return WriteMessage(conn, msg)
} 


cmd/relay/main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/maxlanies/bore/internal/proto"
)

var (
	addr    = flag.String("addr", ":12345", "Address to listen on")
	timeout = flag.Duration("timeout", 10*time.Minute, "Connection timeout")
)

// Error codes for better client handling
const (
	ErrEmptyToken     = "EMPTY_TOKEN"
	ErrTunnelNotFound = "TUNNEL_NOT_FOUND"
	ErrAlreadyExists  = "ALREADY_EXISTS"
	ErrInvalidMessage = "INVALID_MESSAGE"
)

type Tunnel struct {
	Token      string
	ClientConn net.Conn
	mu         sync.Mutex
	LastPing   time.Time
}

type Server struct {
	addr     string
	timeout  time.Duration
	tunnels  map[string]*Tunnel
	mu       sync.RWMutex
	logger   *log.Logger
}

func NewServer(addr string, timeout time.Duration) *Server {
	return &Server{
		addr:    addr,
		timeout: timeout,
		tunnels: make(map[string]*Tunnel),
		logger:  log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()

	// Read initial message
	data, err := proto.ReadMessageFromConn(conn)
	if err != nil {
		s.logger.Printf("Failed to read message from %s: %v", remoteAddr, err)
		proto.WriteError(conn, fmt.Sprintf("%s: %v", ErrInvalidMessage, err))
		return
	}

	var msg proto.Message
	if err := proto.UnmarshalMessage(data, &msg); err != nil {
		s.logger.Printf("Failed to unmarshal message from %s: %v", remoteAddr, err)
		proto.WriteError(conn, fmt.Sprintf("%s: %v", ErrInvalidMessage, err))
		return
	}

	// Handle registration
	if msg.Type == string(proto.MessageTypeRegister) {
		s.handleRegistration(conn, msg)
		return
	}

	// Handle consumer connection
	if msg.Type == string(proto.MessageTypeConnect) {
		s.handleClientConnection(conn, msg)
		return
	}

	// Handle ping
	if msg.Type == string(proto.MessageTypePing) {
		if err := proto.WritePong(conn); err != nil {
			s.logger.Printf("Failed to send pong to %s: %v", remoteAddr, err)
		}
		return
	}

	s.logger.Printf("Unknown message type from %s: %s", remoteAddr, msg.Type)
	proto.WriteError(conn, fmt.Sprintf("%s: unknown message type %s", ErrInvalidMessage, msg.Type))
}

func (s *Server) handleRegistration(conn net.Conn, msg proto.Message) {
	remoteAddr := conn.RemoteAddr().String()
	token := msg.Token
	if token == "" {
		s.logger.Printf("Empty token in registration from %s", remoteAddr)
		proto.WriteError(conn, ErrEmptyToken)
		return
	}

	s.mu.Lock()
	// Check if tunnel already exists
	if existing, ok := s.tunnels[token]; ok {
		s.logger.Printf("Tunnel %s already registered from %s, closing old connection", token, remoteAddr)
		existing.mu.Lock()
		existing.ClientConn.Close()
		existing.mu.Unlock()
		proto.WriteError(conn, ErrAlreadyExists)
		return
	}

	// Create new tunnel
	tunnel := &Tunnel{
		Token:      token,
		ClientConn: conn,
		LastPing:   time.Now(),
	}
	s.tunnels[token] = tunnel
	s.mu.Unlock()

	s.logger.Printf("Tunnel registered from %s: %s", remoteAddr, token)

	// Send registration response
	response := proto.ConfigMessage{
		Type:      proto.MessageTypeConfig,
		Token:     token,
		PublicURL: fmt.Sprintf("bore://%s", token),
		Status:    "registered",
	}

	data, err := proto.MarshalMessage(response)
	if err != nil {
		s.logger.Printf("Failed to marshal response to %s: %v", remoteAddr, err)
		proto.WriteError(conn, fmt.Sprintf("%s: %v", ErrInvalidMessage, err))
		return
	}

	if _, err := conn.Write(data); err != nil {
		s.logger.Printf("Failed to send response to %s: %v", remoteAddr, err)
		return
	}

	// Keep connection alive with ping/pong
	for {
		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(s.timeout))

		// Try to read a message
		data, err := proto.ReadMessageFromConn(conn)
		if err != nil {
			s.logger.Printf("Tunnel %s connection closed from %s: %v", token, remoteAddr, err)
			s.mu.Lock()
			delete(s.tunnels, token)
			s.mu.Unlock()
			return
		}

		// Reset deadline
		conn.SetReadDeadline(time.Time{})

		// Check if it's a ping message
		var msg proto.Message
		if err := proto.UnmarshalMessage(data, &msg); err != nil {
			continue
		}
		if msg.Type == string(proto.MessageTypePing) {
			tunnel.mu.Lock()
			tunnel.LastPing = time.Now()
			tunnel.mu.Unlock()
			if err := proto.WritePong(conn); err != nil {
				s.logger.Printf("Failed to send pong to %s: %v", remoteAddr, err)
			}
		}
	}
}

func (s *Server) handleClientConnection(conn net.Conn, msg proto.Message) {
	remoteAddr := conn.RemoteAddr().String()
	token := msg.Token
	if token == "" {
		s.logger.Printf("Empty token in connection from %s", remoteAddr)
		proto.WriteError(conn, ErrEmptyToken)
		return
	}

	s.mu.RLock()
	tunnel, ok := s.tunnels[token]
	s.mu.RUnlock()

	if !ok {
		s.logger.Printf("Tunnel not found from %s: %s", remoteAddr, token)
		proto.WriteError(conn, ErrTunnelNotFound)
		return
	}

	// Start bidirectional copy with proper cleanup
	done := make(chan struct{})
	var once sync.Once
	closeOnce := func() { once.Do(func() { close(done) }) }

	go func() {
		_, err := io.Copy(tunnel.ClientConn, conn)
		if err != nil {
			s.logger.Printf("Copy error (client->tunnel) from %s: %v", remoteAddr, err)
		}
		tunnel.ClientConn.Close()
		conn.Close()
		closeOnce()
	}()

	go func() {
		_, err := io.Copy(conn, tunnel.ClientConn)
		if err != nil {
			s.logger.Printf("Copy error (tunnel->client) from %s: %v", remoteAddr, err)
		}
		tunnel.ClientConn.Close()
		conn.Close()
		closeOnce()
	}()

	<-done

	// Cleanup tunnel if it's no longer needed
	s.mu.Lock()
	if t, ok := s.tunnels[token]; ok && t.ClientConn == tunnel.ClientConn {
		delete(s.tunnels, token)
	}
	s.mu.Unlock()
}

func (s *Server) Start() error {
	// Start server
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	s.logger.Printf("Server started on %s", s.addr)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		s.logger.Println("Shutting down...")
		// Close all tunnels
		s.mu.Lock()
		for _, tunnel := range s.tunnels {
			tunnel.ClientConn.Close()
		}
		s.mu.Unlock()
		listener.Close()
	}()

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			if err == net.ErrClosed {
				break
			}
			s.logger.Printf("Accept error: %v", err)
			continue
		}

		go s.handleClient(conn)
	}

	return nil
}

func main() {
	flag.Parse()

	server := NewServer(*addr, *timeout)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
} 
