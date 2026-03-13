package unit

import (
	"testing"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoqTransport_Server(t *testing.T) {
	transport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleServer,
		mcpmoqt.WithAddr("127.0.0.1:0"),
	)
	require.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestNewMoqTransport_Client(t *testing.T) {
	transport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleClient,
		mcpmoqt.WithAddr("127.0.0.1:0"),
	)
	require.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestNewMoqTransport_InvalidRole(t *testing.T) {
	_, err := mcpmoqt.NewMoqTransport(999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown role")
}

func TestWithMultipleOptions(t *testing.T) {
	transport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleServer,
		mcpmoqt.WithAddr("127.0.0.1:8080"),
		mcpmoqt.WithALPN("moq-01"),
	)
	require.NoError(t, err)
	assert.NotNil(t, transport)
}
