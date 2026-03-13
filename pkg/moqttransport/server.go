package mcpmoqt

import (
	"context"
	"fmt"

	"github.com/mengelbart/moqtransport"
	"github.com/mengelbart/moqtransport/quicmoq"
)

// MOQTServerTransport implements the server side of MCP over MOQT transport.
type MOQTServerTransport struct {
	cfg       *transportConfig
	sessionID string
}

// NewMOQTServerTransport creates a new server transport.
func NewMOQTServerTransport(opts ...Option) (*MOQTServerTransport, error) {
	cfg, err := applyOptions(roleServer, opts)
	if err != nil {
		return nil, err
	}

	// Set default TLS config if not provided
	if cfg.tlsServer == nil {
		tlsCfg, err := defaultTLSConfig(roleServer, cfg.alpn)
		if err != nil {
			return nil, err
		}
		cfg.tlsServer = tlsCfg
	}

	return &MOQTServerTransport{cfg: cfg, sessionID: generateSessionID()}, nil
}

// Draft: draft-jennings-mcp-over-moqt-00 §2
// MCP messages are mapped onto MOQT objects via control tracks.
// NOTE: This implementation targets moqtransport (draft-11, moq-00).
// It is not wire-compatible with draft-16 stream/datagram encodings.
// Connect implements the Transport interface for server.
func (t *MOQTServerTransport) Connect(ctx context.Context) (Connection, error) {
	// Listen for QUIC connections
	quicConn, ln, _, err := listenAndAcceptMOQT(ctx, t.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to listen and accept QUIC connection: %w", err)
	}

	// Wrap QUIC connection with quicmoq to implement moqtransport.Connection
	conn := quicmoq.NewServer(quicConn)

	// Create publisher slot for server-to-client track
	sendSlot := newPublisherSlot()

	// Create MOQT session with discovery and subscribe handlers
	session := &moqtransport.Session{
		Handler: &discoveryHandler{
			sessionID: t.sessionID,
		},
		SubscribeHandler: &subscribeHandler{
			sessionID: t.sessionID,
			sendTrack: trackServerToClient,
			sendSlot:  sendSlot,
		},
	}

	// Run MOQT session
	if err := runSession(ctx, session, conn); err != nil {
		ln.Close()
		return nil, fmt.Errorf("failed to run MOQT session: %w", err)
	}

	// Subscribe to client-to-server control track
	recv, err := session.Subscribe(ctx, controlNamespace(t.sessionID), trackClientToServer)
	if err != nil {
		ln.Close()
		return nil, fmt.Errorf("failed to subscribe to client-to-server control track: %w", err)
	}

	// Create and return control connection
	return newControlConn(conn, t.sessionID, recv, sendSlot), nil
}
