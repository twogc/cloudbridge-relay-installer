package masque

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
	"go.uber.org/zap"
)

// Capsule types as defined in RFC 9297
const (
	CapsuleTypeDatagram = 0x00
	CapsuleTypeClose    = 0x01
)

// CapsuleHandler handles HTTP Capsules when DATAGRAM is not available
type CapsuleHandler struct {
	logger *zap.Logger
	mx     sync.RWMutex
	closed bool
}

// Capsule represents an HTTP Capsule (RFC 9297)
type Capsule struct {
	Type   uint64
	Length uint64
	Value  []byte
}

// NewCapsuleHandler creates a new capsule handler
func NewCapsuleHandler(logger *zap.Logger) *CapsuleHandler {
	return &CapsuleHandler{
		logger: logger,
	}
}

// HandleCapsuleStream handles capsules on an HTTP stream (fallback when DATAGRAM not supported)
func (h *CapsuleHandler) HandleCapsuleStream(stream *http3.Stream, datagramHandler func([]byte) error) error {
	h.mx.RLock()
	if h.closed {
		h.mx.RUnlock()
		return fmt.Errorf("capsule handler is closed")
	}
	h.mx.RUnlock()

	reader := quicvarint.NewReader(stream)
	
	for {
		// Read capsule type
		capsuleType, err := quicvarint.Read(reader)
		if err != nil {
			if err == io.EOF {
				return nil // Stream closed normally
			}
			return fmt.Errorf("failed to read capsule type: %w", err)
		}

		// Read capsule length
		capsuleLength, err := quicvarint.Read(reader)
		if err != nil {
			return fmt.Errorf("failed to read capsule length: %w", err)
		}

		// Read capsule value
		value := make([]byte, capsuleLength)
		if capsuleLength > 0 {
			_, err = io.ReadFull(reader, value)
			if err != nil {
				return fmt.Errorf("failed to read capsule value: %w", err)
			}
		}

		capsule := &Capsule{
			Type:   capsuleType,
			Length: capsuleLength,
			Value:  value,
		}

		h.logger.Debug("Received capsule",
			zap.Uint64("type", capsule.Type),
			zap.Uint64("length", capsule.Length))

		// Handle capsule based on type
		switch capsule.Type {
		case CapsuleTypeDatagram:
			// Handle datagram capsule - extract the datagram payload
			if len(capsule.Value) == 0 {
				h.logger.Warn("Empty datagram capsule")
				continue
			}
			
			// For MASQUE, the capsule value contains the HTTP datagram
			if datagramHandler != nil {
				if err := datagramHandler(capsule.Value); err != nil {
					h.logger.Error("Failed to handle datagram from capsule", zap.Error(err))
				}
			}

		case CapsuleTypeClose:
			// Handle close capsule
			h.logger.Info("Received close capsule")
			return nil

		default:
			// Unknown capsule type - skip according to RFC 9297
			h.logger.Debug("Skipping unknown capsule type", zap.Uint64("type", capsule.Type))
		}
	}
}

// SendDatagramCapsule sends a datagram as an HTTP Capsule
func (h *CapsuleHandler) SendDatagramCapsule(stream *http3.Stream, data []byte) error {
	h.mx.RLock()
	if h.closed {
		h.mx.RUnlock()
		return fmt.Errorf("capsule handler is closed")
	}
	h.mx.RUnlock()

	// Create datagram capsule
	capsule := &Capsule{
		Type:   CapsuleTypeDatagram,
		Length: uint64(len(data)),
		Value:  data,
	}

	return h.writeCapsule(stream, capsule)
}

// SendCloseCapsule sends a close capsule
func (h *CapsuleHandler) SendCloseCapsule(stream *http3.Stream) error {
	h.mx.RLock()
	if h.closed {
		h.mx.RUnlock()
		return fmt.Errorf("capsule handler is closed")
	}
	h.mx.RUnlock()

	// Create close capsule
	capsule := &Capsule{
		Type:   CapsuleTypeClose,
		Length: 0,
		Value:  nil,
	}

	return h.writeCapsule(stream, capsule)
}

