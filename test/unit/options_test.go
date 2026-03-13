package config

import (
	"testing"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/require"
)

func TestWithAddr(t *testing.T) {
	addr := "127.0.0.1:8080"
	transport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleClient,
		mcpmoqt.WithAddr(addr),
	)
	require.NoError(t, err)
	// We can't directly access the addr field, but we can verify the transport is created
	require.NotNil(t, transport)
}

func TestWithALPN(t *testing.T) {
	alpn := []string{"moq-01"}
	transport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleClient,
		mcpmoqt.WithALPN(alpn...),
	)
	require.NoError(t, err)
	require.NotNil(t, transport)
}

func TestWithALPNEmpty(t *testing.T) {
	_, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleClient,
		mcpmoqt.WithALPN(),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "alpn must not be empty")
}

func TestWithAddrEmpty(t *testing.T) {
	_, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleClient,
		mcpmoqt.WithAddr(""),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "addr must not be empty")
}
