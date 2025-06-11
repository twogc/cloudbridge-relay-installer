package proto

import (
	"bytes"
	"testing"
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

func TestReadMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     RelayMessage
		wantErr bool
	}{
		{
			"register",
			&RegisterMessage{Message: Message{Type: MessageTypeRegister, Token: "test"}},
			false,
		},
		{
			"config",
			&ConfigMessage{Message: Message{Type: MessageTypeConfig}, PublicURL: "http://test"},
			false,
		},
		{
			"error",
			&ErrorMessage{Message: Message{Type: MessageTypeError}, Error: "test error"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write message
			var buf bytes.Buffer
			if err := WriteMessage(&buf, tt.msg); err != nil {
				t.Fatalf("WriteMessage() error = %v", err)
			}

			// Read message
			got, err := ReadMessage(&buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.GetType() != tt.msg.GetType() {
				t.Errorf("ReadMessage().GetType() = %v, want %v", got.GetType(), tt.msg.GetType())
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	errMsg := "test error"
	if err := WriteError(&buf, errMsg); err != nil {
		t.Fatalf("WriteError() error = %v", err)
	}

	msg, err := ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	if errMsg, ok := msg.(*ErrorMessage); !ok || errMsg.Error != errMsg {
		t.Errorf("WriteError() = %v, want %v", msg, errMsg)
	}
}

func TestWritePong(t *testing.T) {
	var buf bytes.Buffer
	token := "test-token"
	if err := WritePong(&buf, token); err != nil {
		t.Fatalf("WritePong() error = %v", err)
	}

	msg, err := ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	if pong, ok := msg.(*PongMessage); !ok || pong.Token != token {
		t.Errorf("WritePong() = %v, want %v", msg, token)
	}
} 