// writeCapsule writes a capsule to the stream
func (h *CapsuleHandler) writeCapsule(stream *http3.Stream, capsule *Capsule) error {
	// Calculate total size
	typeSize := quicvarint.Len(capsule.Type)
	lengthSize := quicvarint.Len(capsule.Length)
	totalSize := typeSize + lengthSize + len(capsule.Value)

	// Prepare buffer
	buf := make([]byte, totalSize)
	offset := 0

	// Write type using Append
	typeBuf := quicvarint.Append(nil, capsule.Type)
	copy(buf[offset:], typeBuf)
	offset += len(typeBuf)

	// Write length using Append
	lengthBuf := quicvarint.Append(nil, capsule.Length)
	copy(buf[offset:], lengthBuf)
	offset += len(lengthBuf)

	// Write value
	if len(capsule.Value) > 0 {
		copy(buf[offset:], capsule.Value)
	}

	// Write to stream
	_, err := stream.Write(buf)
	if err != nil {
		return fmt.Errorf("failed to write capsule: %w", err)
	}

	h.logger.Debug("Sent capsule",
		zap.Uint64("type", capsule.Type),
		zap.Uint64("length", capsule.Length))

	return nil
}

// Close closes the capsule handler
func (h *CapsuleHandler) Close() error {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.closed = true
	return nil
}

// MASQUETransport represents a transport that can fallback between DATAGRAM and Capsule
type MASQUETransport struct {
	logger          *zap.Logger
	stream          *http3.Stream
	capsuleHandler  *CapsuleHandler
	useDatagrams    bool
	datagramSupport bool
	mx              sync.RWMutex
}

// NewMASQUETransport creates a new MASQUE transport with fallback support
func NewMASQUETransport(logger *zap.Logger, stream *http3.Stream) *MASQUETransport {
	transport := &MASQUETransport{
		logger:         logger,
		stream:         stream,
		capsuleHandler: NewCapsuleHandler(logger),
	}

	// Check if DATAGRAM is supported on the connection
	transport.datagramSupport = transport.checkDatagramSupport()
	transport.useDatagrams = transport.datagramSupport

	transport.logger.Info("MASQUE transport initialized",
		zap.Bool("datagram_support", transport.datagramSupport),
		zap.Bool("use_datagrams", transport.useDatagrams))

	return transport
}

// checkDatagramSupport checks if QUIC DATAGRAM is supported
func (t *MASQUETransport) checkDatagramSupport() bool {
	// In a real implementation, this would check the QUIC connection's datagram support
	// For now, we'll assume it's supported and handle errors gracefully
	return true
}

// SendData sends data using DATAGRAM or Capsule fallback
func (t *MASQUETransport) SendData(data []byte) error {
	t.mx.RLock()
	useDatagrams := t.useDatagrams
	t.mx.RUnlock()

	if useDatagrams {
		// Try to send as DATAGRAM first
		err := t.stream.SendDatagram(data)
		if err != nil {
			t.logger.Warn("Failed to send datagram, falling back to capsule", zap.Error(err))
			
			// Fallback to capsule
			t.mx.Lock()
			t.useDatagrams = false
			t.mx.Unlock()
			
			return t.capsuleHandler.SendDatagramCapsule(t.stream, data)
		}
		return nil
	}

	// Use capsule
	return t.capsuleHandler.SendDatagramCapsule(t.stream, data)
}

// ReceiveData receives data from DATAGRAM or Capsule
func (t *MASQUETransport) ReceiveData(ctx context.Context) ([]byte, error) {
	t.mx.RLock()
	useDatagrams := t.useDatagrams
	t.mx.RUnlock()

	if useDatagrams {
		// Try to receive DATAGRAM first
		data, err := t.stream.ReceiveDatagram(ctx)
		if err != nil {
			// Check if this is a datagram-related error that suggests fallback
			t.logger.Debug("Failed to receive datagram", zap.Error(err))
			// In a real implementation, we'd analyze the error type
			// For now, we'll continue with datagram mode
			return nil, err
		}
		return data, nil
	}

	// Use capsule - this would require a more complex implementation
	// that reads from the stream and parses capsules
	return nil, fmt.Errorf("capsule receive not implemented in this example")
}

// Close closes the transport
func (t *MASQUETransport) Close() error {
	return t.capsuleHandler.Close()
}

// GetTransportInfo returns information about the current transport mode
func (t *MASQUETransport) GetTransportInfo() map[string]interface{} {
	t.mx.RLock()
	defer t.mx.RUnlock()
	
	return map[string]interface{}{
		"datagram_support": t.datagramSupport,
		"use_datagrams":    t.useDatagrams,
		"transport_mode":   t.getTransportMode(),
	}
}

func (t *MASQUETransport) getTransportMode() string {
	if t.useDatagrams {
		return "DATAGRAM"
	}
	return "CAPSULE"
}
