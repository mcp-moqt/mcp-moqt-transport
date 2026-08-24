# MCP Capabilities

This service registers practical MCP **Tools**, **Prompts**, and **Resources** via `pkg/mcpservice`.

## Tools

| Name | Description |
|------|-------------|
| `health_check` | Runtime / health snapshot |
| `get_transport_info` | ALPN, draft, feature list |
| `list_data_tracks` | Available MOQT data tracks |
| `estimate_rtt` | Simple RTT estimate from samples |

## Prompts

| Name | Arguments |
|------|-----------|
| `debug_moqt_session` | `symptom` (required), `addr` (optional) |
| `configure_transport` | `environment` (required), `goal` (optional) |

## Resources

| URI | MIME |
|-----|------|
| `moqt://docs/overview` | text/markdown |
| `moqt://config/example` | application/yaml |
| `moqt://tracks/catalog` | application/json |
| `moqt://install/quickstart` | text/markdown |

## Registration

```go
server := mcp.NewServer(&mcp.Implementation{
    Name:    mcpservice.ServiceName,
    Version: mcpservice.ServiceVersion,
}, nil)
mcpservice.RegisterCapabilities(server)
```

## Transports

| Mode | Command | Use case |
|------|---------|----------|
| stdio | `mcp-moqt server -stdio` | Cursor / Claude Desktop / marketplace hosts |
| QUIC/MOQT | `mcp-moqt server -addr HOST:PORT` | High-performance MCP over MOQT |
