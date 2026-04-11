package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transport, err := mcpmoqt.NewMoqTransport(
		mcpmoqt.RoleServer,
		mcpmoqt.WithAddr("127.0.0.1:8081"),
	)
	if err != nil {
		log.Fatalf("new transport: %v", err)
	}

	resourceHandler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)
	toolHandler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackTools)
	notificationHandler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackNotifications)

	err = resourceHandler.Subscribe("file-resource", func(msg *mcpmoqt.DataTrackMessage) error {
		data, _ := json.MarshalIndent(msg, "", "  ")
		fmt.Printf("Received resource update: %s\n", data)
		return nil
	})
	if err != nil {
		log.Printf("Failed to subscribe to resource: %v", err)
	}

	err = toolHandler.Subscribe("calculator-tool", func(msg *mcpmoqt.DataTrackMessage) error {
		data, _ := json.MarshalIndent(msg, "", "  ")
		fmt.Printf("Received tool invocation: %s\n", data)
		return nil
	})
	if err != nil {
		log.Printf("Failed to subscribe to tool: %v", err)
	}

	err = notificationHandler.Subscribe("progress-notification", func(msg *mcpmoqt.DataTrackMessage) error {
		data, _ := json.MarshalIndent(msg, "", "  ")
		fmt.Printf("Received notification: %s\n", data)
		return nil
	})
	if err != nil {
		log.Printf("Failed to subscribe to notification: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "data-track-example-server",
		Version: "v0.4.0",
	}, nil)

	resourceMsg := &mcpmoqt.DataTrackMessage{
		TrackType: mcpmoqt.DataTrackResources,
		TrackName: "file-resource",
		Data: map[string]interface{}{
			"uri":      "file:///example.txt",
			"content":  "Hello, World!",
			"mimeType": "text/plain",
		},
		Metadata: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"size":      13,
		},
	}

	fmt.Println("Data track example initialized successfully")
	fmt.Println("Data track handlers created:")
	fmt.Printf("  - Resources: %v\n", mcpmoqt.DataTrackResources)
	fmt.Printf("  - Tools: %v\n", mcpmoqt.DataTrackTools)
	fmt.Printf("  - Notifications: %v\n", mcpmoqt.DataTrackNotifications)

	data, _ := json.MarshalIndent(resourceMsg, "", "  ")
	fmt.Printf("\nExample resource message:\n%s\n", data)

	fmt.Println("\nStarting server...")

	if err := server.Run(ctx, transport); err != nil {
		log.Printf("server error: %v", err)
	}
}
