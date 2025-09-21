package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIceConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  IceConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: IceConfig{
				StunServers:       []string{"stun1.example.com:3478", "stun2.example.com:3478"},
				TurnServers:       []string{"turn1.example.com:3478"},
				IceTimeout:        "30s",
				ConnectionTimeout: "10s",
				GatheringTimeout:  "5s",
				CandidateTypes:    []string{"host", "srflx", "relay"},
			},
			wantErr: false,
		},
		{
			name: "valid config with minimal fields",
			config: IceConfig{
				StunServers: []string{"stun.example.com:3478"},
			},
			wantErr: false,
		},
		{
			name: "empty stun servers",
			config: IceConfig{
				StunServers: []string{},
			},
			wantErr: true,
			errMsg:  "at least one STUN server is required",
		},
		{
			name: "nil stun servers",
			config: IceConfig{
				StunServers: nil,
			},
			wantErr: true,
			errMsg:  "at least one STUN server is required",
		},
		{
			name: "invalid ice timeout",
			config: IceConfig{
				StunServers: []string{"stun.example.com:3478"},
				IceTimeout:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid ice_timeout",
		},
		{
			name: "invalid connection timeout",
			config: IceConfig{
				StunServers:       []string{"stun.example.com:3478"},
				ConnectionTimeout: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid connection_timeout",
		},
		{
			name: "invalid gathering timeout",
			config: IceConfig{
				StunServers:      []string{"stun.example.com:3478"},
				GatheringTimeout: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid gathering_timeout",
		},
		{
			name: "valid duration formats",
			config: IceConfig{
				StunServers:       []string{"stun.example.com:3478"},
				IceTimeout:        "30s",
				ConnectionTimeout: "10s",
				GatheringTimeout:  "5s",
			},
			wantErr: false,
		},
		{
			name: "valid duration with minutes",
			config: IceConfig{
				StunServers:       []string{"stun.example.com:3478"},
				IceTimeout:        "2m30s",
				ConnectionTimeout: "1m",
				GatheringTimeout:  "30s",
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

func TestIceConfig_Validate_EdgeCases(t *testing.T) {
	t.Run("empty strings for timeouts", func(t *testing.T) {
		config := IceConfig{
			StunServers:       []string{"stun.example.com:3478"},
			IceTimeout:        "",
			ConnectionTimeout: "",
			GatheringTimeout:  "",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("zero duration", func(t *testing.T) {
		config := IceConfig{
			StunServers: []string{"stun.example.com:3478"},
			IceTimeout:  "0s",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("very large duration", func(t *testing.T) {
		config := IceConfig{
			StunServers: []string{"stun.example.com:3478"},
			IceTimeout:  "24h",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})
}

func TestIceConfig_CandidateTypesValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  IceConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid candidate types",
			config: IceConfig{
				StunServers:    []string{"stun.example.com:3478"},
				CandidateTypes: []string{"host", "srflx", "relay"},
			},
			wantErr: false,
		},
		{
			name: "valid candidate types with prflx",
			config: IceConfig{
				StunServers:    []string{"stun.example.com:3478"},
				CandidateTypes: []string{"host", "srflx", "relay", "prflx"},
			},
			wantErr: false,
		},
		{
			name: "empty candidate types",
			config: IceConfig{
				StunServers:    []string{"stun.example.com:3478"},
				CandidateTypes: []string{},
			},
			wantErr: false,
		},
		{
			name: "invalid candidate type",
			config: IceConfig{
				StunServers:    []string{"stun.example.com:3478"},
				CandidateTypes: []string{"invalid"},
			},
			wantErr: true,
			errMsg:  "invalid candidate_type: invalid",
		},
		{
			name: "mixed valid and invalid candidate types",
			config: IceConfig{
				StunServers:    []string{"stun.example.com:3478"},
				CandidateTypes: []string{"host", "invalid", "relay"},
			},
			wantErr: true,
			errMsg:  "invalid candidate_type: invalid",
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

func TestIceConfig_StructFields(t *testing.T) {
	t.Run("struct field access", func(t *testing.T) {
		config := IceConfig{
			StunServers:       []string{"stun1.example.com:3478", "stun2.example.com:3478"},
			TurnServers:       []string{"turn1.example.com:3478"},
			IceTimeout:        "30s",
			ConnectionTimeout: "10s",
			GatheringTimeout:  "5s",
			CandidateTypes:    []string{"host", "srflx"},
		}

		// Test field access
		assert.Len(t, config.StunServers, 2)
		assert.Len(t, config.TurnServers, 1)
		assert.Equal(t, "30s", config.IceTimeout)
		assert.Equal(t, "10s", config.ConnectionTimeout)
		assert.Equal(t, "5s", config.GatheringTimeout)
		assert.Len(t, config.CandidateTypes, 2)
	})

	t.Run("empty struct", func(t *testing.T) {
		config := IceConfig{}

		assert.Nil(t, config.StunServers)
		assert.Nil(t, config.TurnServers)
		assert.Empty(t, config.IceTimeout)
		assert.Empty(t, config.ConnectionTimeout)
		assert.Empty(t, config.GatheringTimeout)
		assert.Nil(t, config.CandidateTypes)
	})
}
