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
		Version: "1.3.0",
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

	// Exercise tools / prompts / resources when the demo server exposes them.
	if tools, err := session.ListTools(ctx, nil); err == nil {
		log.Printf("tools: %d", len(tools.Tools))
		if _, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "health_check"}); err != nil {
			log.Printf("health_check: %v", err)
		} else {
			log.Printf("health_check: ok")
		}
	}

	if prompts, err := session.ListPrompts(ctx, nil); err == nil {
		log.Printf("prompts: %d", len(prompts.Prompts))
	}

	if resources, err := session.ListResources(ctx, nil); err == nil {
		log.Printf("resources: %d", len(resources.Resources))
		if res, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "moqt://docs/overview"}); err == nil && len(res.Contents) > 0 {
			log.Printf("read resource moqt://docs/overview (%d bytes)", len(res.Contents[0].Text))
		}
	}
}
