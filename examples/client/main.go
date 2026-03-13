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
		mcpmoqt.RoleClient,
		mcpmoqt.WithAddr(*addr),
	)
	if err != nil {
		log.Fatalf("new client transport: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "example-client",
		Version: "v0.2.0",
	}, nil)

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	if err := session.Ping(ctx, nil); err != nil {
		log.Fatalf("ping: %v", err)
	}

	log.Printf("connected to %s; ping ok", *addr)
}
