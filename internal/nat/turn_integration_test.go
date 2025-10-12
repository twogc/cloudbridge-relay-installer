package nat

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewTURNIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	assert.NotNil(t, ti)
	assert.NotNil(t, ti.logger)
	assert.Empty(t, ti.stunServers)
	assert.Empty(t, ti.turnServers)
}

func TestLoadICEServersFromConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	// Create temporary config file
	config := ICEServerConfig{
		STUNServers: []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
		},
		TURNServers: []TURNServer{
			{
				URLs:       "turn:turn.example.com:3478?transport=udp",
				Username:   "testuser",
				Credential: "testpass",
				Realm:      "example.com",
			},
		},
	}

	tmpFile, err := os.CreateTemp("", "ice-servers-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	configData, err := json.Marshal(config)
	require.NoError(t, err)

	_, err = tmpFile.Write(configData)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	err = ti.LoadICEServersFromConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 2, len(ti.GetSTUNServers()))
	assert.Equal(t, 1, len(ti.GetTURNServers()))
	assert.Equal(t, "stun:stun.l.google.com:19302", ti.GetSTUNServers()[0])
	assert.Equal(t, "testuser", ti.GetTURNServers()[0].Username)
}

func TestLoadICEServersFromConfig_WithEnvVarSubstitution(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	// Set environment variable
	os.Setenv("TURN_PASSWORD", "secret-password")
	defer os.Unsetenv("TURN_PASSWORD")

	// Create config with placeholder
	config := ICEServerConfig{
		STUNServers: []string{"stun:stun.example.com:19302"},
		TURNServers: []TURNServer{
			{
				URLs:       "turn:turn.example.com:3478",
				Username:   "user",
				Credential: "${TURN_PASSWORD}",
				Realm:      "example.com",
			},
		},
	}

	tmpFile, err := os.CreateTemp("", "ice-servers-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	configData, err := json.Marshal(config)
	require.NoError(t, err)

	_, err = tmpFile.Write(configData)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	err = ti.LoadICEServersFromConfig(tmpFile.Name())
	require.NoError(t, err)

	// Verify password was substituted
	assert.Equal(t, "secret-password", ti.GetTURNServers()[0].Credential)
}

func TestLoadFromEnvironment_WithDefaults(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	// Clear any existing env vars
	os.Unsetenv("STUN_SERVERS")
	os.Unsetenv("TURN_URL")
	os.Unsetenv("TURN_USERNAME")
	os.Unsetenv("TURN_PASSWORD")
	os.Unsetenv("TURN_REALM")

	err := ti.LoadFromEnvironment()
	require.NoError(t, err)

	// Should have default STUN servers
	assert.GreaterOrEqual(t, len(ti.GetSTUNServers()), 1)
	assert.Contains(t, ti.GetSTUNServers()[0], "stun:")

	// No TURN servers without config
	assert.Equal(t, 0, len(ti.GetTURNServers()))
}

func TestLoadFromEnvironment_WithTURNConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	// Set TURN environment variables
	os.Setenv("TURN_URL", "turn:192.168.1.100:3478?transport=udp,turns:192.168.1.100:5349?transport=tcp")
	os.Setenv("TURN_USERNAME", "cloudbridge")
	os.Setenv("TURN_PASSWORD", "test-password")
	os.Setenv("TURN_REALM", "cloudbridge.2gc.ru")
	defer func() {
		os.Unsetenv("TURN_URL")
		os.Unsetenv("TURN_USERNAME")
		os.Unsetenv("TURN_PASSWORD")
		os.Unsetenv("TURN_REALM")
	}()

	err := ti.LoadFromEnvironment()
	require.NoError(t, err)

	// Verify TURN servers
	assert.Equal(t, 2, len(ti.GetTURNServers()))
	assert.Equal(t, "turn:192.168.1.100:3478?transport=udp", ti.GetTURNServers()[0].URLs)
	assert.Equal(t, "turns:192.168.1.100:5349?transport=tcp", ti.GetTURNServers()[1].URLs)
	assert.Equal(t, "cloudbridge", ti.GetTURNServers()[0].Username)
	assert.Equal(t, "test-password", ti.GetTURNServers()[0].Credential)
	assert.Equal(t, "cloudbridge.2gc.ru", ti.GetTURNServers()[0].Realm)
}

func TestLoadFromEnvironment_CustomSTUNServers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	// Set custom STUN servers
	os.Setenv("STUN_SERVERS", "stun:custom1.com:19302, stun:custom2.com:19302")
	defer os.Unsetenv("STUN_SERVERS")

	err := ti.LoadFromEnvironment()
	require.NoError(t, err)

	assert.Equal(t, 2, len(ti.GetSTUNServers()))
	assert.Equal(t, "stun:custom1.com:19302", ti.GetSTUNServers()[0])
	assert.Equal(t, "stun:custom2.com:19302", ti.GetSTUNServers()[1])
}

func TestGetICEServersForClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{
		"stun:stun.example.com:19302",
	}
	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478",
			Username:   "user1",
			Credential: "pass1",
			Realm:      "example.com",
		},
	}

	result := ti.GetICEServersForClient()

	assert.Contains(t, result, "iceServers")
	assert.Contains(t, result, "iceTransportPolicy")
	assert.Equal(t, "all", result["iceTransportPolicy"])

	iceServers := result["iceServers"].([]map[string]interface{})
	assert.Equal(t, 2, len(iceServers))

	// Check STUN server
	assert.Equal(t, "stun:stun.example.com:19302", iceServers[0]["urls"])
	assert.NotContains(t, iceServers[0], "username")

	// Check TURN server
	assert.Equal(t, "turn:turn.example.com:3478", iceServers[1]["urls"])
	assert.Equal(t, "user1", iceServers[1]["username"])
	assert.Equal(t, "pass1", iceServers[1]["credential"])
}

