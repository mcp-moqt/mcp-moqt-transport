package mcpmoqt

import (
	"context"

	"github.com/mengelbart/moqtransport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	DefaultMaxRequestID uint64 = 100
)

type Transport = mcp.Transport

type Connection = mcp.Connection

// MOQTTransport is a transport that communicates over MOQT.
type MOQTTransport struct {
	// Session is the underlying MOQT session
	Session *moqtransport.Session

	// SessionID is the MCP session identifier
	SessionID string

	// Namespace is the MOQT namespace for MCP tracks
	Namespace []string

	// ControlTrackNamespace is the namespace for control tracks
	ControlTrackNamespace []string
}

// Connect implements the Transport interface.
func (t *MOQTTransport) Connect(ctx context.Context) (Connection, error) {
	if t.Session == nil {
		return nil, NewTransportError("Connect", ErrSessionNotFound)
	}

	// This is a placeholder implementation
	// The actual implementation is in server.go and client.go
	return nil, NewTransportError("Connect", ErrSessionNotFound)
}
