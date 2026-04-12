package mcpmoqt

import (
	"sync"
	"testing"
	"time"
)

func TestDefaultQUICConnectionConfig(t *testing.T) {
	config := DefaultQUICConnectionConfig()

	if config.MaxIdleTimeout != 30*time.Second {
		t.Errorf("Expected MaxIdleTimeout 30s, got %v", config.MaxIdleTimeout)
	}

	if config.KeepAlivePeriod != 15*time.Second {
		t.Errorf("Expected KeepAlivePeriod 15s, got %v", config.KeepAlivePeriod)
	}

	if config.MaxIncomingStreams != 100 {
		t.Errorf("Expected MaxIncomingStreams 100, got %d", config.MaxIncomingStreams)
	}

	if !config.EnableDatagrams {
		t.Error("Expected EnableDatagrams to be true")
	}

	if config.Allow0RTT {
		t.Error("Expected Allow0RTT to be false by default")
	}
}

func TestQUICConnectionConfigToQUICConfig(t *testing.T) {
	config := QUICConnectionConfig{
		MaxIdleTimeout:                 60 * time.Second,
		KeepAlivePeriod:                30 * time.Second,
		MaxIncomingStreams:             200,
		MaxIncomingUniStreams:          150,
		EnableDatagrams:                true,
		InitialStreamReceiveWindow:     1 << 21,
		InitialConnectionReceiveWindow: 1 << 22,
		MaxStreamReceiveWindow:         1 << 23,
		MaxConnectionReceiveWindow:     1 << 24,
		Allow0RTT:                      true,
	}

	quicConfig := config.ToQUICConfig()

	if quicConfig.MaxIdleTimeout != 60*time.Second {
		t.Errorf("Expected MaxIdleTimeout 60s, got %v", quicConfig.MaxIdleTimeout)
	}

	if quicConfig.KeepAlivePeriod != 30*time.Second {
		t.Errorf("Expected KeepAlivePeriod 30s, got %v", quicConfig.KeepAlivePeriod)
	}

	if quicConfig.MaxIncomingStreams != 200 {
		t.Errorf("Expected MaxIncomingStreams 200, got %d", quicConfig.MaxIncomingStreams)
	}

	if !quicConfig.EnableDatagrams {
		t.Error("Expected EnableDatagrams to be true")
	}

	if !quicConfig.Allow0RTT {
		t.Error("Expected Allow0RTT to be true")
	}
}

func TestNewQUICConnectionPool(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 10, 5*time.Minute)

	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}

	if pool.maxPoolSize != 10 {
		t.Errorf("Expected maxPoolSize 10, got %d", pool.maxPoolSize)
	}

	if pool.maxIdleTime != 5*time.Minute {
		t.Errorf("Expected maxIdleTime 5m, got %v", pool.maxIdleTime)
	}

	if pool.connections == nil {
		t.Error("Expected connections map to be initialized")
	}
}

func TestQUICConnectionPoolStats(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 10, 5*time.Minute)

	stats := pool.Stats()

	if stats.TotalConnections != 0 {
		t.Errorf("Expected TotalConnections 0, got %d", stats.TotalConnections)
	}

	if stats.ActiveConnections != 0 {
		t.Errorf("Expected ActiveConnections 0, got %d", stats.ActiveConnections)
	}

	if stats.IdleConnections != 0 {
		t.Errorf("Expected IdleConnections 0, got %d", stats.IdleConnections)
	}
}

func TestQUICConnectionPoolClose(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 10, 5*time.Minute)

	err := pool.Close()
	if err != nil {
		t.Errorf("Unexpected error closing pool: %v", err)
	}

	if !pool.closed.Load() {
		t.Error("Expected pool to be marked as closed")
	}

	err = pool.Close()
	if err != nil {
		t.Errorf("Unexpected error on second close: %v", err)
	}
}

func TestQUICConnectionPoolCleanupExpired(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 10, 1*time.Millisecond)

	pool.CleanupExpired()

	stats := pool.Stats()
	if stats.ConnectionsClosed != 0 {
		t.Errorf("Expected ConnectionsClosed 0, got %d", stats.ConnectionsClosed)
	}
}

