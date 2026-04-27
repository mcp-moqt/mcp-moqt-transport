package integration

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func getAvailablePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

func TestMCPServerClient_RunAndPing(t *testing.T) {
	port := getAvailablePort(t)
	serverAddr := fmt.Sprintf("127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serverTransport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleServer,
		mcpmoqt.WithAddr(serverAddr),
	)
	require.NoError(t, err)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.2.0",
	}, nil)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.Run(ctx, serverTransport)
	}()

	time.Sleep(2 * time.Second)

	clientTransport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleClient,
		mcpmoqt.WithAddr(serverAddr),
	)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.2.0",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	err = session.Ping(ctx, nil)
	require.NoError(t, err)

	session.Close()
	cancel()

	select {
	case err := <-serverErrCh:
		t.Logf("server exited: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not exit after client close")
	}
}

func TestMCPServerClient_MultipleConnections(t *testing.T) {
	t.Skip("Multiple connections not supported - server only accepts one connection at a time")

	port := getAvailablePort(t)
	serverAddr := fmt.Sprintf("127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serverTransport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleServer,
		mcpmoqt.WithAddr(serverAddr),
	)
	require.NoError(t, err)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.2.0",
	}, nil)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.Run(ctx, serverTransport)
	}()

	time.Sleep(2 * time.Second)

	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("connection_%d", i), func(t *testing.T) {
			clientTransport, err := mcpmoqt.NewMoqTransport(
				mcpmoqt.RoleClient,
				mcpmoqt.WithAddr(serverAddr),
			)
			require.NoError(t, err)

			client := mcp.NewClient(&mcp.Implementation{
				Name:    fmt.Sprintf("test-client-%d", i),
				Version: "1.2.0",
			}, nil)

			session, err := client.Connect(ctx, clientTransport, nil)
			require.NoError(t, err)
			defer session.Close()

			err = session.Ping(ctx, nil)
			require.NoError(t, err)
		})
	}

	cancel()
	select {
	case err := <-serverErrCh:
		t.Logf("server exited: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not exit after test completion")
	}
}
