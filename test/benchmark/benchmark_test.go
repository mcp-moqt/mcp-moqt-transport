package benchmark

import (
	"testing"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
)

func BenchmarkNewMoqTransport_Server(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transport, err := mcpmoqt.NewMoqTransport(
			mcpmoqt.RoleServer,
			mcpmoqt.WithAddr("127.0.0.1:0"),
		)
		if err != nil {
			b.Fatal(err)
		}
		if transport == nil {
			b.Fatal("transport is nil")
		}
	}
}

func BenchmarkNewMoqTransport_Client(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transport, err := mcpmoqt.NewMoqTransport(
			mcpmoqt.RoleClient,
			mcpmoqt.WithAddr("127.0.0.1:0"),
		)
		if err != nil {
			b.Fatal(err)
		}
		if transport == nil {
			b.Fatal("transport is nil")
		}
	}
}

func BenchmarkWithMultipleOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transport, err := mcpmoqt.NewMoqTransport(
			mcpmoqt.RoleServer,
			mcpmoqt.WithAddr("127.0.0.1:0"),
			mcpmoqt.WithALPN("moq-00"),
		)
		if err != nil {
			b.Fatal(err)
		}
		if transport == nil {
			b.Fatal("transport is nil")
		}
	}
}
