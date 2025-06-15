package proto

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestMessage_GetType(t *testing.T) {
	msg := Message{Type: MessageTypeRegister}
	if got := msg.GetType(); got != MessageTypeRegister {
		t.Errorf("GetType() = %v, want %v", got, MessageTypeRegister)
	}
}

func TestMessage_GetToken(t *testing.T) {
	msg := Message{Token: "test-token"}
	if got := msg.GetToken(); got != "test-token" {
		t.Errorf("GetToken() = %v, want %v", got, "test-token")
	}
}

func TestMessage_GetVersion(t *testing.T) {
	msg := Message{Version: "1.0.0"}
	if got := msg.GetVersion(); got != "1.0.0" {
		t.Errorf("GetVersion() = %v, want %v", got, "1.0.0")
	}
}

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name    string
		msgType MessageType
		want    bool
	}{
		{"register", MessageTypeRegister, true},
		{"config", MessageTypeConfig, true},
		{"connect", MessageTypeConnect, true},
		{"ping", MessageTypePing, true},
		{"pong", MessageTypePong, true},
		{"error", MessageTypeError, true},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(tt.msgType)
			if (msg != nil) != tt.want {
				t.Errorf("NewMessage() = %v, want %v", msg != nil, tt.want)
			}
			if msg != nil && msg.GetType() != tt.msgType {
				t.Errorf("NewMessage().GetType() = %v, want %v", msg.GetType(), tt.msgType)
			}
		})
	}
}

func TestMessageSerialization(t *testing.T) {
	tests := []struct {
		name     string
		msg      *Message
		wantJSON string
	}{
		{
			name: "register message",
			msg: &Message{
				Type:  MessageTypeRegister,
				Token: "test-token",
			},
			wantJSON: `{"type":1,"token":"test-token"}`,
		},
		{
			name: "config message",
			msg: &Message{
				Type: MessageTypeConfig,
				Config: &ConfigMessage{
					Token: "test-token",
				},
			},
			wantJSON: `{"type":4,"config":{"token":"test-token"}}`,
		},
		{
			name: "error message",
			msg: &Message{
				Type:  MessageTypeError,
				Error: "test error",
			},
			wantJSON: `{"type":6,"error":"test error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("Failed to marshal message: %v", err)
			}

			if string(data) != tt.wantJSON {
				t.Errorf("JSON = %s, want %s", string(data), tt.wantJSON)
			}

			// Test JSON deserialization
			var got Message
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Failed to unmarshal message: %v", err)
			}

			if got.Type != tt.msg.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.msg.Type)
			}
		})
	}
}

func TestReadMessage(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    *Message
		wantErr bool
	}{
		{
			name: "valid message",
			data: []byte{0, 0, 0, 25, 123, 34, 116, 121, 112, 101, 34, 58, 49, 44, 34, 116, 111, 107, 101, 110, 34, 58, 34, 116, 101, 115, 116, 34, 125},
			want: &Message{
				Type:  MessageTypeRegister,
				Token: "test",
			},
			wantErr: false,
		},
		{
			name:    "invalid length",
			data:    []byte{0, 0, 0, 1},
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			data:    []byte{0, 0, 0, 5, 123, 34, 116, 121, 112},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadMessage(bytes.NewReader(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Type != tt.want.Type {
				t.Errorf("ReadMessage().Type = %v, want %v", got.Type, tt.want.Type)
			}
		})
	}
}

func TestWriteMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     *Message
		wantErr bool
	}{
		{
			name: "register message",
			msg: &Message{
				Type:  MessageTypeRegister,
				Token: "test",
			},
			wantErr: false,
		},
		{
			name: "config message",
			msg: &Message{
				Type: MessageTypeConfig,
				Config: &ConfigMessage{
					Token: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "error message",
			msg: &Message{
				Type:  MessageTypeError,
				Error: "test error",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteMessage(&buf, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Read back the message
				got, err := ReadMessage(&buf)
				if err != nil {
					t.Fatalf("Failed to read back message: %v", err)
				}

				if got.Type != tt.msg.Type {
					t.Errorf("Type = %v, want %v", got.Type, tt.msg.Type)
				}
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name    string
		err     string
		wantErr bool
	}{
		{
			name:    "valid error",
			err:     "test error",
			wantErr: false,
		},
		{
			name:    "empty error",
			err:     "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteError(&buf, tt.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Read back the message
				got, err := ReadMessage(&buf)
				if err != nil {
					t.Fatalf("Failed to read back error message: %v", err)
				}

				if got.Type != MessageTypeError {
					t.Errorf("Type = %v, want %v", got.Type, MessageTypeError)
				}

				if got.Error != tt.err {
					t.Errorf("Error = %v, want %v", got.Error, tt.err)
				}
			}
		})
	}
}

func TestMessageTimeout(t *testing.T) {
	// Create a slow reader
	slowReader := &slowReader{
		delay: 100 * time.Millisecond,
		data:  []byte{0, 0, 0, 5, 123, 34, 116, 121, 112, 101, 34, 58, 49, 125},
	}

	// Set a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Try to read message
	_, err := ReadMessageWithContext(ctx, slowReader)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// slowReader is a reader that introduces a delay
type slowReader struct {
	delay time.Duration
	data  []byte
	pos   int
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	time.Sleep(r.delay)
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
} 