func TestValidateConfiguration_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{"stun:stun.example.com:19302"}
	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478",
			Username:   "user",
			Credential: "pass",
			Realm:      "example.com",
		},
	}

	err := ti.ValidateConfiguration()
	assert.NoError(t, err)
}

func TestValidateConfiguration_NoServers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	err := ti.ValidateConfiguration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no ICE servers configured")
}

func TestValidateConfiguration_MissingUsername(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{"stun:stun.example.com:19302"}
	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478",
			Username:   "", // Missing username
			Credential: "pass",
			Realm:      "example.com",
		},
	}

	err := ti.ValidateConfiguration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing username")
}

func TestValidateConfiguration_MissingCredential(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{"stun:stun.example.com:19302"}
	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478",
			Username:   "user",
			Credential: "", // Missing credential
			Realm:      "example.com",
		},
	}

	err := ti.ValidateConfiguration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing credential")
}

func TestGetConfigSummary(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{
		"stun:stun1.example.com:19302",
		"stun:stun2.example.com:19302",
	}
	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478",
			Username:   "user",
			Credential: "secret-password",
			Realm:      "example.com",
		},
	}

	summary := ti.GetConfigSummary()

	assert.Equal(t, 2, summary["stun_servers_count"])
	assert.Equal(t, 1, summary["turn_servers_count"])

	turnSummary := summary["turn_servers"].([]map[string]interface{})
	assert.Equal(t, 1, len(turnSummary))
	assert.Equal(t, "turn:turn.example.com:3478", turnSummary[0]["url"])
	assert.Equal(t, "user", turnSummary[0]["username"])
	// Credential should not be in summary
	assert.NotContains(t, turnSummary[0], "credential")
}

func TestIntegrateWithPionICE(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{
		"stun:stun.example.com:19302",
	}
	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478?transport=udp",
			Username:   "user",
			Credential: "pass",
		},
		{
			URLs:       "turns:turn.example.com:5349?transport=tcp",
			Username:   "user",
			Credential: "pass",
		},
	}

	urls, err := ti.IntegrateWithPionICE()
	require.NoError(t, err)

	assert.Equal(t, 3, len(urls))
	assert.Equal(t, "stun:stun.example.com:19302", urls[0])
	assert.Equal(t, "turn:turn.example.com:3478?transport=udp", urls[1])
	assert.Equal(t, "turns:turn.example.com:5349?transport=tcp", urls[2])
}

func TestGetTURNCredentialsForPeer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478",
			Username:   "cloudbridge",
			Credential: "test-password",
			Realm:      "example.com",
		},
	}

	username, password, err := ti.GetTURNCredentialsForPeer("peer-123")
	require.NoError(t, err)

	assert.Contains(t, username, "cloudbridge")
	assert.Contains(t, username, "peer-123")
	assert.Equal(t, "test-password", password)
}

func TestGetTURNCredentialsForPeer_NoServers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	_, _, err := ti.GetTURNCredentialsForPeer("peer-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no TURN servers configured")
}

func TestHealthCheck_Healthy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{"stun:stun.example.com:19302"}
	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478",
			Username:   "user",
			Credential: "pass",
			Realm:      "example.com",
		},
	}

	health := ti.HealthCheck()

	assert.Equal(t, "turn_integration", health["component"])
	assert.Equal(t, "healthy", health["status"])
	assert.Equal(t, 1, health["stun_servers"])
	assert.Equal(t, 1, health["turn_servers"])
}

func TestHealthCheck_Degraded_NoTURN(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{"stun:stun.example.com:19302"}
	// No TURN servers

	health := ti.HealthCheck()

	assert.Equal(t, "degraded", health["status"])
	issues := health["issues"].([]string)
	assert.Contains(t, issues, "no TURN servers configured")
}

func TestHealthCheck_Unhealthy_MissingCredentials(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{"stun:stun.example.com:19302"}
	ti.turnServers = []TURNServer{
		{
			URLs:       "turn:turn.example.com:3478",
			Username:   "user",
			Credential: "", // Missing credential
			Realm:      "example.com",
		},
	}

	health := ti.HealthCheck()

	assert.Equal(t, "unhealthy", health["status"])
	issues := health["issues"].([]string)
	assert.Contains(t, issues, "TURN server 0 missing credentials")
}

func BenchmarkLoadFromEnvironment(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	os.Setenv("TURN_URL", "turn:192.168.1.100:3478")
	os.Setenv("TURN_USERNAME", "user")
	os.Setenv("TURN_PASSWORD", "pass")
	defer func() {
		os.Unsetenv("TURN_URL")
		os.Unsetenv("TURN_USERNAME")
		os.Unsetenv("TURN_PASSWORD")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ti := NewTURNIntegration(logger)
		ti.LoadFromEnvironment()
	}
}

func BenchmarkGetICEServersForClient(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	ti := NewTURNIntegration(logger)

	ti.stunServers = []string{
		"stun:stun1.example.com:19302",
		"stun:stun2.example.com:19302",
	}
	ti.turnServers = []TURNServer{
		{URLs: "turn:turn1.example.com:3478", Username: "user", Credential: "pass"},
		{URLs: "turn:turn2.example.com:3478", Username: "user", Credential: "pass"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ti.GetICEServersForClient()
	}
}
