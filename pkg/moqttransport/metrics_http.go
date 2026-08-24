package mcpmoqt

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// MetricsHTTPServer exposes Prometheus text-format metrics at /metrics.
type MetricsHTTPServer struct {
	addr    string
	metrics *Metrics
	server  *http.Server
}

// NewMetricsHTTPServer creates an HTTP server that scrapes m (defaults to DefaultMetrics).
func NewMetricsHTTPServer(addr string, m *Metrics) *MetricsHTTPServer {
	if m == nil {
		m = DefaultMetrics()
	}
	mux := http.NewServeMux()
	s := &MetricsHTTPServer{addr: addr, metrics: m}
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	s.server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// Start listens and serves until Close. Blocks.
func (s *MetricsHTTPServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.addr = ln.Addr().String()
	s.server.Addr = s.addr
	err = s.server.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Addr returns the bound address (available after Start begins listening).
func (s *MetricsHTTPServer) Addr() string {
	return s.addr
}

// Close shuts down the HTTP server.
func (s *MetricsHTTPServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *MetricsHTTPServer) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	snap := s.metrics.Snapshot()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	var b strings.Builder
	writePromCounter(&b, "mcp_moqt_messages_sent_total", "Total MCP messages sent", float64(snap.MessagesSent))
	writePromCounter(&b, "mcp_moqt_messages_received_total", "Total MCP messages received", float64(snap.MessagesReceived))
	writePromCounter(&b, "mcp_moqt_messages_acked_total", "Total messages acknowledged", float64(snap.MessagesAcked))
	writePromCounter(&b, "mcp_moqt_messages_nacked_total", "Total messages negatively acknowledged", float64(snap.MessagesNacked))
	writePromCounter(&b, "mcp_moqt_messages_timeout_total", "Total message ACK timeouts", float64(snap.MessagesTimeout))
	writePromCounter(&b, "mcp_moqt_bytes_sent_total", "Total bytes sent", float64(snap.BytesSent))
	writePromCounter(&b, "mcp_moqt_bytes_received_total", "Total bytes received", float64(snap.BytesReceived))
	writePromCounter(&b, "mcp_moqt_errors_total", "Total errors", float64(snap.Errors))
	writePromCounter(&b, "mcp_moqt_retries_total", "Total write retries", float64(snap.Retries))
	writePromCounter(&b, "mcp_moqt_heartbeats_sent_total", "Total heartbeats sent", float64(snap.HeartbeatsSent))
	writePromCounter(&b, "mcp_moqt_heartbeats_failed_total", "Total heartbeat failures", float64(snap.HeartbeatsFailed))
	writePromCounter(&b, "mcp_moqt_connections_total", "Total connections opened", float64(snap.ConnectionsTotal))
	writePromGauge(&b, "mcp_moqt_connections_active", "Currently active connections", float64(snap.ConnectionsActive))
	writePromGauge(&b, "mcp_moqt_tracks_active", "Currently active tracks", float64(snap.TracksActive))
	writePromCounter(&b, "mcp_moqt_tracks_total", "Total tracks opened", float64(snap.TracksTotal))
	writePromCounter(&b, "mcp_moqt_data_track_messages_total", "Total data track messages", float64(snap.DataTrackMessages))
	writePromCounter(&b, "mcp_moqt_data_track_bytes_total", "Total data track payload bytes", float64(snap.DataTrackBytes))
	writePromGauge(&b, "mcp_moqt_uptime_seconds", "Process metrics uptime in seconds", snap.Uptime.Seconds())
	writePromGauge(&b, "mcp_moqt_ack_rate", "ACK rate (acked/sent)", snap.AckRate)
	writePromGauge(&b, "mcp_moqt_error_rate", "Error rate", snap.ErrorRate)
	writePromGauge(&b, "mcp_moqt_throughput_in_bytes_per_sec", "Inbound throughput (bytes/s)", snap.ThroughputIn)
	writePromGauge(&b, "mcp_moqt_throughput_out_bytes_per_sec", "Outbound throughput (bytes/s)", snap.ThroughputOut)
	_, _ = w.Write([]byte(b.String()))
}

func writePromCounter(b *strings.Builder, name, help string, v float64) {
	fmt.Fprintf(b, "# HELP %s %s\n# TYPE %s counter\n%s %g\n", name, help, name, name, v)
}

func writePromGauge(b *strings.Builder, name, help string, v float64) {
	fmt.Fprintf(b, "# HELP %s %s\n# TYPE %s gauge\n%s %g\n", name, help, name, name, v)
}

// Handler returns an http.Handler for /metrics using DefaultMetrics (for custom muxes).
func MetricsHandler(m *Metrics) http.Handler {
	if m == nil {
		m = DefaultMetrics()
	}
	s := &MetricsHTTPServer{metrics: m}
	return http.HandlerFunc(s.handleMetrics)
}
