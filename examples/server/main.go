package main

import (
	"context"
	"flag"
	"log"

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
		Name:    "example-server",
		Version: "v0.2.0",
	}, nil)

	if err := server.Run(ctx, transport); err != nil {
		log.Fatalf("server run: %v", err)
	}
}
