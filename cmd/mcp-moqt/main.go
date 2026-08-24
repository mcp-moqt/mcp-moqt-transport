// Command mcp-moqt is a friendly CLI for installing, running, and diagnosing
// the MCP over MOQT transport service.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	appconfig "github.com/mcp-moqt/mcp-moqt-transport/pkg/config"
	"github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"
	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("%s %s\n", mcpservice.ServiceName, mcpservice.ServiceVersion)
	case "server":
		runServer(os.Args[2:])
	case "client":
		runClient(os.Args[2:])
	case "doctor":
		runDoctor(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Print(`mcp-moqt — MCP over MOQT 友好工具

Usage:
  mcp-moqt <command> [flags]

Commands:
  version   显示版本
  server    启动 MCP 服务（支持 -stdio 或 QUIC/MOQT）
  client    连接 QUIC/MOQT 服务并演示 capabilities
  doctor    本地环境自检 + capabilities 冒烟（in-memory）
  help      显示帮助

Examples:
  mcp-moqt server -stdio
  mcp-moqt server -addr 127.0.0.1:8080
  mcp-moqt server -addr 127.0.0.1:8080 -multi -metrics 127.0.0.1:9090
  mcp-moqt client -addr 127.0.0.1:8080
  mcp-moqt doctor
  mcp-moqt doctor -addr 127.0.0.1:8080 -metrics 127.0.0.1:9090
  mcp-moqt doctor -tls-cert server.crt -tls-key server.key
  mcp-moqt server -config configs/config.production.yaml
`)
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "QUIC listen address (ignored with -stdio)")
	configPath := fs.String("config", "", "YAML/JSON config file (overrides -addr/-metrics TLS)")
	stdio := fs.Bool("stdio", false, "serve over stdin/stdout (for Cursor / MCP hosts)")
	metrics := fs.String("metrics", "", "Prometheus /metrics listen address (e.g. 127.0.0.1:9090)")
	multi := fs.Bool("multi", false, "accept multiple QUIC clients (Serve loop)")
	_ = fs.Parse(args)

	opts := []mcpmoqt.Option{}
	if *configPath != "" {
		cfg, err := appconfig.LoadFromFile(*configPath)
		if err != nil {
			fatalf("load config: %v", err)
		}
		if err := cfg.Validate(); err != nil {
			fatalf("invalid config: %v", err)
		}
		opts = append(opts, mcpmoqt.WithConfig(cfg))
	} else {
		opts = append(opts, mcpmoqt.WithAddr(*addr))
		if *metrics != "" {
			opts = append(opts, mcpmoqt.WithMetricsAddr(*metrics))
		}
	}

	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{
		Name:    mcpservice.ServiceName,
		Version: mcpservice.ServiceVersion,
	}, nil)
	mcpservice.RegisterCapabilities(server)

	if *stdio {
		fmt.Fprintf(os.Stderr, "mcp-moqt server (stdio) ready; tools/prompts/resources enabled\n")
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			fatalf("server run: %v", err)
		}
		return
	}

	if *multi {
		raw, err := mcpmoqt.NewMOQTServerTransport(opts...)
		if err != nil {
			fatalf("new transport: %v", err)
		}
		fmt.Printf("listening (multi-accept; metrics=%q)\n", *metrics)
		if err := raw.Serve(ctx, func(ctx context.Context, conn mcpmoqt.Connection) error {
			s := mcp.NewServer(&mcp.Implementation{
				Name:    mcpservice.ServiceName,
				Version: mcpservice.ServiceVersion,
			}, nil)
			mcpservice.RegisterCapabilities(s)
			ss, err := s.Connect(ctx, &mcpmoqt.PreconnectedTransport{Conn: conn}, nil)
			if err != nil {
				return err
			}
			defer ss.Close()
			return ss.Wait()
		}); err != nil {
			fatalf("serve: %v", err)
		}
		return
	}

	transport, err := mcpmoqt.NewMoqTransport(mcpmoqt.RoleServer, opts...)
	if err != nil {
		fatalf("new transport: %v", err)
	}

	fmt.Printf("listening (QUIC/MOQT; tools/prompts/resources enabled; metrics=%q)\n", *metrics)
	if err := server.Run(ctx, transport); err != nil {
		fatalf("server run: %v", err)
	}
}

func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "server address")
	_ = fs.Parse(args)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	transport, err := mcpmoqt.NewMoqTransport(mcpmoqt.RoleClient, mcpmoqt.WithAddr(*addr))
	if err != nil {
		fatalf("new transport: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "mcp-moqt-cli",
		Version: mcpservice.ServiceVersion,
	}, nil)

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		fatalf("connect: %v", err)
	}
	defer session.Close()

	if err := session.Ping(ctx, nil); err != nil {
		fatalf("ping: %v", err)
	}
	fmt.Println("ping: ok")
	printCapabilities(ctx, session)
}

