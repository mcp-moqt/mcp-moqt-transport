# Security / TLS

## Local development defaults

| Side | Default |
|------|---------|
| Server | Ephemeral self-signed certificate (TLS 1.2+) |
| Client | `InsecureSkipVerify: true` |

These defaults are for **local demos only**.

## Production checklist

1. Provide a real certificate and key to the server:

```go
tlsCert, err := tls.LoadX509KeyPair("server.crt", "server.key")
// ...
mcpmoqt.WithTLSServerConfig(&tls.Config{
    Certificates: []tls.Certificate{tlsCert},
    MinVersion:   tls.VersionTLS12,
    NextProtos:   []string{"moq-00"},
})
```

2. On the client, verify the server certificate (do **not** skip verify):

```go
mcpmoqt.WithTLSClientConfig(&tls.Config{
    RootCAs:    pool, // system or custom CA pool
    MinVersion: tls.VersionTLS12,
    NextProtos: []string{"moq-00"},
    // InsecureSkipVerify: false (default)
})
```

3. Prefer binding to a private interface or placing a firewall in front of UDP/QUIC.
4. Keep dependencies updated (`go get -u` / CI `govulncheck`).

## License

MIT — see `LICENSE`.
