package mcpmoqt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mengelbart/moqtransport"
	"github.com/mengelbart/moqtransport/quicmoq"
)

type MOQTServerTransport struct {
	cfg       *transportConfig
	sessionID string
	connPool  *QUICConnectionPool
	
	// Multi-connection support
	connections     map[Connection]bool
	connectionsMu   sync.RWMutex
	listener        *quicListener
	closed          bool
	closeMu         sync.Mutex
}

type quicListener struct {
	ln        interface{}
	closeFunc func() error
}

func NewMOQTServerTransport(opts ...Option) (*MOQTServerTransport, error) {
	cfg, err := applyOptions(roleServer, opts)
	if err != nil {
		return nil, err
	}

	if cfg.tlsServer == nil {
		tlsCfg, err := defaultTLSConfig(roleServer, cfg.alpn)
		if err != nil {
			return nil, err
		}
		cfg.tlsServer = tlsCfg
	}

	// Create QUIC connection pool for server-side connections
	connPool := NewQUICConnectionPool(DefaultQUICConnectionConfig(), 10, 30*time.Second)
	connPool.SetTLSConfig(nil, cfg.tlsServer)
	connPool.StartCleanup(10 * time.Second)

	return &MOQTServerTransport{
		cfg:         cfg,
		sessionID:   generateSessionID(),
		connPool:    connPool,
		connections: make(map[Connection]bool),
	}, nil
}

func (t *MOQTServerTransport) Connect(ctx context.Context) (Connection, error) {
	quicConn, ln, _, err := listenAndAcceptMOQT(ctx, t.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to listen and accept QUIC connection: %w", err)
	}

	// Store listener reference
	t.closeMu.Lock()
	if !t.closed {
		t.listener = &quicListener{
			ln:        ln,
			closeFunc: func() error { return ln.Close() },
		}
	}
	t.closeMu.Unlock()

	conn := quicmoq.NewServer(quicConn)

	sendSlot := NewPublisherSlot()

	session := &moqtransport.Session{
		InitialMaxRequestID: DefaultMaxRequestID,
		Handler: &discoveryHandler{
			sessionID: t.sessionID,
		},
		SubscribeHandler: &subscribeHandler{
			sessionID: t.sessionID,
			sendTrack: trackServerToClient,
			sendSlot:  sendSlot,
		},
	}

	if err := runSession(ctx, session, conn); err != nil {
		ln.Close()
		return nil, fmt.Errorf("failed to run MOQT session: %w", err)
	}

	recv, err := session.Subscribe(ctx, controlNamespace(t.sessionID), trackClientToServer)
	if err != nil {
		ln.Close()
		return nil, fmt.Errorf("failed to subscribe to client-to-server control track: %w", err)
	}

	// Create control connection
	controlConn := newControlConn(conn, t.sessionID, recv, sendSlot)

	// Set OnClose callback
	controlConn.OnClose = func() {
		t.RemoveConnection(controlConn)
	}

	// Add to connections map
	t.connectionsMu.Lock()
	t.connections[controlConn] = true
	t.connectionsMu.Unlock()

	return controlConn, nil
}

func (t *MOQTServerTransport) Close() error {
	t.closeMu.Lock()
	if t.closed {
		t.closeMu.Unlock()
		return nil
	}
	t.closed = true
	t.closeMu.Unlock()

	// Close all active connections
	t.connectionsMu.Lock()
	for conn := range t.connections {
		_ = conn.Close()
	}
	t.connections = make(map[Connection]bool)
	t.connectionsMu.Unlock()

	// Close listener
	if t.listener != nil {
		_ = t.listener.closeFunc()
	}

	// Close connection pool
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
