package mcpmoqt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mengelbart/moqtransport"
	"github.com/mengelbart/moqtransport/quicmoq"
)

type MOQTClientTransport struct {
	cfg    *transportConfig
	connPool *QUICConnectionPool
}

func NewMOQTClientTransport(opts ...Option) (*MOQTClientTransport, error) {
	cfg, err := applyOptions(roleClient, opts)
	if err != nil {
		return nil, err
	}

	if cfg.tlsClient == nil {
		tlsCfg, err := defaultTLSConfig(roleClient, cfg.alpn)
		if err != nil {
			return nil, err
		}
		cfg.tlsClient = tlsCfg
	}

	// Create QUIC connection pool
	connPool := NewQUICConnectionPool(DefaultQUICConnectionConfig(), 10, 30*time.Second)
	connPool.SetTLSConfig(cfg.tlsClient, nil)
	connPool.StartCleanup(10 * time.Second)

	return &MOQTClientTransport{cfg: cfg, connPool: connPool}, nil
}

func (t *MOQTClientTransport) Connect(ctx context.Context) (Connection, error) {
	// Get connection from pool
	quicConn, err := t.connPool.Get(ctx, t.cfg.addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get QUIC connection from pool: %w", err)
	}

	conn := quicmoq.NewClient(quicConn)

	sendSlot := NewPublisherSlot()

	session := &moqtransport.Session{
		InitialMaxRequestID: DefaultMaxRequestID,
		Handler: noOpHandler{},
		SubscribeHandler: &subscribeHandler{
			sessionID: "",
			sendTrack: trackClientToServer,
			sendSlot:  sendSlot,
		},
	}

	if err := runSession(ctx, session, conn); err != nil {
		t.connPool.Put(quicConn)
		return nil, fmt.Errorf("failed to run MOQT session: %w", err)
	}

	sessionID, err := t.discoverSessionID(ctx, session)
	if err != nil {
		t.connPool.Put(quicConn)
		return nil, fmt.Errorf("failed to discover session ID: %w", err)
	}

	recv, err := session.Subscribe(ctx, controlNamespace(sessionID), trackServerToClient)
	if err != nil {
		t.connPool.Put(quicConn)
		return nil, fmt.Errorf("failed to subscribe to server-to-client control track: %w", err)
	}

	// Create control connection
	controlConn := newControlConn(conn, sessionID, recv, sendSlot)
	return controlConn, nil
}

func (t *MOQTClientTransport) Close() error {
	if t.connPool != nil {
		return t.connPool.Close()
	}
	return nil
}

func (t *MOQTClientTransport) discoverSessionID(ctx context.Context, session *moqtransport.Session) (string, error) {
	namespace := []string{"mcp", "discovery"}
	fetch, err := session.Fetch(ctx, namespace, "sessions")
	if err != nil {
		return "", fmt.Errorf("failed to fetch discovery track: %w", err)
	}
	defer fetch.Close()

	obj, err := fetch.ReadObject(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read discovery response: %w", err)
	}

	var discoveryResp struct {
		Result struct {
			SessionID string `json:"session_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(obj.Payload, &discoveryResp); err != nil {
		return "", fmt.Errorf("failed to parse discovery response: %w", err)
	}

	if discoveryResp.Result.SessionID == "" {
		return "", fmt.Errorf("empty session ID in discovery response")
	}

	return discoveryResp.Result.SessionID, nil
}
