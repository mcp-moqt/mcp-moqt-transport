# Configuration

## Load once

```go
cfg, err := config.LoadFromFile("config.yaml")
// or
cfg := config.LoadFromEnv()
```

Environment variables:

| Variable | Meaning |
|----------|---------|
| `MCP_MOQT_ADDR` | Listen / dial address |
| `MCP_MOQT_ALPN` | Comma-separated ALPN list |
| `MCP_MOQT_ENABLE_DATAGRAMS` | `true` / `false` |
| `MCP_MOQT_METRICS_ADDR` | Prometheus `/metrics` listen address |
| `MCP_MOQT_TLS_CERT` | Server TLS certificate file |
| `MCP_MOQT_TLS_KEY` | Server TLS private key file |
| `MCP_MOQT_TLS_CA` | Client CA bundle file |
| `MCP_MOQT_TLS_INSECURE_SKIP_VERIFY` | Client skip verify (`true` = dev only) |

Production YAML example: `configs/config.production.yaml`

```yaml
addr: "0.0.0.0:8080"
metrics_addr: "127.0.0.1:9090"
tls:
  cert_file: "/etc/mcp-moqt/server.crt"
  key_file: "/etc/mcp-moqt/server.key"
```

CLI:

```bash
mcp-moqt server -config configs/config.production.yaml
```

## Hot reload

```go
w, err := config.WatchFile("config.yaml", time.Second, func(cfg *config.Config) {
    // apply runtime-safe fields (logging, feature flags, etc.)
})
if err != nil {
    log.Fatal(err)
}
defer w.Stop()
```

`WatchFile` polls `ModTime` (default interval 1s), validates on reload, and calls the handler with a cloned `*Config`. Invalid files keep the previous good config and may report via `Watcher.OnError`.

> Note: changing listen `addr` at runtime does not automatically rebind the QUIC listener; restart or call a custom rebind if needed.

## Metrics

```go
transport, err := mcpmoqt.NewMoqTransport(
    mcpmoqt.RoleServer,
    mcpmoqt.WithAddr("127.0.0.1:8080"),
    mcpmoqt.WithMetricsAddr("127.0.0.1:9090"),
)
```

Scrape `http://127.0.0.1:9090/metrics` (Prometheus text format). Or mount `mcpmoqt.MetricsHandler(nil)` on your own mux.

## Multi-connection server

```go
raw, _ := mcpmoqt.NewMOQTServerTransport(mcpmoqt.WithAddr("127.0.0.1:8080"))
_ = raw.Serve(ctx, func(ctx context.Context, conn mcpmoqt.Connection) error {
    s := mcp.NewServer(...)
    ss, err := s.Connect(ctx, &mcpmoqt.PreconnectedTransport{Conn: conn}, nil)
    if err != nil {
        return err
    }
    defer ss.Close()
    return ss.Wait()
})
```

CLI: `mcp-moqt server -addr 127.0.0.1:8080 -multi -metrics 127.0.0.1:9090`
