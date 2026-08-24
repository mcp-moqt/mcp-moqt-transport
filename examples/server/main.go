package main

import (
	"context"
	"flag"
	"log"

	"github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"
	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "server address")
	flag.Parse()

	ctx := context.Background()

	transport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleServer,
		mcpmoqt.WithAddr(*addr),
	)
	if err != nil {
		log.Fatalf("new transport: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    mcpservice.ServiceName,
		Version: mcpservice.ServiceVersion,
	}, nil)

	// Register tools / prompts / resources for marketplace completeness.
	mcpservice.RegisterCapabilities(server)

	log.Printf("MCP over MOQT server listening on %s", *addr)
	log.Printf("capabilities: tools, prompts, resources")

	if err := server.Run(ctx, transport); err != nil {
		log.Fatalf("server run: %v", err)
	}
}
