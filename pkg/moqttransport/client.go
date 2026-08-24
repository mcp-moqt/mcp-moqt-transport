package mcpmoqt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mengelbart/moqtransport"
	"github.com/mengelbart/moqtransport/quicmoq"
	"github.com/quic-go/quic-go"
)

type MOQTClientTransport struct {
	cfg      *transportConfig
	connPool *QUICConnectionPool

	// Last established MOQT session (for data-track attach helpers).
	session   *moqtransport.Session
	sessionID string
	quicConn  *quic.Conn
}

func NewMOQTClientTransport(opts ...Option) (*MOQTClientTransport, error) {
	cfg, err := applyOptions(roleClient, opts)
	if err != nil {
		return nil, err
	}

	if _, err := resolveClientTLS(cfg); err != nil {
		return nil, err
	}

	poolCfg := DefaultQUICConnectionConfig()
	if cfg.quicConfig != nil {
		poolCfg = quicConfigToPoolConfig(cfg.quicConfig, poolCfg)
	}
	connPool := NewQUICConnectionPool(poolCfg, 10, 30*time.Second)
	connPool.SetTLSConfig(cfg.tlsClient, nil)
	connPool.StartCleanup(10 * time.Second)

	return &MOQTClientTransport{cfg: cfg, connPool: connPool}, nil
}

func (t *MOQTClientTransport) Connect(ctx context.Context) (Connection, error) {
	quicConn, err := t.connPool.Get(ctx, t.cfg.addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get QUIC connection from pool: %w", err)
	}

	conn := quicmoq.NewClient(quicConn)
	sendSlot := NewPublisherSlot()

	session := &moqtransport.Session{
		InitialMaxRequestID: DefaultMaxRequestID,
		Handler:             noOpHandler{},
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

	t.session = session
	t.sessionID = sessionID
	t.quicConn = quicConn

	controlConn := newControlConn(conn, sessionID, recv, sendSlot, func() {
		// Return QUIC connection to pool when MCP control session ends.
		t.connPool.Put(quicConn)
		t.session = nil
		t.quicConn = nil
	})
	return controlConn, nil
}

// SessionID returns the discovered MCP session ID after Connect, if available.
func (t *MOQTClientTransport) SessionID() string {
	return t.sessionID
}

// SubscribeDataTrack registers a local handler and attaches a MOQT subscribe
// to the matching data-track namespace for the discovered session.
func (t *MOQTClientTransport) SubscribeDataTrack(ctx context.Context, trackType DataTrackType, trackName string, handler func(*DataTrackMessage) error) (*DataTrackHandler, error) {
	if t.session == nil || t.sessionID == "" {
		return nil, fmt.Errorf("client session not connected")
	}
	dh := NewDataTrackHandler(t.sessionID, trackType)
	if err := dh.Subscribe(ctx, trackName, handler); err != nil {
		return nil, err
	}
	if err := dh.AttachSessionSubscribe(ctx, t.session, trackName); err != nil {
		_ = dh.Close()
		return nil, err
	}
	return dh, nil
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

// quicConfigToPoolConfig overlays selected fields from *quic.Config onto pool defaults.
func quicConfigToPoolConfig(src *quic.Config, base QUICConnectionConfig) QUICConnectionConfig {
	if src == nil {
		return base
	}
	if src.MaxIdleTimeout > 0 {
		base.MaxIdleTimeout = src.MaxIdleTimeout
	}
	if src.KeepAlivePeriod > 0 {
		base.KeepAlivePeriod = src.KeepAlivePeriod
	}
	if src.MaxIncomingStreams > 0 {
		base.MaxIncomingStreams = src.MaxIncomingStreams
	}
	if src.MaxIncomingUniStreams > 0 {
		base.MaxIncomingUniStreams = src.MaxIncomingUniStreams
	}
	base.EnableDatagrams = src.EnableDatagrams
	if src.InitialStreamReceiveWindow > 0 {
		base.InitialStreamReceiveWindow = src.InitialStreamReceiveWindow
	}
	if src.InitialConnectionReceiveWindow > 0 {
		base.InitialConnectionReceiveWindow = src.InitialConnectionReceiveWindow
	}
	if src.MaxStreamReceiveWindow > 0 {
		base.MaxStreamReceiveWindow = src.MaxStreamReceiveWindow
	}
	if src.MaxConnectionReceiveWindow > 0 {
		base.MaxConnectionReceiveWindow = src.MaxConnectionReceiveWindow
	}
	base.Allow0RTT = src.Allow0RTT
	return base
}
