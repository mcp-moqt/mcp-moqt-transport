package integration

import (
	"context"
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMCPServerClient_RunAndPing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start server in a goroutine
	serverAddr := "127.0.0.1:8080"
	serverTransport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleServer,
		mcpmoqt.WithAddr(serverAddr),
	)
	require.NoError(t, err)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "v0.2.0",
	}, nil)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.Run(ctx, serverTransport)
	}()

	// Give server time to start
	time.Sleep(1 * time.Second)

	// Create client and connect
	clientTransport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleClient,
		mcpmoqt.WithAddr(serverAddr),
	)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "v0.2.0",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	// Ping server
	err = session.Ping(ctx, nil)
	require.NoError(t, err)

	// Wait for server to exit (since it's a single connection server)
	select {
	case err := <-serverErrCh:
		// Server should exit with context canceled or similar
		require.Error(t, err)
	case <-ctx.Done():
		t.Fatal("test timed out")
	}
}