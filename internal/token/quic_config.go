package token

import (
	"fmt"
	"time"
)

// QuicConfig represents QUIC configuration
type QuicConfig struct {
	HandshakeTimeout    string `json:"handshake_timeout"`
	IdleTimeout         string `json:"idle_timeout"`
	MaxStreams          int    `json:"max_streams"`
	ConnectionMigration bool   `json:"connection_migration"`
	Multiplexing        bool   `json:"multiplexing"`
}

// Validate validates QUIC configuration
func (q *QuicConfig) Validate() error {
	if q.HandshakeTimeout != "" {
		if _, err := time.ParseDuration(q.HandshakeTimeout); err != nil {
			return fmt.Errorf("invalid handshake_timeout: %v", err)
		}
	}

	if q.IdleTimeout != "" {
		if _, err := time.ParseDuration(q.IdleTimeout); err != nil {
			return fmt.Errorf("invalid idle_timeout: %v", err)
		}
	}

	if q.MaxStreams <= 0 {
		q.MaxStreams = 100 // default value
	}

	return nil
}

