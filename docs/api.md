# MCP over MOQT Transport API Documentation

## Overview

The MCP over MOQT Transport package provides a transport layer implementation for the MCP (Model Context Protocol) SDK, enabling communication over QUIC + MOQT (Media Over QUIC Transport).

## Core Types

### Transport Interface

The `Transport` interface is re-exported from the MCP SDK:

```go
type Transport interface {
    Connect(ctx context.Context) (Connection, error)
}
```

### Connection Interface

The `Connection` interface is re-exported from the MCP SDK:

```go
type Connection interface {
    SessionID() string
    Read(ctx context.Context) (jsonrpc.Message, error)
    Write(ctx context.Context, msg jsonrpc.Message) error
    Close() error
}
```

## Transport Implementations

### MOQTClientTransport

The `MOQTClientTransport` implements the client side of the MCP over MOQT transport.

#### NewMOQTClientTransport

```go
func NewMOQTClientTransport(opts ...Option) (*MOQTClientTransport, error)
```

Creates a new client transport with the specified options. If no TLS config is provided, it uses a default client TLS config with `InsecureSkipVerify` enabled for local development.

#### Connect

```go
func (t *MOQTClientTransport) Connect(ctx context.Context) (Connection, error)
```

Establishes a connection to the server. It:

1. Dials a QUIC connection to the server
2. Creates a MOQT session
3. Discovers the session ID via FETCH
4. Subscribes to the server-to-client control track
5. Returns a control connection

### MOQTServerTransport

The `MOQTServerTransport` implements the server side of the MCP over MOQT transport.

#### NewMOQTServerTransport

```go
func NewMOQTServerTransport(opts ...Option) (*MOQTServerTransport, error)
```

Creates a new server transport with the specified options. If no TLS config is provided, it uses a default self-signed certificate for local development.

#### Connect

```go
func (t *MOQTServerTransport) Connect(ctx context.Context) (Connection, error)
```

Accepts a connection from a client. It:

1. Listens for QUIC connections
2. Creates a MOQT session with discovery and subscribe handlers
3. Runs the MOQT session
4. Subscribes to the client-to-server control track
5. Returns a control connection

## Options

The following options are available for configuring the transport:

### WithAddr

```go
func WithAddr(addr string) Option
```

Sets the address to listen on (server) or connect to (client).

### WithTLSServerConfig

```go
func WithTLSServerConfig(cfg *tls.Config) Option
```

Sets the TLS config for the server.

### WithTLSClientConfig

```go
func WithTLSClientConfig(cfg *tls.Config) Option
```

Sets the TLS config for the client.

### WithQUICConfig

```go
func WithQUICConfig(cfg *quic.Config) Option
```

Sets the QUIC config for both client and server.

### WithALPN

```go
func WithALPN(alpn string) Option
```

Sets the ALPN protocol for TLS. Defaults to "moq-00" for draft-11 compatibility.

## Error Handling

### ErrConnectionClosed

```go
var ErrConnectionClosed = mcp.ErrConnectionClosed
```

Returned when sending a message to a connection that is closed or in the process of closing.

## Example Usage

### Server Example

```go
transport, err := mcpmoqt.NewMOQTServerTransport(
    mcpmoqt.WithAddr("127.0.0.1:8080"),
)
if err != nil {
    log.Fatalf("new transport: %v", err)
}

server := mcp.NewServer(&mcp.Implementation{
    Name:    "example-server",
    Version: "v0.2.0",
}, nil)

if err := server.Run(ctx, transport); err != nil {
    log.Fatalf("server run: %v", err)
}
```

### Client Example

```go
transport, err := mcpmoqt.NewMOQTClientTransport(
    mcpmoqt.WithAddr("127.0.0.1:8080"),
)
if err != nil {
    log.Fatalf("new client transport: %v", err)
}

client := mcp.NewClient(&mcp.Implementation{
    Name:    "example-client",
    Version: "v0.2.0",
}, nil)
	session, err := client.Connect(ctx, transport, nil)
if err != nil {
    log.Fatalf("client connect: %v", err)
}
defer session.Close()

if err := session.Ping(ctx, nil); err != nil {
    log.Fatalf("ping: %v", err)
}
```
