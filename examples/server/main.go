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
	addr := flag.String("addr", "127.0.0.1:8080", "QUIC server address (ignored with -stdio)")
	stdio := flag.Bool("stdio", false, "serve over stdin/stdout")
	flag.Parse()

	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    mcpservice.ServiceName,
		Version: mcpservice.ServiceVersion,
	}, nil)
	mcpservice.RegisterCapabilities(server)

	if *stdio {
		log.Printf("MCP over MOQT demo server (stdio); capabilities: tools, prompts, resources")
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			log.Fatalf("server run: %v", err)
		}
		return
	}

	transport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleServer,
		mcpmoqt.WithAddr(*addr),
	)
	if err != nil {
		log.Fatalf("new transport: %v", err)
	}

	log.Printf("MCP over MOQT server listening on %s", *addr)
	log.Printf("capabilities: tools, prompts, resources")
	log.Printf("tip: use -stdio for Cursor/MCP hosts; see configs/mcp/")

	if err := server.Run(ctx, transport); err != nil {
		log.Fatalf("server run: %v", err)
	}
}
