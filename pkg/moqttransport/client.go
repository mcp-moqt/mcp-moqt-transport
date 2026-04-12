package mcpmoqt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mengelbart/moqtransport"
	"github.com/mengelbart/moqtransport/quicmoq"
)

type MOQTClientTransport struct {
	cfg *transportConfig
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

	return &MOQTClientTransport{cfg: cfg}, nil
}

func (t *MOQTClientTransport) Connect(ctx context.Context) (Connection, error) {
	quicConn, _, err := dialMOQT(ctx, t.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to dial QUIC connection: %w", err)
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
		return nil, fmt.Errorf("failed to run MOQT session: %w", err)
	}

	sessionID, err := t.discoverSessionID(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to discover session ID: %w", err)
	}

	recv, err := session.Subscribe(ctx, controlNamespace(sessionID), trackServerToClient)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to server-to-client control track: %w", err)
	}

	return newControlConn(conn, sessionID, recv, sendSlot), nil
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
