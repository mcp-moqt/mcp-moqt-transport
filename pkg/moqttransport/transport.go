package mcpmoqt

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mengelbart/moqtransport"
)

// ErrConnectionClosed is returned when sending a message to a connection that
// is closed or in the process of closing.
var ErrConnectionClosed = mcp.ErrConnectionClosed

// Transport re-exports the MCP SDK Transport interface.
type Transport = mcp.Transport

// Connection re-exports the MCP SDK Connection interface.
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
		return nil, errors.New("MOQT session not initialized")
	}

	// This is a placeholder implementation
	// The actual implementation is in server.go and client.go
	return nil, errors.New("not implemented")
}
