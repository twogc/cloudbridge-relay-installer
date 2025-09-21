package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuicConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  QuicConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: QuicConfig{
				HandshakeTimeout:    "10s",
				IdleTimeout:         "30s",
				MaxStreams:          100,
				ConnectionMigration: true,
				Multiplexing:        true,
			},
			wantErr: false,
		},
		{
			name: "valid minimal config",
			config: QuicConfig{
				MaxStreams: 50,
			},
			wantErr: false,
		},
		{
			name: "empty config - should set default max_streams",
			config: QuicConfig{
				MaxStreams: 0,
			},
			wantErr: false,
		},
		{
			name: "negative max_streams - should set default",
			config: QuicConfig{
				MaxStreams: -10,
			},
			wantErr: false,
		},
		{
			name: "invalid handshake_timeout",
			config: QuicConfig{
				HandshakeTimeout: "invalid",
				MaxStreams:       100,
			},
			wantErr: true,
			errMsg:  "invalid handshake_timeout",
		},
		{
			name: "invalid idle_timeout",
			config: QuicConfig{
				IdleTimeout: "invalid",
				MaxStreams:  100,
			},
			wantErr: true,
			errMsg:  "invalid idle_timeout",
		},
		{
			name: "valid duration formats",
			config: QuicConfig{
				HandshakeTimeout: "5s",
				IdleTimeout:      "1m",
				MaxStreams:       100,
			},
			wantErr: false,
		},
		{
			name: "valid duration with minutes and seconds",
			config: QuicConfig{
				HandshakeTimeout: "2m30s",
				IdleTimeout:      "5m",
				MaxStreams:       100,
			},
			wantErr: false,
		},
		{
			name: "empty timeouts",
			config: QuicConfig{
				HandshakeTimeout: "",
				IdleTimeout:      "",
				MaxStreams:       100,
			},
			wantErr: false,
		},
		{
			name: "zero duration",
			config: QuicConfig{
				HandshakeTimeout: "0s",
				IdleTimeout:      "0s",
				MaxStreams:       100,
			},
			wantErr: false,
		},
		{
			name: "large duration",
			config: QuicConfig{
				HandshakeTimeout: "24h",
				IdleTimeout:      "1h",
				MaxStreams:       100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestQuicConfig_Validate_DefaultMaxStreams(t *testing.T) {
	t.Run("zero max_streams sets default", func(t *testing.T) {
		config := QuicConfig{
			MaxStreams: 0,
		}
		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, 100, config.MaxStreams)
	})

	t.Run("negative max_streams sets default", func(t *testing.T) {
		config := QuicConfig{
			MaxStreams: -50,
		}
		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, 100, config.MaxStreams)
	})

	t.Run("positive max_streams unchanged", func(t *testing.T) {
		config := QuicConfig{
			MaxStreams: 200,
		}
		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, 200, config.MaxStreams)
	})
}

func TestQuicConfig_StructFields(t *testing.T) {
	t.Run("struct field access", func(t *testing.T) {
		config := QuicConfig{
			HandshakeTimeout:    "10s",
			IdleTimeout:         "30s",
			MaxStreams:          150,
			ConnectionMigration: true,
			Multiplexing:        false,
		}

		assert.Equal(t, "10s", config.HandshakeTimeout)
		assert.Equal(t, "30s", config.IdleTimeout)
		assert.Equal(t, 150, config.MaxStreams)
		assert.True(t, config.ConnectionMigration)
		assert.False(t, config.Multiplexing)
	})

	t.Run("empty struct", func(t *testing.T) {
		config := QuicConfig{}

		assert.Empty(t, config.HandshakeTimeout)
		assert.Empty(t, config.IdleTimeout)
		assert.Equal(t, 0, config.MaxStreams)
		assert.False(t, config.ConnectionMigration)
		assert.False(t, config.Multiplexing)
	})

	t.Run("connection migration enabled", func(t *testing.T) {
		config := QuicConfig{
			ConnectionMigration: true,
			MaxStreams:          100,
		}
		assert.True(t, config.ConnectionMigration)
	})

	t.Run("connection migration disabled", func(t *testing.T) {
		config := QuicConfig{
			ConnectionMigration: false,
			MaxStreams:          100,
		}
		assert.False(t, config.ConnectionMigration)
	})

	t.Run("multiplexing enabled", func(t *testing.T) {
		config := QuicConfig{
			Multiplexing: true,
			MaxStreams:   100,
		}
		assert.True(t, config.Multiplexing)
	})

	t.Run("multiplexing disabled", func(t *testing.T) {
		config := QuicConfig{
			Multiplexing: false,
			MaxStreams:   100,
		}
		assert.False(t, config.Multiplexing)
	})
}

func TestQuicConfig_EdgeCases(t *testing.T) {
	t.Run("very small max_streams", func(t *testing.T) {
		config := QuicConfig{
			MaxStreams: 1,
		}
		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, 1, config.MaxStreams)
	})

	t.Run("very large max_streams", func(t *testing.T) {
		config := QuicConfig{
			MaxStreams: 10000,
		}
		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, 10000, config.MaxStreams)
	})

	t.Run("all boolean combinations", func(t *testing.T) {
		testCases := []struct {
			connectionMigration bool
			multiplexing        bool
		}{
			{true, true},
			{true, false},
			{false, true},
			{false, false},
		}

		for _, tc := range testCases {
			t.Run("", func(t *testing.T) {
				config := QuicConfig{
					ConnectionMigration: tc.connectionMigration,
					Multiplexing:        tc.multiplexing,
					MaxStreams:          100,
				}
				err := config.Validate()
				require.NoError(t, err)
				assert.Equal(t, tc.connectionMigration, config.ConnectionMigration)
				assert.Equal(t, tc.multiplexing, config.Multiplexing)
			})
		}
	})

	t.Run("duration edge cases", func(t *testing.T) {
		testCases := []struct {
			handshakeTimeout string
			idleTimeout      string
			shouldPass       bool
		}{
			{"1ns", "1ns", true},
			{"1ms", "1ms", true},
			{"1s", "1s", true},
			{"1m", "1m", true},
			{"1h", "1h", true},
			{"1h30m45s", "2h15m30s", true},
			{"invalid", "1s", false},
			{"1s", "invalid", false},
			{"", "1s", true},
			{"1s", "", true},
			{"", "", true},
		}

		for _, tc := range testCases {
			t.Run("", func(t *testing.T) {
				config := QuicConfig{
					HandshakeTimeout: tc.handshakeTimeout,
					IdleTimeout:      tc.idleTimeout,
					MaxStreams:       100,
				}
				err := config.Validate()
				if tc.shouldPass {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})
}
