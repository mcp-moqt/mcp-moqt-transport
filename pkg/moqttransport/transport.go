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

// MOQTTransport is a legacy struct for advanced MOQT session wiring.
// Prefer NewMoqTransport(RoleServer|RoleClient, opts...) for MCP SDK integration.
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

// Connect is not implemented on MOQTTransport; use NewMoqTransport instead.
func (t *MOQTTransport) Connect(ctx context.Context) (Connection, error) {
	_ = ctx
	return nil, NewTransportError("Connect", ErrSessionNotFound)
}
