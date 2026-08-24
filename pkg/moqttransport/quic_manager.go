package mcpmoqt

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"
)

var (
	ErrConnectionPoolClosed  = errors.New("connection pool is closed")
	ErrNoAvailableConnection = errors.New("no available connection in pool")
	ErrPoolFull              = errors.New("connection pool is full")
)

type QUICConnectionConfig struct {
	MaxIdleTimeout                 time.Duration
	KeepAlivePeriod                time.Duration
	MaxIncomingStreams             int64
	MaxIncomingUniStreams          int64
	EnableDatagrams                bool
	InitialStreamReceiveWindow     uint64
	InitialConnectionReceiveWindow uint64
	MaxStreamReceiveWindow         uint64
	MaxConnectionReceiveWindow     uint64
	Allow0RTT                      bool
}

func DefaultQUICConnectionConfig() QUICConnectionConfig {
	return QUICConnectionConfig{
		MaxIdleTimeout:                 30 * time.Second,
		KeepAlivePeriod:                15 * time.Second,
		MaxIncomingStreams:             100,
		MaxIncomingUniStreams:          100,
		EnableDatagrams:                true,
		InitialStreamReceiveWindow:     1 << 20,
		InitialConnectionReceiveWindow: 1 << 20,
		MaxStreamReceiveWindow:         6 << 20,
		MaxConnectionReceiveWindow:     15 << 20,
		Allow0RTT:                      false,
	}
}

func (c QUICConnectionConfig) ToQUICConfig() *quic.Config {
	return &quic.Config{
		MaxIdleTimeout:                 c.MaxIdleTimeout,
		KeepAlivePeriod:                c.KeepAlivePeriod,
		MaxIncomingStreams:             c.MaxIncomingStreams,
		MaxIncomingUniStreams:          c.MaxIncomingUniStreams,
		EnableDatagrams:                c.EnableDatagrams,
		InitialStreamReceiveWindow:     c.InitialStreamReceiveWindow,
		InitialConnectionReceiveWindow: c.InitialConnectionReceiveWindow,
		MaxStreamReceiveWindow:         c.MaxStreamReceiveWindow,
		MaxConnectionReceiveWindow:     c.MaxConnectionReceiveWindow,
		Allow0RTT:                      c.Allow0RTT,
	}
}

type pooledConnection struct {
	conn       *quic.Conn
	addr       string
	createdAt  time.Time
	lastUsedAt time.Time
	inUse      atomic.Bool
}

func (pc *pooledConnection) markInUse() {
	pc.lastUsedAt = time.Now()
	pc.inUse.Store(true)
}

func (pc *pooledConnection) markIdle() {
	pc.inUse.Store(false)
}

func (pc *pooledConnection) isExpired(maxIdle time.Duration) bool {
	if pc.inUse.Load() {
		return false
	}
	return time.Since(pc.lastUsedAt) > maxIdle
}

type QUICConnectionPool struct {
	mu          sync.RWMutex
	connections map[string][]*pooledConnection
	config      QUICConnectionConfig
	tlsClient   *tls.Config
	tlsServer   *tls.Config
	maxPoolSize int
	maxIdleTime time.Duration
	closed      atomic.Bool
	done        chan struct{}

	stats QUICPoolStats
}

type QUICPoolStats struct {
	TotalConnections   int64
	ActiveConnections  int64
	IdleConnections    int64
	ConnectionsCreated int64
	ConnectionsClosed  int64
	PoolHits           int64
	PoolMisses         int64
}

func NewQUICConnectionPool(config QUICConnectionConfig, maxPoolSize int, maxIdleTime time.Duration) *QUICConnectionPool {
	return &QUICConnectionPool{
		connections: make(map[string][]*pooledConnection),
		config:      config,
		maxPoolSize: maxPoolSize,
		maxIdleTime: maxIdleTime,
		done:        make(chan struct{}),
	}
}

func (p *QUICConnectionPool) SetTLSConfig(client, server *tls.Config) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tlsClient = client
	p.tlsServer = server
}

