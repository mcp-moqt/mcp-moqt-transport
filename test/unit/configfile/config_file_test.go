package configfile

import (
	"os"
	"path/filepath"
	"testing"

	mcpconfig "github.com/mcp-moqt/mcp-moqt-transport/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromFile_YAML(t *testing.T) {
	content := `addr: "192.168.1.1:8080"
alpn:
  - "moq-01"
  - "moq-02"
enable_datagrams: false
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	config, err := mcpconfig.LoadFromFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.1:8080", config.Addr)
	assert.Equal(t, []string{"moq-01", "moq-02"}, config.ALPN)
	assert.False(t, config.EnableDatagrams)
}

func TestLoadFromFile_JSON(t *testing.T) {
	content := `{
  "addr": "10.0.0.1:9000",
  "alpn": ["moq-03"],
  "enable_datagrams": true
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	config, err := mcpconfig.LoadFromFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1:9000", config.Addr)
	assert.Equal(t, []string{"moq-03"}, config.ALPN)
	assert.True(t, config.EnableDatagrams)
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := mcpconfig.LoadFromFile("nonexistent.yaml")
	require.Error(t, err)
}

func TestConfig_SaveToFile_YAML(t *testing.T) {
	config := &mcpconfig.Config{
		Addr:            "127.0.0.1:8080",
		ALPN:            []string{"moq-00"},
		EnableDatagrams: true,
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.yaml")

	err := config.SaveToFile(tmpFile)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "addr: 127.0.0.1:8080")
}

func TestConfig_SaveToFile_JSON(t *testing.T) {
	config := &mcpconfig.Config{
		Addr:            "127.0.0.1:8080",
		ALPN:            []string{"moq-00"},
		EnableDatagrams: true,
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.json")

	err := config.SaveToFile(tmpFile)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"addr"`)
}
