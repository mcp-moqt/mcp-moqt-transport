# Deployment Guide

This guide covers running MCP over MOQT Transport in production-like environments.

## Transport modes

| Mode | Command | Use case |
|------|---------|----------|
| **stdio** | `mcp-moqt server -stdio` | Cursor, Claude Desktop, MCP marketplaces |
| **QUIC/MOQT** | `mcp-moqt server -addr :8080` | Distributed / low-latency MCP between services |

stdio does **not** use QUIC/MOQT; it is the recommended path for MCP Host integration. QUIC/MOQT is the project's core transport for service-to-service communication.

## Single instance (QUIC)

```bash
mcp-moqt server \
  -addr 0.0.0.0:8080 \
  -metrics 127.0.0.1:9090
```

Or with a config file:

```bash
mcp-moqt server -config configs/config.production.yaml
```

## Multi-client server

```bash
mcp-moqt server -addr 0.0.0.0:8080 -multi -metrics 127.0.0.1:9090
```

Each QUIC connection gets its own MCP session and data-track handlers.

## TLS (production)

1. Obtain a certificate (Let's Encrypt, internal CA, or cloud LB termination).
2. Set paths in config:

```yaml
tls:
  cert_file: "/etc/mcp-moqt/server.crt"
  key_file: "/etc/mcp-moqt/server.key"
```

3. On clients, provide `tls.ca_file` and **do not** set `insecure_skip_verify: true`.

Environment variables:

- `MCP_MOQT_TLS_CERT`
- `MCP_MOQT_TLS_KEY`
- `MCP_MOQT_TLS_CA`
- `MCP_MOQT_TLS_INSECURE_SKIP_VERIFY` (dev only)

See [security.md](./security.md) for the full checklist.

## Firewall / networking

- QUIC uses **UDP** (not TCP). Open the server port on UDP.
- Prefer binding to a private interface (`10.x`, `172.16–31.x`) and expose via VPN or internal LB.
- NAT / UDP hole-punching may be required for cross-network clients.

## Prometheus metrics

Enable with `-metrics` or `metrics_addr` in config. Endpoints:

- `GET /metrics` — Prometheus text format
- `GET /healthz` — liveness probe

Example scrape config:

```yaml
scrape_configs:
  - job_name: mcp-moqt
    static_configs:
      - targets: ["127.0.0.1:9090"]
```

Import the sample dashboard from `configs/grafana/dashboard.json`.

## Docker

```bash
docker compose up --build
```

Mount TLS certificates and set `MCP_MOQT_TLS_CERT` / `MCP_MOQT_TLS_KEY` in compose overrides for production.

## Health checks

```bash
mcp-moqt doctor
mcp-moqt doctor -addr 127.0.0.1:8080 -metrics 127.0.0.1:9090
mcp-moqt doctor -tls-cert /etc/mcp-moqt/server.crt -tls-key /etc/mcp-moqt/server.key
```

## Multi-instance notes

- Each instance needs a unique UDP listen address (or separate hosts).
- Sticky routing is not required; each QUIC connection is independent.
- Use a load balancer that supports QUIC/UDP if exposing multiple backends (e.g. QUIC-aware proxy or DNS round-robin per host).

## Release

Tag and publish:

```bash
git tag v1.5.0
git push origin v1.5.0
```

GitHub Actions `release.yml` builds multi-platform binaries and `SHA256SUMS.txt`.