func (p *QUICConnectionPool) Get(ctx context.Context, addr string) (*quic.Conn, error) {
	if p.closed.Load() {
		return nil, ErrConnectionPoolClosed
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if conns, ok := p.connections[addr]; ok && len(conns) > 0 {
		for _, pc := range conns {
			if !pc.inUse.Load() && !pc.isExpired(p.maxIdleTime) {
				pc.markInUse()
				p.stats.PoolHits++
				p.stats.IdleConnections--
				p.stats.ActiveConnections++
				return pc.conn, nil
			}
		}
	}

	if p.maxPoolSize > 0 && p.stats.TotalConnections >= int64(p.maxPoolSize) {
		return nil, ErrPoolFull
	}

	p.stats.PoolMisses++

	conn, err := p.dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	pc := &pooledConnection{
		conn:       conn,
		addr:       addr,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
	}
	pc.markInUse()

	p.connections[addr] = append(p.connections[addr], pc)
	p.stats.TotalConnections++
	p.stats.ActiveConnections++
	p.stats.ConnectionsCreated++

	return conn, nil
}

func (p *QUICConnectionPool) Put(conn *quic.Conn) {
	if p.closed.Load() {
		conn.CloseWithError(0, "pool closed")
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	addr := conn.RemoteAddr().String()
	if conns, ok := p.connections[addr]; ok {
		for _, pc := range conns {
			if pc.conn == conn {
				pc.markIdle()
				p.stats.ActiveConnections--
				p.stats.IdleConnections++
				return
			}
		}
	}
}

func (p *QUICConnectionPool) dial(ctx context.Context, addr string) (*quic.Conn, error) {
	if p.tlsClient == nil {
		return nil, errors.New("TLS client config not set")
	}

	return quic.DialAddr(ctx, addr, p.tlsClient, p.config.ToQUICConfig())
}

func (p *QUICConnectionPool) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(p.done)

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conns := range p.connections {
		for _, pc := range conns {
			pc.conn.CloseWithError(0, "pool closed")
			p.stats.ConnectionsClosed++
		}
	}

	p.connections = make(map[string][]*pooledConnection)
	p.stats.TotalConnections = 0
	p.stats.ActiveConnections = 0
	p.stats.IdleConnections = 0

	return nil
}

func (p *QUICConnectionPool) CleanupExpired() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for addr, conns := range p.connections {
		var active []*pooledConnection
		for _, pc := range conns {
			if pc.isExpired(p.maxIdleTime) {
				pc.conn.CloseWithError(0, "connection expired")
				p.stats.ConnectionsClosed++
				p.stats.TotalConnections--
				p.stats.IdleConnections--
			} else {
				active = append(active, pc)
			}
		}
		if len(active) == 0 {
			delete(p.connections, addr)
		} else {
			p.connections[addr] = active
		}
	}
}

func (p *QUICConnectionPool) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-p.done:
				return
			case <-ticker.C:
				p.CleanupExpired()
			}
		}
	}()
}

func (p *QUICConnectionPool) Stats() QUICPoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

func (p *QUICConnectionPool) IsFull() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.maxPoolSize > 0 && p.stats.TotalConnections >= int64(p.maxPoolSize)
}

func (p *QUICConnectionPool) Available() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.maxPoolSize <= 0 {
		return -1
	}
	available := p.maxPoolSize - int(p.stats.TotalConnections)
	if available < 0 {
		return 0
	}
	return available
}

