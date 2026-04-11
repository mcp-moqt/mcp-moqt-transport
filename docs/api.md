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
    Version: "v0.6.0",
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
    Version: "v0.6.0",
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

## Reliability Features

Reliability features (ack/heartbeat/retry/metrics) are now implemented in v0.6.0.

### Acknowledgment Tracker

The `AckTracker` provides message acknowledgment tracking with timeout support.

```go
config := mcpmoqt.DefaultAckConfig()
config.Timeout = 5 * time.Second

tracker := mcpmoqt.NewAckTracker(config)
defer tracker.Close()

ackCh := tracker.Track("message-id")

go func() {
    time.Sleep(100 * time.Millisecond)
    tracker.Ack("message-id")
}()

ack := <-ackCh
fmt.Printf("Status: %v\n", ack.Status)
```

#### AckStatus

```go
type AckStatus int

const (
    AckStatusPending  AckStatus = iota
    AckStatusAcked
    AckStatusNacked
    AckStatusTimeout
)
```

### Heartbeat

The `Heartbeat` provides connection health monitoring.

```go
config := mcpmoqt.DefaultHeartbeatConfig()
config.Interval = 30 * time.Second
config.OnHeartbeat = func() error {
    return sendPing()
}

hb := mcpmoqt.NewHeartbeat(config)
hb.Start(ctx)
defer hb.Stop()

fmt.Printf("Healthy: %v\n", hb.IsHealthy())
```

### Retry Mechanism

The `Retry` provides exponential backoff retry logic.

```go
config := mcpmoqt.DefaultRetryConfig()
config.MaxAttempts = 3
config.Multiplier = 2.0

retry := mcpmoqt.NewRetry(config)

err := retry.Do(ctx, func() error {
    return sendMessage()
})
```

### Metrics

The `Metrics` provides performance monitoring and statistics.

```go
metrics := mcpmoqt.NewMetrics()

metrics.RecordMessageSent(100)
metrics.RecordAck()
metrics.RecordConnectionOpened()

snapshot := metrics.Snapshot()
fmt.Printf("Messages: %d\n", snapshot.MessagesSent)
fmt.Printf("Ack Rate: %.2f%%\n", snapshot.AckRate*100)
fmt.Printf("Throughput: %.2f bytes/sec\n", snapshot.ThroughputOut)
```

#### MetricsSnapshot

```go
type MetricsSnapshot struct {
    MessagesSent      int64
    MessagesReceived  int64
    MessagesAcked     int64
    MessagesNacked    int64
    BytesSent         int64
    BytesReceived     int64
    Errors            int64
    Retries           int64
    ConnectionsActive int64
    TracksActive      int64
    Uptime            time.Duration
    AckRate           float64
    ErrorRate         float64
    ThroughputIn      float64
    ThroughputOut     float64
}
```

## Data Tracks

### DataTrackHandler

The `DataTrackHandler` manages data tracks for resources, tools, and notifications.

```go
handler := mcpmoqt.NewDataTrackHandler("session-id", mcpmoqt.DataTrackResources)

handler.Subscribe(ctx, "resource-name", func(msg *mcpmoqt.DataTrackMessage) error {
    fmt.Printf("Received: %+v\n", msg)
    return nil
})

handler.Publish(ctx, "resource-name", &mcpmoqt.DataTrackMessage{
    TrackType: mcpmoqt.DataTrackResources,
    TrackName: "resource-name",
    Data:      map[string]interface{}{"key": "value"},
})

defer handler.Close()
```

### DataTrackType

```go
type DataTrackType string

const (
    DataTrackResources     DataTrackType = "resources"
    DataTrackTools         DataTrackType = "tools"
    DataTrackNotifications DataTrackType = "notifications"
)
```
