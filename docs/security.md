# Security / TLS

## Local development defaults

| Side | Default |
|------|---------|
| Server | Ephemeral self-signed certificate (TLS 1.2+) |
| Client | `InsecureSkipVerify: true` |

These defaults are for **local demos only**. Never expose them to the public internet.

## Production checklist

### TLS & certificates

- [ ] Use a real server certificate and private key (`tls.cert_file` / `tls.key_file`)
- [ ] Set `MinVersion: TLS 1.2` (enforced by the transport)
- [ ] Set ALPN to `moq-00` (or your negotiated MOQT ALPN)
- [ ] On clients: provide `tls.ca_file` and keep `insecure_skip_verify: false`
- [ ] Monitor certificate expiry (`mcp-moqt doctor -tls-cert ... -tls-key ...`)
- [ ] Rotate certificates before expiry; automate with your PKI / ACME workflow

### Network

- [ ] Bind to private interfaces or place UDP/QUIC behind a firewall
- [ ] Restrict source IPs where possible (QUIC is connection-oriented but still UDP)
- [ ] Do not forward QUIC ports to the internet without TLS and auth at the application layer

### Application

- [ ] Run `mcp-moqt doctor` after deploy
- [ ] Enable metrics on localhost or a protected network (`127.0.0.1:9090`)
- [ ] Keep Go toolchain ≥ 1.25.13 and run `govulncheck` in CI
- [ ] Pin dependency versions in `go.mod`; review upgrades regularly

### MCP Host (stdio)

- [ ] stdio mode inherits the Host's process isolation — only install trusted binaries
- [ ] Use `configs/mcp/*.json` with absolute paths to `mcp-moqt` when possible

### Secrets

- [ ] Do not commit `.crt`, `.key`, or `.env` files
- [ ] Use environment variables or a secrets manager for cert paths in production

## Server TLS (code)

```go
mcpmoqt.NewMoqTransport(
    mcpmoqt.RoleServer,
    mcpmoqt.WithAddr("0.0.0.0:8080"),
    mcpmoqt.WithTLSCertFiles("/etc/mcp-moqt/server.crt", "/etc/mcp-moqt/server.key"),
)
```

## Client TLS (code)

```go
mcpmoqt.NewMoqTransport(
    mcpmoqt.RoleClient,
    mcpmoqt.WithAddr("mcp.example.com:8080"),
    mcpmoqt.WithTLSCAFile("/etc/mcp-moqt/ca.crt"),
)
```

## Config file (YAML)

```yaml
addr: "0.0.0.0:8080"
tls:
  cert_file: "/etc/mcp-moqt/server.crt"
  key_file: "/etc/mcp-moqt/server.key"
```

See `configs/config.production.yaml` for a full example.

## License

MIT — see `LICENSE`.