func (p *QUICConnectionPool) IsHealthy() bool {
	if p.closed.Load() {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats.ActiveConnections >= 0 && p.stats.IdleConnections >= 0
}

type QUICStreamManager struct {
	mu      sync.RWMutex
	streams map[int64]*quic.Stream
	nextID  atomic.Int64
	closed  atomic.Bool
	done    chan struct{}

	stats QUICStreamStats
}

type QUICStreamStats struct {
	TotalStreams  int64
	ActiveStreams int64
	StreamsOpened int64
	StreamsClosed int64
	BytesSent     int64
	BytesReceived int64
}

func NewQUICStreamManager() *QUICStreamManager {
	return &QUICStreamManager{
		streams: make(map[int64]*quic.Stream),
		done:    make(chan struct{}),
	}
}

func (m *QUICStreamManager) OpenStream(conn *quic.Conn) (*quic.Stream, int64, error) {
	if m.closed.Load() {
		return nil, 0, ErrConnectionClosed
	}

	stream, err := conn.OpenStream()
	if err != nil {
		return nil, 0, err
	}

	id := m.nextID.Add(1)

	m.mu.Lock()
	m.streams[id] = stream
	m.stats.TotalStreams++
	m.stats.ActiveStreams++
	m.stats.StreamsOpened++
	m.mu.Unlock()

	return stream, id, nil
}

func (m *QUICStreamManager) OpenStreamSync(ctx context.Context, conn *quic.Conn) (*quic.Stream, int64, error) {
	if m.closed.Load() {
		return nil, 0, ErrConnectionClosed
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, 0, err
	}

	id := m.nextID.Add(1)

	m.mu.Lock()
	m.streams[id] = stream
	m.stats.TotalStreams++
	m.stats.ActiveStreams++
	m.stats.StreamsOpened++
	m.mu.Unlock()

	return stream, id, nil
}

func (m *QUICStreamManager) OpenUniStream(conn *quic.Conn) (*quic.SendStream, int64, error) {
	if m.closed.Load() {
		return nil, 0, ErrConnectionClosed
	}

	stream, err := conn.OpenUniStream()
	if err != nil {
		return nil, 0, err
	}

	id := m.nextID.Add(1)

	m.mu.Lock()
	m.stats.TotalStreams++
	m.stats.ActiveStreams++
	m.stats.StreamsOpened++
	m.mu.Unlock()

	return stream, id, nil
}

func (m *QUICStreamManager) CloseStream(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stream, ok := m.streams[id]; ok {
		delete(m.streams, id)
		m.stats.TotalStreams--
		m.stats.ActiveStreams--
		m.stats.StreamsClosed++
		return stream.Close()
	}

	return nil
}

func (m *QUICStreamManager) GetStream(id int64) (*quic.Stream, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stream, ok := m.streams[id]
	return stream, ok
}

func (m *QUICStreamManager) Close() error {
	if !m.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(m.done)

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, stream := range m.streams {
		if stream != nil {
			stream.Close()
		}
		delete(m.streams, id)
		m.stats.StreamsClosed++
	}

	m.streams = make(map[int64]*quic.Stream)
	m.stats.TotalStreams = 0
	m.stats.ActiveStreams = 0

	return nil
}

func (m *QUICStreamManager) Stats() QUICStreamStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

func (m *QUICStreamManager) RecordBytesSent(n int) {
	m.mu.Lock()
	m.stats.BytesSent += int64(n)
	m.mu.Unlock()
}

func (m *QUICStreamManager) RecordBytesReceived(n int) {
	m.mu.Lock()
	m.stats.BytesReceived += int64(n)
	m.mu.Unlock()
}

type QUICConnectionState struct {
	RemoteAddr       net.Addr
	LocalAddr        net.Addr
	IsReady          bool
	LastActivity     time.Time
	ConnectionUptime time.Duration
}

func GetConnectionState(conn *quic.Conn) QUICConnectionState {
	return QUICConnectionState{
		RemoteAddr:       conn.RemoteAddr(),
		LocalAddr:        conn.LocalAddr(),
		IsReady:          true,
		LastActivity:     time.Now(),
		ConnectionUptime: 0,
	}
}

type QUICConfigBuilder struct {
	config *quic.Config
}

func NewQUICConfigBuilder() *QUICConfigBuilder {
	return &QUICConfigBuilder{
		config: &quic.Config{
			EnableDatagrams: true,
		},
	}
}

func (b *QUICConfigBuilder) WithMaxIdleTimeout(d time.Duration) *QUICConfigBuilder {
	b.config.MaxIdleTimeout = d
	return b
}

func (b *QUICConfigBuilder) WithKeepAlivePeriod(d time.Duration) *QUICConfigBuilder {
	b.config.KeepAlivePeriod = d
	return b
}

func (b *QUICConfigBuilder) WithMaxIncomingStreams(n int64) *QUICConfigBuilder {
	b.config.MaxIncomingStreams = n
	return b
}

func (b *QUICConfigBuilder) WithMaxIncomingUniStreams(n int64) *QUICConfigBuilder {
	b.config.MaxIncomingUniStreams = n
	return b
}

func (b *QUICConfigBuilder) WithDatagrams(enabled bool) *QUICConfigBuilder {
	b.config.EnableDatagrams = enabled
	return b
}

func (b *QUICConfigBuilder) WithInitialStreamReceiveWindow(n uint64) *QUICConfigBuilder {
	b.config.InitialStreamReceiveWindow = n
	return b
}

func (b *QUICConfigBuilder) WithInitialConnectionReceiveWindow(n uint64) *QUICConfigBuilder {
	b.config.InitialConnectionReceiveWindow = n
	return b
}

func (b *QUICConfigBuilder) WithMaxStreamReceiveWindow(n uint64) *QUICConfigBuilder {
	b.config.MaxStreamReceiveWindow = n
	return b
}

func (b *QUICConfigBuilder) WithMaxConnectionReceiveWindow(n uint64) *QUICConfigBuilder {
	b.config.MaxConnectionReceiveWindow = n
	return b
}

func (b *QUICConfigBuilder) With0RTT(enabled bool) *QUICConfigBuilder {
	b.config.Allow0RTT = enabled
	return b
}

func (b *QUICConfigBuilder) Build() *quic.Config {
	return b.config
}
