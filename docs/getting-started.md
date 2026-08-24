# Getting Started

## Install

```bash
# From a local checkout
bash scripts/install.sh
# or
go install github.com/mcp-moqt/mcp-moqt-transport/cmd/mcp-moqt@latest
```

Windows:

```powershell
powershell -File scripts/install.ps1
```

## Run (stdio — Cursor / Claude Desktop / MCP hosts)

```bash
mcp-moqt server -stdio
```

Sample host configs:

- `configs/mcp/cursor.mcp.json`
- `configs/mcp/claude_desktop.json`
- `configs/mcp/generic.mcp.json`

## Run (QUIC + MOQT)

```bash
mcp-moqt server -addr 127.0.0.1:8080
mcp-moqt client -addr 127.0.0.1:8080
```

Production config file:

```bash
mcp-moqt server -config configs/config.production.yaml
```

Multi-client server:

```bash
mcp-moqt server -addr 127.0.0.1:8080 -multi -metrics 127.0.0.1:9090
```

## Doctor

```bash
mcp-moqt doctor
mcp-moqt doctor -addr 127.0.0.1:8080 -metrics 127.0.0.1:9090
mcp-moqt doctor -tls-cert server.crt -tls-key server.key
```

Validates the Go toolchain, runs an in-memory smoke test for tools / prompts / resources, and optionally probes QUIC, metrics, and TLS certificates.

## TLS

Local defaults use a self-signed server certificate and client `InsecureSkipVerify`.
For production, use `tls.cert_file` / `tls.key_file` in config or `WithTLSServerConfig` / `WithTLSClientConfig`. See [security.md](./security.md) and [deployment.md](./deployment.md).

## Further reading

- [capabilities.md](./capabilities.md) — Tools, Prompts, Resources
- [configuration.md](./configuration.md) — Config files, hot reload, metrics
- [api.md](./api.md) — Public API
