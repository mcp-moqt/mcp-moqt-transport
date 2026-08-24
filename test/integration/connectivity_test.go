package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/mcp-moqt/mcp-moqt-transport/pkg/config"
	"github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"
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
		Version: mcpservice.ServiceVersion,
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
		Version: mcpservice.ServiceVersion,
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
	port := getAvailablePort(t)
	serverAddr := fmt.Sprintf("127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	raw, err := mcpmoqt.NewMOQTServerTransport(mcpmoqt.WithAddr(serverAddr))
	require.NoError(t, err)
	defer raw.Close()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- raw.Serve(ctx, func(ctx context.Context, conn mcpmoqt.Connection) error {
			server := mcp.NewServer(&mcp.Implementation{
				Name:    "test-server",
				Version: mcpservice.ServiceVersion,
			}, nil)
			ss, err := server.Connect(ctx, &mcpmoqt.PreconnectedTransport{Conn: conn}, nil)
			if err != nil {
				return err
			}
			defer ss.Close()
			return ss.Wait()
		})
	}()

	time.Sleep(500 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			clientTransport, err := mcpmoqt.NewMOQTClientTransport(mcpmoqt.WithAddr(serverAddr))
			require.NoError(t, err)
			defer clientTransport.Close()

			client := mcp.NewClient(&mcp.Implementation{
				Name:    fmt.Sprintf("test-client-%d", i),
				Version: mcpservice.ServiceVersion,
			}, nil)

			session, err := client.Connect(ctx, clientTransport, nil)
			require.NoError(t, err)
			defer session.Close()

			require.NoError(t, session.Ping(ctx, nil))
		}(i)
	}
	wg.Wait()

	cancel()
	select {
	case err := <-serveErr:
		t.Logf("serve exited: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not exit")
	}
}

func TestMetricsHTTP_PrometheusExposition(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	_ = ln.Close()

	m := mcpmoqt.NewMetrics()
	m.RecordMessageSent(10)
	m.RecordConnectionOpened()

	srv := mcpmoqt.NewMetricsHTTPServer(addr, m)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()
	defer srv.Close()

	deadline := time.Now().Add(3 * time.Second)
	var resp *http.Response
	for time.Now().Before(deadline) {
		resp, err = http.Get("http://" + srv.Addr() + "/metrics")
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "mcp_moqt_messages_sent_total")
	require.Contains(t, string(body), "# TYPE mcp_moqt_messages_sent_total counter")
}

func TestConfigHotReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	require.NoError(t, os.WriteFile(path, []byte("addr: \"127.0.0.1:8080\"\nalpn: [\"moq-00\"]\nenable_datagrams: true\n"), 0644))

	var mu sync.Mutex
	var configs []*config.Config
	w, err := config.WatchFile(path, 50*time.Millisecond, func(cfg *config.Config) {
		mu.Lock()
		configs = append(configs, cfg)
		mu.Unlock()
	})
	require.NoError(t, err)
	defer w.Stop()

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(configs) >= 1
	}, 2*time.Second, 20*time.Millisecond)

	require.NoError(t, os.WriteFile(path, []byte("addr: \"127.0.0.1:9090\"\nalpn: [\"moq-00\"]\nenable_datagrams: false\n"), 0644))

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		if len(configs) < 2 {
			return false
		}
		last := configs[len(configs)-1]
		return last.Addr == "127.0.0.1:9090" && !last.EnableDatagrams
	}, 3*time.Second, 50*time.Millisecond)
}
