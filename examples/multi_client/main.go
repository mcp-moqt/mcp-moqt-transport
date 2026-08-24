// Multi-client example: one QUIC/MOQT server with concurrent MCP clients.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mcp-moqt/mcp-moqt-transport/pkg/mcpservice"
	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "server listen/connect address")
	clients := flag.Int("clients", 5, "number of concurrent clients")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	raw, err := mcpmoqt.NewMOQTServerTransport(mcpmoqt.WithAddr(*addr))
	if err != nil {
		log.Fatalf("server transport: %v", err)
	}
	defer raw.Close()

	go func() {
		_ = raw.Serve(ctx, func(ctx context.Context, conn mcpmoqt.Connection) error {
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
		})
	}()

	time.Sleep(1500 * time.Millisecond)

	var wg sync.WaitGroup
	errCh := make(chan error, *clients)
	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cctx, ccancel := context.WithTimeout(ctx, 20*time.Second)
			defer ccancel()

			transport, err := mcpmoqt.NewMoqTransport(mcpmoqt.RoleClient, mcpmoqt.WithAddr(*addr))
			if err != nil {
				errCh <- fmt.Errorf("client %d transport: %w", id, err)
				return
			}
			client := mcp.NewClient(&mcp.Implementation{
				Name:    fmt.Sprintf("multi-client-%d", id),
				Version: mcpservice.ServiceVersion,
			}, nil)
			session, err := client.Connect(cctx, transport, nil)
			if err != nil {
				errCh <- fmt.Errorf("client %d connect: %w", id, err)
				return
			}
			defer session.Close()
			if err := session.Ping(cctx, nil); err != nil {
				errCh <- fmt.Errorf("client %d ping: %w", id, err)
				return
			}
			log.Printf("client %d: ping ok", id)
		}(i + 1)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		log.Printf("error: %v", err)
	}
	fmt.Printf("multi_client done: %d clients against %s\n", *clients, *addr)
}
