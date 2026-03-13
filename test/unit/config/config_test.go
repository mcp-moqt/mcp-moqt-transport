package config

import (
	"os"
	"testing"

	mcpconfig "github.com/mcp-moqt/mcp-moqt-transport/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := mcpconfig.DefaultConfig()

	assert.Equal(t, "localhost:0", config.Addr)
	assert.Equal(t, []string{"moq-00"}, config.ALPN)
	assert.True(t, config.EnableDatagrams)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *mcpconfig.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &mcpconfig.Config{
				Addr:            "127.0.0.1:8080",
				ALPN:            []string{"moq-00"},
				EnableDatagrams: true,
			},
			wantErr: false,
		},
		{
			name: "empty addr",
			config: &mcpconfig.Config{
				Addr:            "",
				ALPN:            []string{"moq-00"},
				EnableDatagrams: true,
			},
			wantErr: true,
		},
		{
			name: "empty ALPN",
			config: &mcpconfig.Config{
				Addr:            "127.0.0.1:8080",
				ALPN:            []string{},
				EnableDatagrams: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original env vars
	origAddr := os.Getenv("MCP_MOQT_ADDR")
	origALPN := os.Getenv("MCP_MOQT_ALPN")
	origDatagrams := os.Getenv("MCP_MOQT_ENABLE_DATAGRAMS")

	defer func() {
		os.Setenv("MCP_MOQT_ADDR", origAddr)
		os.Setenv("MCP_MOQT_ALPN", origALPN)
		os.Setenv("MCP_MOQT_ENABLE_DATAGRAMS", origDatagrams)
	}()

	// Test with custom env vars
	os.Setenv("MCP_MOQT_ADDR", "192.168.1.1:9000")
	os.Setenv("MCP_MOQT_ALPN", "moq-01,moq-02")
	os.Setenv("MCP_MOQT_ENABLE_DATAGRAMS", "false")

	config := mcpconfig.LoadFromEnv()

	assert.Equal(t, "192.168.1.1:9000", config.Addr)
	assert.Equal(t, []string{"moq-01", "moq-02"}, config.ALPN)
	assert.False(t, config.EnableDatagrams)
}

func TestLoadFromEnv_Empty(t *testing.T) {
	// Save original env vars
	origAddr := os.Getenv("MCP_MOQT_ADDR")
	origALPN := os.Getenv("MCP_MOQT_ALPN")
	origDatagrams := os.Getenv("MCP_MOQT_ENABLE_DATAGRAMS")

	defer func() {
		os.Setenv("MCP_MOQT_ADDR", origAddr)
		os.Setenv("MCP_MOQT_ALPN", origALPN)
		os.Setenv("MCP_MOQT_ENABLE_DATAGRAMS", origDatagrams)
	}()

	// Clear env vars
	os.Unsetenv("MCP_MOQT_ADDR")
	os.Unsetenv("MCP_MOQT_ALPN")
	os.Unsetenv("MCP_MOQT_ENABLE_DATAGRAMS")

	config := mcpconfig.LoadFromEnv()

	// Should use defaults
	assert.Equal(t, "localhost:0", config.Addr)
	assert.Equal(t, []string{"moq-00"}, config.ALPN)
	assert.True(t, config.EnableDatagrams)
}
