package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== MCP-MOQT Transport Features Demo ===")
	fmt.Println()

	demoAckTracker(ctx)
	fmt.Println()

	demoHeartbeat(ctx)
	fmt.Println()

	demoRetry(ctx)
	fmt.Println()

	demoMetrics(ctx)
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}

func demoAckTracker(ctx context.Context) {
	fmt.Println("--- Acknowledgment Tracker Demo ---")

	config := mcpmoqt.DefaultAckConfig()
	config.Timeout = 2 * time.Second

	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	ackCh := tracker.Track("msg-001")
	fmt.Printf("Tracking message: msg-001\n")

	go func() {
		time.Sleep(500 * time.Millisecond)
		tracker.Ack("msg-001")
	}()

	select {
	case ack := <-ackCh:
		fmt.Printf("Received ACK: ID=%s, Status=%v\n", ack.ID, ack.Status)
	case <-ctx.Done():
		fmt.Println("Context cancelled")
	}

	fmt.Printf("Pending messages: %d\n", tracker.PendingCount())
}

func demoHeartbeat(ctx context.Context) {
	fmt.Println("--- Heartbeat Demo ---")

	heartbeatCount := 0
	config := mcpmoqt.DefaultHeartbeatConfig()
	config.Interval = 1 * time.Second
	config.Timeout = 500 * time.Millisecond
	config.OnHeartbeat = func() error {
		heartbeatCount++
		fmt.Printf("Heartbeat #%d sent\n", heartbeatCount)
		return nil
	}
	config.OnTimeout = func() {
		fmt.Println("Heartbeat timeout!")
	}

	hb := mcpmoqt.NewHeartbeat(config)

	if err := hb.Start(ctx); err != nil {
		log.Printf("Failed to start heartbeat: %v", err)
		return
	}
	defer hb.Stop()

	fmt.Printf("Heartbeat started, interval: %v\n", config.Interval)

	time.Sleep(3 * time.Second)

	fmt.Printf("Heartbeat running: %v\n", hb.IsRunning())
	fmt.Printf("Heartbeat healthy: %v\n", hb.IsHealthy())
	fmt.Printf("Total heartbeats: %d\n", heartbeatCount)
}

func demoRetry(ctx context.Context) {
	fmt.Println("--- Retry Mechanism Demo ---")

	config := mcpmoqt.DefaultRetryConfig()
	config.MaxAttempts = 3
	config.InitialDelay = 100 * time.Millisecond

	retry := mcpmoqt.NewRetry(config)

	attempts := 0
	err := retry.Do(ctx, func() error {
		attempts++
		fmt.Printf("Attempt %d\n", attempts)
		if attempts < 3 {
			return fmt.Errorf("temporary error")
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Retry failed after %d attempts: %v\n", attempts, err)
	} else {
		fmt.Printf("Retry succeeded after %d attempts\n", attempts)
	}
}

func demoMetrics(ctx context.Context) {
	fmt.Println("--- Metrics Demo ---")

	metrics := mcpmoqt.NewMetrics()

	metrics.RecordMessageSent(100)
	metrics.RecordMessageSent(200)
	metrics.RecordMessageReceived(150)
	metrics.RecordAck()
	metrics.RecordAck()
	metrics.RecordNack()
	metrics.RecordConnectionOpened()
	metrics.RecordTrackOpened()
	metrics.RecordHeartbeatSent()

	snapshot := metrics.Snapshot()

	data, _ := json.MarshalIndent(snapshot, "", "  ")
	fmt.Printf("Metrics Snapshot:\n%s\n", string(data))

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Messages Sent: %d\n", metrics.MessagesSent())
	fmt.Printf("  Messages Received: %d\n", metrics.MessagesReceived())
	fmt.Printf("  Ack Rate: %.2f%%\n", snapshot.AckRate*100)
	fmt.Printf("  Uptime: %v\n", metrics.Uptime())
}
