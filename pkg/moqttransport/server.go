package mcpmoqt

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mengelbart/moqtransport"
	"github.com/mengelbart/moqtransport/quicmoq"
	"github.com/quic-go/quic-go"
)

type MOQTServerTransport struct {
	cfg      *transportConfig
	connPool *QUICConnectionPool

	connections   map[Connection]bool
	connectionsMu sync.RWMutex

	// sessionHandlers maps MCP session ID -> data track handlers for that connection.
	sessionHandlers map[string][]*DataTrackHandler
	sessionsMu      sync.RWMutex

	quicLn   *quic.Listener
	listenMu sync.Mutex
	closed   bool
	closeMu  sync.Mutex

	metricsAddr   string
	metricsServer *MetricsHTTPServer
}

func NewMOQTServerTransport(opts ...Option) (*MOQTServerTransport, error) {
	cfg, err := applyOptions(roleServer, opts)
	if err != nil {
		return nil, err
	}

	if _, err := resolveServerTLS(cfg); err != nil {
		return nil, err
	}

	poolCfg := DefaultQUICConnectionConfig()
	if cfg.quicConfig != nil {
		poolCfg = quicConfigToPoolConfig(cfg.quicConfig, poolCfg)
	}
	connPool := NewQUICConnectionPool(poolCfg, 10, 30*time.Second)
	connPool.SetTLSConfig(nil, cfg.tlsServer)
	connPool.StartCleanup(10 * time.Second)

	t := &MOQTServerTransport{
		cfg:             cfg,
		connPool:        connPool,
		connections:     make(map[Connection]bool),
		sessionHandlers: make(map[string][]*DataTrackHandler),
		metricsAddr:     cfg.metricsAddr,
	}
	return t, nil
}

// Listen starts the QUIC listener if not already started.
func (t *MOQTServerTransport) Listen() error {
	t.listenMu.Lock()
	defer t.listenMu.Unlock()

	t.closeMu.Lock()
	closed := t.closed
	t.closeMu.Unlock()
	if closed {
		return fmt.Errorf("server transport is closed")
	}

	if t.quicLn != nil {
		return nil
	}

	ln, err := quic.ListenAddr(t.cfg.addr, t.cfg.tlsServer, t.cfg.quicConfig)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	t.quicLn = ln

	if t.metricsAddr != "" && t.metricsServer == nil {
		t.metricsServer = NewMetricsHTTPServer(t.metricsAddr, DefaultMetrics())
		go func() {
			_ = t.metricsServer.Start()
		}()
	}

	return nil
}

// Addr returns the listener address after Listen, or empty string.
func (t *MOQTServerTransport) Addr() net.Addr {
	t.listenMu.Lock()
	defer t.listenMu.Unlock()
	if t.quicLn == nil {
		return nil
	}
	return t.quicLn.Addr()
}

// Connect implements mcp.Transport: listens (once) and accepts the next QUIC connection.
func (t *MOQTServerTransport) Connect(ctx context.Context) (Connection, error) {
	if err := t.Listen(); err != nil {
		return nil, err
	}
	return t.Accept(ctx)
}

// Accept waits for the next QUIC client and returns an MCP Connection.
func (t *MOQTServerTransport) Accept(ctx context.Context) (Connection, error) {
	t.listenMu.Lock()
	ln := t.quicLn
	t.listenMu.Unlock()
	if ln == nil {
		return nil, fmt.Errorf("listener not started; call Listen or Connect first")
	}

	quicConn, err := ln.Accept(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to accept: %w", err)
	}

	return t.setupConnection(ctx, quicConn)
}

// Serve repeatedly Accepts connections and invokes handle for each (typically in a new goroutine).
// handle is called synchronously per accept by default; callers should spawn goroutines inside
// handle if they need concurrent sessions.
func (t *MOQTServerTransport) Serve(ctx context.Context, handle func(context.Context, Connection) error) error {
	if handle == nil {
		return fmt.Errorf("handle must not be nil")
	}
	if err := t.Listen(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, err := t.Accept(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return err
			}
		}

		go func(c Connection) {
			_ = handle(ctx, c)
		}(conn)
	}
}