func TestPooledConnectionMarkInUse(t *testing.T) {
	pc := &pooledConnection{
		conn:       nil,
		addr:       "test:1234",
		createdAt:  time.Now(),
		lastUsedAt: time.Now().Add(-1 * time.Hour),
	}

	if pc.inUse.Load() {
		t.Error("Expected inUse to be false initially")
	}

	pc.markInUse()

	if !pc.inUse.Load() {
		t.Error("Expected inUse to be true after markInUse")
	}

	if time.Since(pc.lastUsedAt) > time.Second {
		t.Error("Expected lastUsedAt to be updated")
	}
}

func TestPooledConnectionMarkIdle(t *testing.T) {
	pc := &pooledConnection{}
	pc.markInUse()

	if !pc.inUse.Load() {
		t.Error("Expected inUse to be true")
	}

	pc.markIdle()

	if pc.inUse.Load() {
		t.Error("Expected inUse to be false after markIdle")
	}
}

func TestPooledConnectionIsExpired(t *testing.T) {
	pc := &pooledConnection{
		lastUsedAt: time.Now().Add(-2 * time.Hour),
	}

	pc.markInUse()

	if pc.isExpired(1 * time.Hour) {
		t.Error("Expected connection not to be expired when in use")
	}

	pc.markIdle()
	pc.lastUsedAt = time.Now().Add(-2 * time.Hour)

	if !pc.isExpired(1 * time.Hour) {
		t.Error("Expected connection to be expired")
	}

	if pc.isExpired(3 * time.Hour) {
		t.Error("Expected connection not to be expired with longer timeout")
	}
}

func TestNewQUICStreamManager(t *testing.T) {
	manager := NewQUICStreamManager()

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.streams == nil {
		t.Error("Expected streams map to be initialized")
	}

	if manager.closed.Load() {
		t.Error("Expected manager not to be closed initially")
	}
}

func TestQUICStreamManagerClose(t *testing.T) {
	manager := NewQUICStreamManager()

	err := manager.Close()
	if err != nil {
		t.Errorf("Unexpected error closing manager: %v", err)
	}

	if !manager.closed.Load() {
		t.Error("Expected manager to be marked as closed")
	}

	err = manager.Close()
	if err != nil {
		t.Errorf("Unexpected error on second close: %v", err)
	}
}

func TestQUICStreamManagerStats(t *testing.T) {
	manager := NewQUICStreamManager()

	stats := manager.Stats()

	if stats.TotalStreams != 0 {
		t.Errorf("Expected TotalStreams 0, got %d", stats.TotalStreams)
	}

	if stats.ActiveStreams != 0 {
		t.Errorf("Expected ActiveStreams 0, got %d", stats.ActiveStreams)
	}

	if stats.StreamsOpened != 0 {
		t.Errorf("Expected StreamsOpened 0, got %d", stats.StreamsOpened)
	}
}

func TestQUICStreamManagerRecordBytes(t *testing.T) {
	manager := NewQUICStreamManager()

	manager.RecordBytesSent(100)
	manager.RecordBytesSent(50)

	manager.RecordBytesReceived(200)
	manager.RecordBytesReceived(75)

	stats := manager.Stats()

	if stats.BytesSent != 150 {
		t.Errorf("Expected BytesSent 150, got %d", stats.BytesSent)
	}

	if stats.BytesReceived != 275 {
		t.Errorf("Expected BytesReceived 275, got %d", stats.BytesReceived)
	}
}

func TestQUICStreamManagerGetStream(t *testing.T) {
	manager := NewQUICStreamManager()

	stream, ok := manager.GetStream(1)
	if ok {
		t.Error("Expected GetStream to return false for non-existent stream")
	}
	if stream != nil {
		t.Error("Expected nil stream for non-existent ID")
	}
}

func TestQUICStreamManagerCloseStream(t *testing.T) {
	manager := NewQUICStreamManager()

	err := manager.CloseStream(999)
	if err != nil {
		t.Errorf("Unexpected error closing non-existent stream: %v", err)
	}
}

