# MCP over MOQT Transport Design Document

## 1. Overview

The MCP over MOQT Transport project provides a transport layer implementation for the MCP (Model Context Protocol) SDK, enabling communication over QUIC + MOQT (Media Over QUIC Transport). This design document outlines the architecture, components, and implementation details of the project.

## 2. Architecture

### 2.1 High-Level Architecture

The project follows a layered architecture:

1. **Transport Layer**: Implements the MCP SDK's `Transport` and `Connection` interfaces
2. **MOQT Layer**: Manages MOQT sessions and tracks
3. **QUIC Layer**: Provides the underlying QUIC connection
4. **TLS Layer**: Ensures secure communication

### 2.2 Component Diagram

```
+-------------------------+
| MCP SDK                 |
+-------------------------+
            |
            v
+-------------------------+
| Transport Interface     |
+-------------------------+
            |
            v
+-------------------------+
| MOQTClientTransport     |
| MOQTServerTransport     |
+-------------------------+
            |
            v
+-------------------------+
| MOQT Session            |
+-------------------------+
            |
            v
+-------------------------+
| QUIC Connection         |
+-------------------------+
            |
            v
+-------------------------+
| TLS                     |
+-------------------------+
```

## 3. Core Components

### 3.1 Transport Implementations

#### 3.1.1 MOQTClientTransport

The `MOQTClientTransport` handles client-side communication. It:
- Dials QUIC connections to servers
- Creates MOQT sessions
- Discovers session IDs via FETCH
- Subscribes to server-to-client control tracks
- Returns control connections for MCP message exchange

#### 3.1.2 MOQTServerTransport

The `MOQTServerTransport` handles server-side communication. It:
- Listens for QUIC connections
- Creates MOQT sessions with discovery and subscribe handlers
- Runs MOQT sessions
- Subscribes to client-to-server control tracks
- Returns control connections for MCP message exchange

### 3.2 Control Connection

The `controlConn` implements the `Connection` interface and handles MCP message exchange over MOQT control tracks. It:
- Reads JSON-RPC messages from the receive track
- Writes JSON-RPC messages to the send track
- Manages connection lifecycle

### 3.3 Session Handlers

#### 3.3.1 discoveryHandler

The `discoveryHandler` handles server-side discovery requests via FETCH. It:
- Responds to FETCH requests for "mcp/discovery/sessions"
- Returns session ID and available control tracks

#### 3.3.2 subscribeHandler

The `subscribeHandler` handles subscribe requests for control tracks. It:
- Accepts subscriptions to the local send track
- Exposes a Publisher via a slot for writing messages

### 3.4 Publisher Slot

The `publisherSlot` manages the publisher for the send track. It:
- Stores the publisher
- Provides synchronization for publisher availability
- Allows waiting for the publisher to be set

## 4. Protocol Mapping

### 4.1 MCP to MOQT Mapping

MCP JSON-RPC messages are mapped to MOQT objects on control tracks:

- **Control Tracks**: Namespace "mcp/<session-id>/control"
  - "client-to-server": Client to server messages
  - "server-to-client": Server to client messages

### 4.2 Discovery

Discovery is implemented via FETCH on "mcp/discovery/sessions":
- Client sends a FETCH request
- Server responds with session ID and available control tracks
- Client uses the session ID to subscribe to control tracks

## 5. Configuration

### 5.1 Transport Options

The transport can be configured with the following options:
- Address: The address to listen on (server) or connect to (client)
- TLS Config: Custom TLS configuration for server and client
- QUIC Config: Custom QUIC configuration
- ALPN: ALPN protocol for TLS (defaults to "moq-00")

### 5.2 Default Behavior

- **Server**: Uses a self-signed TLS certificate for local development
- **Client**: Uses `InsecureSkipVerify` for local development
- **ALPN**: Defaults to "moq-00" for draft-11 compatibility

## 6. Error Handling

### 6.1 Error Types

- **ErrConnectionClosed**: Returned when sending a message to a closed connection
- **Context errors**: Returned when operations are cancelled or timed out
- **QUIC/MOQT errors**: Returned for transport-level errors

### 6.2 Error Propagation

Errors are propagated up the stack to the MCP SDK, which handles them according to its error handling strategy.

## 7. Security Considerations

### 7.1 TLS

- The default TLS configuration is intended for local development only
- In production, users should provide their own TLS certificates
- Client-side `InsecureSkipVerify` should be disabled in production

### 7.2 Session Management

- Session IDs are generated randomly to prevent predictability
- Control tracks are scoped to specific sessions to prevent cross-session interference

## 8. Performance Considerations

### 8.1 QUIC Benefits

- **Multiplexing**: Multiple streams over a single connection
- **0-RTT**: Faster connection establishment
- **Connection migration**: Seamless handoff between networks

### 8.2 MOQT Benefits

- **Object-based delivery**: Efficient delivery of discrete messages
- **Prioritization**: Support for prioritizing control messages
- **Reliability**: Built-in reliability for control messages

## 9. Future Enhancements

### 9.1 Resources/Tools/Notifications Tracks

- Reserve tracks for resources/tools/notifications as specified in the draft
- Implement subscription/publish semantics for these tracks

### 9.2 Enhanced Error Handling

- More comprehensive error model
- Better error recovery strategies

### 9.3 Performance Optimizations

- Connection pooling
- Message batching
- Improved memory management

## 10. Conclusion

The MCP over MOQT Transport implementation provides a robust, secure, and performant transport layer for MCP SDK. It leverages the benefits of QUIC and MOQT to deliver a reliable communication channel for MCP messages. The design is modular, extensible, and aligned with the draft specification, providing a solid foundation for future enhancements.