func (t *MOQTServerTransport) setupConnection(ctx context.Context, quicConn *quic.Conn) (Connection, error) {
	sessionID := generateSessionID()
	dataHandlers := defaultDataTrackHandlers(sessionID)

	conn := quicmoq.NewServer(quicConn)
	sendSlot := NewPublisherSlot()
	control := &subscribeHandler{
		sessionID: sessionID,
		sendTrack: trackServerToClient,
		sendSlot:  sendSlot,
	}

	session := &moqtransport.Session{
		InitialMaxRequestID: DefaultMaxRequestID,
		Handler: &discoveryHandler{
			sessionID: sessionID,
		},
		SubscribeHandler: &compositeSubscribeHandler{
			control: control,
			data:    dataHandlers,
		},
	}

	if err := runSession(ctx, session, conn); err != nil {
		_ = quicConn.CloseWithError(0, "session failed")
		return nil, fmt.Errorf("failed to run MOQT session: %w", err)
	}

	recv, err := session.Subscribe(ctx, controlNamespace(sessionID), trackClientToServer)
	if err != nil {
		_ = quicConn.CloseWithError(0, "subscribe failed")
		return nil, fmt.Errorf("failed to subscribe to client-to-server control track: %w", err)
	}

	controlConn := newControlConn(conn, sessionID, recv, sendSlot)
	controlConn.OnClose = func() {
		t.RemoveConnection(controlConn)
		t.sessionsMu.Lock()
		handlers := t.sessionHandlers[sessionID]
		delete(t.sessionHandlers, sessionID)
		t.sessionsMu.Unlock()
		for _, h := range handlers {
			_ = h.Close()
		}
	}

	t.connectionsMu.Lock()
	t.connections[controlConn] = true
	t.connectionsMu.Unlock()

	t.sessionsMu.Lock()
	t.sessionHandlers[sessionID] = dataHandlers
	t.sessionsMu.Unlock()

	return controlConn, nil
}

func defaultDataTrackHandlers(sessionID string) []*DataTrackHandler {
	return []*DataTrackHandler{
		NewDataTrackHandler(sessionID, DataTrackResources),
		NewDataTrackHandler(sessionID, DataTrackTools),
		NewDataTrackHandler(sessionID, DataTrackNotifications),
	}
}

// DataTrackHandler returns a data-track handler for sessionID (empty = any/first).
func (t *MOQTServerTransport) DataTrackHandler(trackType DataTrackType) *DataTrackHandler {
	return t.DataTrackHandlerFor("", trackType)
}

// DataTrackHandlerFor returns the handler for a specific MCP session and track type.
func (t *MOQTServerTransport) DataTrackHandlerFor(sessionID string, trackType DataTrackType) *DataTrackHandler {
	t.sessionsMu.RLock()
	defer t.sessionsMu.RUnlock()
	if sessionID != "" {
		for _, h := range t.sessionHandlers[sessionID] {
			if h.TrackType() == trackType {
				return h
			}
		}
		return nil
	}
	for _, handlers := range t.sessionHandlers {
		for _, h := range handlers {
			if h.TrackType() == trackType {
				return h
			}
		}
	}
	return nil
}

// SessionID returns an arbitrary active session ID (for single-connection demos/tests).
func (t *MOQTServerTransport) SessionID() string {
	t.sessionsMu.RLock()
	defer t.sessionsMu.RUnlock()
	for id := range t.sessionHandlers {
		return id
	}
	return ""
}

func (t *MOQTServerTransport) Close() error {
	t.closeMu.Lock()
	if t.closed {
		t.closeMu.Unlock()
		return nil
	}
	t.closed = true
	t.closeMu.Unlock()

	t.connectionsMu.Lock()
	conns := make([]Connection, 0, len(t.connections))
	for conn := range t.connections {
		conns = append(conns, conn)
	}
	t.connections = make(map[Connection]bool)
	t.connectionsMu.Unlock()

	for _, conn := range conns {
		_ = conn.Close()
	}

	t.sessionsMu.Lock()
	for id, handlers := range t.sessionHandlers {
		for _, h := range handlers {
			_ = h.Close()
		}
		delete(t.sessionHandlers, id)
	}
	t.sessionsMu.Unlock()

	t.listenMu.Lock()
	if t.quicLn != nil {
		_ = t.quicLn.Close()
		t.quicLn = nil
	}
	t.listenMu.Unlock()

	if t.metricsServer != nil {
		_ = t.metricsServer.Close()
	}

	if t.connPool != nil {
		return t.connPool.Close()
	}
	return nil
}

func (t *MOQTServerTransport) RemoveConnection(conn Connection) {
	t.connectionsMu.Lock()
	delete(t.connections, conn)
	t.connectionsMu.Unlock()
}

func (t *MOQTServerTransport) ActiveConnections() int {
	t.connectionsMu.RLock()
	defer t.connectionsMu.RUnlock()
	return len(t.connections)
}

// PreconnectedTransport wraps an already-accepted Connection for mcp.Server.Connect.
type PreconnectedTransport struct {
	Conn Connection
}

func (p *PreconnectedTransport) Connect(context.Context) (Connection, error) {
	if p.Conn == nil {
		return nil, fmt.Errorf("preconnected connection is nil")
	}
	return p.Conn, nil
}