func TestNewQUICConfigBuilder(t *testing.T) {
	builder := NewQUICConfigBuilder()

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}

	if builder.config == nil {
		t.Fatal("Expected non-nil config")
	}

	if !builder.config.EnableDatagrams {
		t.Error("Expected EnableDatagrams to be true by default")
	}
}

func TestQUICConfigBuilderWithMaxIdleTimeout(t *testing.T) {
	builder := NewQUICConfigBuilder()
	result := builder.WithMaxIdleTimeout(60 * time.Second)

	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}

	if builder.config.MaxIdleTimeout != 60*time.Second {
		t.Errorf("Expected MaxIdleTimeout 60s, got %v", builder.config.MaxIdleTimeout)
	}
}

func TestQUICConfigBuilderWithKeepAlivePeriod(t *testing.T) {
	builder := NewQUICConfigBuilder()
	result := builder.WithKeepAlivePeriod(30 * time.Second)

	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}

	if builder.config.KeepAlivePeriod != 30*time.Second {
		t.Errorf("Expected KeepAlivePeriod 30s, got %v", builder.config.KeepAlivePeriod)
	}
}

func TestQUICConfigBuilderWithMaxIncomingStreams(t *testing.T) {
	builder := NewQUICConfigBuilder()
	result := builder.WithMaxIncomingStreams(500)

	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}

	if builder.config.MaxIncomingStreams != 500 {
		t.Errorf("Expected MaxIncomingStreams 500, got %d", builder.config.MaxIncomingStreams)
	}
}

func TestQUICConfigBuilderWithDatagrams(t *testing.T) {
	builder := NewQUICConfigBuilder()
	result := builder.WithDatagrams(false)

	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}

	if builder.config.EnableDatagrams {
		t.Error("Expected EnableDatagrams to be false")
	}
}

func TestQUICConfigBuilderWith0RTT(t *testing.T) {
	builder := NewQUICConfigBuilder()
	result := builder.With0RTT(true)

	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}

	if !builder.config.Allow0RTT {
		t.Error("Expected Allow0RTT to be true")
	}
}

func TestQUICConfigBuilderWithReceiveWindows(t *testing.T) {
	builder := NewQUICConfigBuilder()

	builder.WithInitialStreamReceiveWindow(1 << 21)
	builder.WithInitialConnectionReceiveWindow(1 << 22)
	builder.WithMaxStreamReceiveWindow(1 << 23)
	builder.WithMaxConnectionReceiveWindow(1 << 24)

	if builder.config.InitialStreamReceiveWindow != 1<<21 {
		t.Errorf("Expected InitialStreamReceiveWindow %d, got %d", 1<<21, builder.config.InitialStreamReceiveWindow)
	}

	if builder.config.InitialConnectionReceiveWindow != 1<<22 {
		t.Errorf("Expected InitialConnectionReceiveWindow %d, got %d", 1<<22, builder.config.InitialConnectionReceiveWindow)
	}

	if builder.config.MaxStreamReceiveWindow != 1<<23 {
		t.Errorf("Expected MaxStreamReceiveWindow %d, got %d", 1<<23, builder.config.MaxStreamReceiveWindow)
	}

	if builder.config.MaxConnectionReceiveWindow != 1<<24 {
		t.Errorf("Expected MaxConnectionReceiveWindow %d, got %d", 1<<24, builder.config.MaxConnectionReceiveWindow)
	}
}

func TestQUICConfigBuilderBuild(t *testing.T) {
	builder := NewQUICConfigBuilder().
		WithMaxIdleTimeout(60 * time.Second).
		WithKeepAlivePeriod(30 * time.Second).
		WithMaxIncomingStreams(500).
		WithDatagrams(false).
		With0RTT(true)

	config := builder.Build()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.MaxIdleTimeout != 60*time.Second {
		t.Errorf("Expected MaxIdleTimeout 60s, got %v", config.MaxIdleTimeout)
	}

	if config.KeepAlivePeriod != 30*time.Second {
		t.Errorf("Expected KeepAlivePeriod 30s, got %v", config.KeepAlivePeriod)
	}

	if config.MaxIncomingStreams != 500 {
		t.Errorf("Expected MaxIncomingStreams 500, got %d", config.MaxIncomingStreams)
	}

	if config.EnableDatagrams {
		t.Error("Expected EnableDatagrams to be false")
	}

	if !config.Allow0RTT {
		t.Error("Expected Allow0RTT to be true")
	}
}

