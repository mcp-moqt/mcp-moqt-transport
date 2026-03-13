package mcpmoqt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mengelbart/moqtransport"
	"github.com/mengelbart/moqtransport/quicmoq"
)

// MOQTClientTransport implements the client side of MCP over MOQT transport.
type MOQTClientTransport struct {
	cfg *transportConfig
}

// NewMOQTClientTransport creates a new client transport.
func NewMOQTClientTransport(opts ...Option) (*MOQTClientTransport, error) {
	cfg, err := applyOptions(roleClient, opts)
	if err != nil {
		return nil, err
	}

	// Set default TLS config if not provided
	if cfg.tlsClient == nil {
		tlsCfg, err := defaultTLSConfig(roleClient, cfg.alpn)
		if err != nil {
			return nil, err
		}
		cfg.tlsClient = tlsCfg
	}

	return &MOQTClientTransport{cfg: cfg}, nil
}

// Draft: draft-jennings-mcp-over-moqt-00 §2
// MCP messages are mapped onto MOQT objects via control tracks.
// NOTE: This implementation targets moqtransport (draft-11, moq-00).
// It is not wire-compatible with draft-16 stream/datagram encodings.
// Connect implements the Transport interface for client.
func (t *MOQTClientTransport) Connect(ctx context.Context) (Connection, error) {
	// Dial QUIC connection to server
	quicConn, _, err := dialMOQT(ctx, t.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to dial QUIC connection: %w", err)
	}
	
	// Wrap QUIC connection with quicmoq to implement moqtransport.Connection
	conn := quicmoq.NewClient(quicConn)
	
	// Create publisher slot for client-to-server track
	sendSlot := newPublisherSlot()
	
	// Create MOQT session with no-op handler and subscribe handler
	session := &moqtransport.Session{
		Handler: noOpHandler{},
		SubscribeHandler: &subscribeHandler{
			sessionID: "", // Will be set when discovery completes
			sendTrack: trackClientToServer,
			sendSlot:  sendSlot,
		},
	}
	
	// Run MOQT session
	if err := runSession(ctx, session, conn); err != nil {
		return nil, fmt.Errorf("failed to run MOQT session: %w", err)
	}
	
	// Discover session ID via fetch
	sessionID, err := t.discoverSessionID(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to discover session ID: %w", err)
	}
	
	// Subscribe to server-to-client control track
	recv, err := session.Subscribe(ctx, controlNamespace(sessionID), trackServerToClient)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to server-to-client control track: %w", err)
	}
	
	// Create and return control connection
	return newControlConn(conn, sessionID, recv, sendSlot), nil
}

// discoverSessionID discovers the session ID from the server using the discovery track.
// According to the draft, clients should use FETCH (not Subscribe) for discovery.
func (t *MOQTClientTransport) discoverSessionID(ctx context.Context, session *moqtransport.Session) (string, error) {
	// Use FETCH to access discovery track (as per draft specification)
	namespace := []string{"mcp", "discovery"}
	fetch, err := session.Fetch(ctx, namespace, "sessions")
	if err != nil {
		return "", fmt.Errorf("failed to fetch discovery track: %w", err)
	}
	defer fetch.Close()

	// Read discovery response
	obj, err := fetch.ReadObject(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read discovery response: %w", err)
	}

	// Parse discovery response to extract session ID
	var discoveryResp struct {
		Result struct {
			SessionID string `json:"session_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(obj.Payload, &discoveryResp); err != nil {
		return "", fmt.Errorf("failed to parse discovery response: %w", err)
	}

	// Validate session ID
	if discoveryResp.Result.SessionID == "" {
		return "", fmt.Errorf("empty session ID in discovery response")
	}

	return discoveryResp.Result.SessionID, nil
}


