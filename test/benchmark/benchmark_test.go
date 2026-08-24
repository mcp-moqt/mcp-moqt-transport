package benchmark

import (
	"net/http"
	"net/http/httptest"
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

func BenchmarkMetricsSnapshot(b *testing.B) {
	m := mcpmoqt.NewMetrics()
	m.RecordMessageSent(512)
	m.RecordDataTrackMessage(256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Snapshot()
	}
}

func BenchmarkMetricsPrometheusRender(b *testing.B) {
	m := mcpmoqt.NewMetrics()
	m.RecordMessageSent(1024)
	m.RecordMessageReceived(512)
	h := mcpmoqt.MetricsHandler(m)
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
}
