package proto

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestMessage_GetType(t *testing.T) {
	msg := Message{Type: MessageTypeHello}
	if msg.Type != MessageTypeHello {
		t.Errorf("Type = %v, want %v", msg.Type, MessageTypeHello)
	}
}

func TestMessage_GetToken(t *testing.T) {
	msg := Message{Token: "test-token"}
	if msg.Token != "test-token" {
		t.Errorf("Token = %v, want %v", msg.Token, "test-token")
	}
}

func TestMessage_GetVersion(t *testing.T) {
	msg := Message{Version: "1.0.0"}
	if msg.Version != "1.0.0" {
		t.Errorf("Version = %v, want %v", msg.Version, "1.0.0")
	}
}

func TestNewMessage(t *testing.T) {
	// This test is not applicable as there is no NewMessage function or similar logic in protocol.go
}

func TestMessageSerialization(t *testing.T) {
	tests := []struct {
		name     string
		msg      *Message
		wantJSON string
	}{
		{
			name: "hello message",
			msg: &Message{
				Type:  MessageTypeHello,
				Token: "test-token",
			},
			wantJSON: `{"type":"hello","token":"test-token"}`,
		},
		{
			name: "tunnel info message",
			msg: &Message{
				Type: MessageTypeTunnelInfo,
				TunnelInfo: &TunnelInfo{
					TunnelID:   "id1",
					LocalPort:  1234,
					RemoteHost: "host",
					RemotePort: 5678,
				},
			},
			wantJSON: `{"type":"tunnel_info","tunnel_info":{"tunnel_id":"id1","local_port":1234,"remote_host":"host","remote_port":5678}}`,
		},
		{
			name: "error message",
			msg: &Message{
				Type:  MessageTypeError,
				Error: &ErrorMessage{Code: "err", Message: "test error"},
			},
			wantJSON: `{"type":"error","error":{"code":"err","message":"test error"}}`,
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
			data: []byte(`{"type":"hello","token":"test"}
`),
			want: &Message{
				Type:  MessageTypeHello,
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
			got, err := ReadMessage(bufio.NewReader(bytes.NewReader(tt.data)))
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
			name: "hello message",
			msg: &Message{
				Type:  MessageTypeHello,
				Token: "test",
			},
			wantErr: false,
		},
		{
			name: "error message",
			msg: &Message{
				Type:  MessageTypeError,
				Error: &ErrorMessage{Code: "err", Message: "test error"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteMessage(bufio.NewWriter(&buf), tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Read back the message
				got, err := ReadMessage(bufio.NewReader(&buf))
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
			err := WriteError(bufio.NewWriter(&buf), "err", tt.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Read back the message
				got, err := ReadMessage(bufio.NewReader(&buf))
				if err != nil {
					t.Fatalf("Failed to read back error message: %v", err)
				}

				if got.Type != MessageTypeError {
					t.Errorf("Type = %v, want %v", got.Type, MessageTypeError)
				}

				if got.Error == nil || got.Error.Message != tt.err {
					t.Errorf("Error = %v, want %v", got.Error, tt.err)
				}
			}
		})
	}
}

// slowReader is a reader that introduces a delay
//nolint:unused // Used in tests
type slowReader struct {
	delay time.Duration
	data  []byte
	pos   int
}

//nolint:unused // Used in tests
func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	time.Sleep(r.delay)
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
