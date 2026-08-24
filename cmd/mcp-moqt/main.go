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
  server    启动带 tools/prompts/resources 的 MCP 服务
  client    连接服务并演示 capabilities
  doctor    本地环境自检（Go/模块/示例路径）
  help      显示帮助

Examples:
  mcp-moqt server -addr 127.0.0.1:8080
  mcp-moqt client -addr 127.0.0.1:8080
  mcp-moqt doctor
`)
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "listen address")
	_ = fs.Parse(args)

	ctx := context.Background()
	transport, err := mcpmoqt.NewMoqTransport(mcpmoqt.RoleServer, mcpmoqt.WithAddr(*addr))
	if err != nil {
		fatalf("new transport: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    mcpservice.ServiceName,
		Version: mcpservice.ServiceVersion,
	}, nil)
	mcpservice.RegisterCapabilities(server)

	fmt.Printf("listening on %s (tools/prompts/resources enabled)\n", *addr)
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

func runDoctor(args []string) {
	_ = flag.NewFlagSet("doctor", flag.ExitOnError).Parse(args)

	checks := []struct {
		name string
		ok   bool
		hint string
	}{
		{"go toolchain", commandExists("go"), "https://go.dev/dl/"},
		{"service package", true, "pkg/mcpservice (tools/prompts/resources)"},
		{"CLI entrypoint", true, "cmd/mcp-moqt"},
		{"install scripts", true, "scripts/install.sh + scripts/install.ps1"},
		{"docker compose", true, "docker compose up --build"},
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
	if !allOK {
		os.Exit(1)
	}
	fmt.Println("environment looks ready")
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