func runDoctor(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	skipCaps := fs.Bool("skip-capabilities", false, "skip in-memory tools/prompts/resources smoke test")
	addr := fs.String("addr", "", "probe QUIC/MOQT server (e.g. 127.0.0.1:8080)")
	metricsAddr := fs.String("metrics", "", "probe Prometheus /healthz (e.g. 127.0.0.1:9090)")
	tlsCert := fs.String("tls-cert", "", "validate TLS certificate file")
	tlsKey := fs.String("tls-key", "", "validate TLS private key file")
	skipUDP := fs.Bool("skip-udp-bind", false, "skip local UDP port bind check for -addr")
	_ = fs.Parse(args)

	checks := []struct {
		name string
		ok   bool
		hint string
	}{
		{"go toolchain", commandExists("go"), goVersionHint()},
		{"service package", true, "pkg/mcpservice (tools/prompts/resources)"},
		{"CLI entrypoint", true, "cmd/mcp-moqt"},
		{"stdio mode", true, "mcp-moqt server -stdio"},
		{"install scripts", true, "scripts/install.sh + scripts/install.ps1"},
		{"docker compose", true, "docker compose up --build"},
		{"mcp config samples", true, "configs/mcp/*.json"},
		{"deployment docs", true, "docs/deployment.md"},
		{"grafana sample", true, "configs/grafana/dashboard.json"},
	}

	fmt.Printf("doctor: %s %s\n", mcpservice.ServiceName, mcpservice.ServiceVersion)
	allOK := true
	for _, c := range checks {
		status := "ok"
		if !c.ok {
			status = "fail"
			allOK = false
		}
		fmt.Printf("  [%s] %s — %s\n", status, c.name, c.hint)
	}

	if !*skipCaps {
		if err := smokeCapabilities(); err != nil {
			fmt.Printf("  [fail] capabilities smoke — %v\n", err)
			allOK = false
		} else {
			fmt.Printf("  [ok] capabilities smoke — tools/prompts/resources list+call/read\n")
		}
	}

	if *tlsCert != "" || *tlsKey != "" {
		if err := checkTLSCertPair(*tlsCert, *tlsKey); err != nil {
			fmt.Printf("  [fail] tls cert pair — %v\n", err)
			allOK = false
		} else {
			fmt.Printf("  [ok] tls cert pair — valid and not expired\n")
		}
	}

	if *addr != "" && !*skipUDP {
		if err := checkUDPPortBind(*addr); err != nil {
			fmt.Printf("  [warn] udp bind %s — %v (port may be in use; use -skip-udp-bind to ignore)\n", *addr, err)
		} else {
			fmt.Printf("  [ok] udp bind %s — port available for listen\n", *addr)
		}
	}

	if *addr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), doctorProbeTimeout())
		defer cancel()
		if err := probeQUICServer(ctx, *addr); err != nil {
			fmt.Printf("  [fail] quic probe %s — %v\n", *addr, err)
			allOK = false
		} else {
			fmt.Printf("  [ok] quic probe %s — connect + ping\n", *addr)
		}
	}

	if *metricsAddr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), doctorProbeTimeout())
		defer cancel()
		if err := probeMetricsEndpoint(ctx, *metricsAddr); err != nil {
			fmt.Printf("  [fail] metrics probe %s — %v\n", *metricsAddr, err)
			allOK = false
		} else {
			fmt.Printf("  [ok] metrics probe %s — /healthz ok\n", *metricsAddr)
		}
	}

	if !allOK {
		os.Exit(1)
	}
	fmt.Println("environment looks ready")
}

func smokeCapabilities() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    mcpservice.ServiceName,
		Version: mcpservice.ServiceVersion,
	}, nil)
	mcpservice.RegisterCapabilities(server)

	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		return fmt.Errorf("server connect: %w", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "doctor", Version: mcpservice.ServiceVersion}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		return fmt.Errorf("client connect: %w", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) < 1 {
		return fmt.Errorf("list tools: %v (count=%d)", err, lenOrZeroTools(tools))
	}
	prompts, err := session.ListPrompts(ctx, nil)
	if err != nil || len(prompts.Prompts) < 1 {
		return fmt.Errorf("list prompts failed")
	}
	resources, err := session.ListResources(ctx, nil)
	if err != nil || len(resources.Resources) < 1 {
		return fmt.Errorf("list resources failed")
	}
	if _, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "health_check"}); err != nil {
		return fmt.Errorf("health_check: %w", err)
	}
	if _, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "moqt://docs/overview"}); err != nil {
		return fmt.Errorf("read resource: %w", err)
	}
	return nil
}

func lenOrZeroTools(r *mcp.ListToolsResult) int {
	if r == nil {
		return 0
	}
	return len(r.Tools)
}

func printCapabilities(ctx context.Context, session *mcp.ClientSession) {
	if tools, err := session.ListTools(ctx, nil); err == nil {
		names := make([]string, 0, len(tools.Tools))
		for _, t := range tools.Tools {
			names = append(names, t.Name)
		}
		fmt.Printf("tools (%d): %s\n", len(names), strings.Join(names, ", "))
	}
	if prompts, err := session.ListPrompts(ctx, nil); err == nil {
		names := make([]string, 0, len(prompts.Prompts))
		for _, p := range prompts.Prompts {
			names = append(names, p.Name)
		}
		fmt.Printf("prompts (%d): %s\n", len(names), strings.Join(names, ", "))
	}
	if resources, err := session.ListResources(ctx, nil); err == nil {
		uris := make([]string, 0, len(resources.Resources))
		for _, r := range resources.Resources {
			uris = append(uris, r.URI)
		}
		fmt.Printf("resources (%d): %s\n", len(uris), strings.Join(uris, ", "))
	}
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
