package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	appconfig "github.com/mcp-moqt/mcp-moqt-transport/pkg/config"
	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func goVersionHint() string {
	return "Go 1.25.13+ recommended (see go.mod)"
}

func probeQUICServer(ctx context.Context, addr string) error {
	if addr == "" {
		return nil
	}
	transport, err := mcpmoqt.NewMoqTransport(mcpmoqt.RoleClient, mcpmoqt.WithAddr(addr))
	if err != nil {
		return err
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "doctor", Version: mcpservice.ServiceVersion}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return err
	}
	defer session.Close()
	return session.Ping(ctx, nil)
}

func probeMetricsEndpoint(ctx context.Context, addr string) error {
	if addr == "" {
		return nil
	}
	url := "http://" + addr + "/healthz"
	if strings.HasPrefix(addr, "http://") {
		url = strings.TrimSuffix(addr, "/") + "/healthz"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthz status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func checkTLSCertPair(certFile, keyFile string) error {
	if certFile == "" && keyFile == "" {
		return nil
	}
	return appconfig.ValidateCertPair(certFile, keyFile)
}

func checkUDPPortBind(addr string) error {
	if addr == "" {
		return nil
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if host == "" {
		host = "127.0.0.1"
	}
	ln, err := net.ListenPacket("udp", net.JoinHostPort(host, port))
	if err != nil {
		return fmt.Errorf("udp bind %s: %w", addr, err)
	}
	_ = ln.Close()
	return nil
}

func doctorProbeTimeout() time.Duration {
	return 8 * time.Second
}