func TestQUICConnectionPoolSetTLSConfig(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 10, 5*time.Minute)

	pool.SetTLSConfig(nil, nil)

	if pool.tlsClient != nil {
		t.Error("Expected tlsClient to be nil")
	}

	if pool.tlsServer != nil {
		t.Error("Expected tlsServer to be nil")
	}
}

func TestQUICConnectionPoolGetClosed(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 10, 5*time.Minute)
	pool.Close()

	_, err := pool.Get(nil, "test:1234")
	if err != ErrConnectionPoolClosed {
		t.Errorf("Expected ErrConnectionPoolClosed, got %v", err)
	}
}

func TestQUICStreamManagerOpenStreamClosed(t *testing.T) {
	manager := NewQUICStreamManager()
	manager.Close()

	_, _, err := manager.OpenStream(nil)
	if err != ErrConnectionClosed {
		t.Errorf("Expected ErrConnectionClosed, got %v", err)
	}
}

func TestQUICStreamManagerOpenStreamSyncClosed(t *testing.T) {
	manager := NewQUICStreamManager()
	manager.Close()

	_, _, err := manager.OpenStreamSync(nil, nil)
	if err != ErrConnectionClosed {
		t.Errorf("Expected ErrConnectionClosed, got %v", err)
	}
}

func TestQUICStreamManagerOpenUniStreamClosed(t *testing.T) {
	manager := NewQUICStreamManager()
	manager.Close()

	_, _, err := manager.OpenUniStream(nil)
	if err != ErrConnectionClosed {
		t.Errorf("Expected ErrConnectionClosed, got %v", err)
	}
}

func TestQUICConnectionPoolIsFull(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 2, 5*time.Minute)

	if pool.IsFull() {
		t.Error("Expected pool not to be full initially")
	}

	pool.stats.TotalConnections = 2

	if !pool.IsFull() {
		t.Error("Expected pool to be full when at max capacity")
	}
}

func TestQUICConnectionPoolAvailable(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 10, 5*time.Minute)

	available := pool.Available()
	if available != 10 {
		t.Errorf("Expected 10 available slots, got %d", available)
	}

	pool.stats.TotalConnections = 3

	available = pool.Available()
	if available != 7 {
		t.Errorf("Expected 7 available slots, got %d", available)
	}

	pool.stats.TotalConnections = 10

	available = pool.Available()
	if available != 0 {
		t.Errorf("Expected 0 available slots, got %d", available)
	}
}

func TestQUICConnectionPoolIsHealthy(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 10, 5*time.Minute)

	if !pool.IsHealthy() {
		t.Error("Expected pool to be healthy initially")
	}

	pool.Close()

	if pool.IsHealthy() {
		t.Error("Expected pool not to be healthy after close")
	}
}

func TestQUICConnectionPoolZeroMaxSize(t *testing.T) {
	config := DefaultQUICConnectionConfig()
	pool := NewQUICConnectionPool(config, 0, 5*time.Minute)

	if pool.IsFull() {
		t.Error("Expected pool not to be full when maxPoolSize is 0 (unlimited)")
	}

	available := pool.Available()
	if available != -1 {
		t.Errorf("Expected -1 for unlimited pool, got %d", available)
	}
}

func TestQUICStreamManagerConcurrentStats(t *testing.T) {
	manager := NewQUICStreamManager()
	defer manager.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			manager.RecordBytesSent(100)
		}()
		go func() {
			defer wg.Done()
			manager.RecordBytesReceived(200)
		}()
	}
	wg.Wait()

	stats := manager.Stats()
	if stats.BytesSent != 10000 {
		t.Errorf("Expected BytesSent 10000, got %d", stats.BytesSent)
	}
	if stats.BytesReceived != 20000 {
		t.Errorf("Expected BytesReceived 20000, got %d", stats.BytesReceived)
	}
}
