package mcpmoqt

import (
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	messagesSent      atomic.Int64
	messagesReceived  atomic.Int64
	messagesAcked     atomic.Int64
	messagesNacked    atomic.Int64
	messagesTimeout   atomic.Int64
	bytesSent         atomic.Int64
	bytesReceived     atomic.Int64
	errors            atomic.Int64
	retries           atomic.Int64
	heartbeatsSent    atomic.Int64
	heartbeatsFailed  atomic.Int64
	connectionsTotal  atomic.Int64
	connectionsActive atomic.Int64
	tracksTotal       atomic.Int64
	tracksActive      atomic.Int64
	dataTrackMessages atomic.Int64
	dataTrackBytes    atomic.Int64

	startTime time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{startTime: time.Now()}
}

func (m *Metrics) RecordMessageSent(bytes int) {
	m.messagesSent.Add(1)
	m.bytesSent.Add(int64(bytes))
}

func (m *Metrics) RecordMessageReceived(bytes int) {
	m.messagesReceived.Add(1)
	m.bytesReceived.Add(int64(bytes))
}

func (m *Metrics) RecordAck() {
	m.messagesAcked.Add(1)
}

func (m *Metrics) RecordNack() {
	m.messagesNacked.Add(1)
}

func (m *Metrics) RecordTimeout() {
	m.messagesTimeout.Add(1)
}

func (m *Metrics) RecordError() {
	m.errors.Add(1)
}

func (m *Metrics) RecordRetry() {
	m.retries.Add(1)
}

func (m *Metrics) RecordHeartbeatSent() {
	m.heartbeatsSent.Add(1)
}

func (m *Metrics) RecordHeartbeatFailed() {
	m.heartbeatsFailed.Add(1)
}

func (m *Metrics) RecordConnectionOpened() {
	m.connectionsTotal.Add(1)
	m.connectionsActive.Add(1)
}

func (m *Metrics) RecordConnectionClosed() {
	m.connectionsActive.Add(-1)
}

func (m *Metrics) RecordTrackOpened() {
	m.tracksTotal.Add(1)
	m.tracksActive.Add(1)
}

func (m *Metrics) RecordTrackClosed() {
	m.tracksActive.Add(-1)
}

func (m *Metrics) RecordDataTrackMessage(bytes int) {
	m.dataTrackMessages.Add(1)
	m.dataTrackBytes.Add(int64(bytes))
}

func (m *Metrics) MessagesSent() int64 {
	return m.messagesSent.Load()
}

func (m *Metrics) MessagesReceived() int64 {
	return m.messagesReceived.Load()
}

func (m *Metrics) MessagesAcked() int64 {
	return m.messagesAcked.Load()
}

func (m *Metrics) MessagesNacked() int64 {
	return m.messagesNacked.Load()
}

func (m *Metrics) MessagesTimeout() int64 {
	return m.messagesTimeout.Load()
}

func (m *Metrics) BytesSent() int64 {
	return m.bytesSent.Load()
}

func (m *Metrics) BytesReceived() int64 {
	return m.bytesReceived.Load()
}

func (m *Metrics) Errors() int64 {
	return m.errors.Load()
}

func (m *Metrics) Retries() int64 {
	return m.retries.Load()
}

func (m *Metrics) HeartbeatsSent() int64 {
	return m.heartbeatsSent.Load()
}

func (m *Metrics) HeartbeatsFailed() int64 {
	return m.heartbeatsFailed.Load()
}

func (m *Metrics) ConnectionsTotal() int64 {
	return m.connectionsTotal.Load()
}

func (m *Metrics) ConnectionsActive() int64 {
	return m.connectionsActive.Load()
}

func (m *Metrics) TracksTotal() int64 {
	return m.tracksTotal.Load()
}

func (m *Metrics) TracksActive() int64 {
	return m.tracksActive.Load()
}

func (m *Metrics) DataTrackMessages() int64 {
	return m.dataTrackMessages.Load()
}

func (m *Metrics) DataTrackBytes() int64 {
	return m.dataTrackBytes.Load()
}

func (m *Metrics) Uptime() time.Duration {
	return time.Since(m.startTime)
}

type MetricsSnapshot struct {
	MessagesSent      int64         `json:"messages_sent"`
	MessagesReceived  int64         `json:"messages_received"`
	MessagesAcked     int64         `json:"messages_acked"`
	MessagesNacked    int64         `json:"messages_nacked"`
	MessagesTimeout   int64         `json:"messages_timeout"`
	BytesSent         int64         `json:"bytes_sent"`
	BytesReceived     int64         `json:"bytes_received"`
	Errors            int64         `json:"errors"`
	Retries           int64         `json:"retries"`
	HeartbeatsSent    int64         `json:"heartbeats_sent"`
	HeartbeatsFailed  int64         `json:"heartbeats_failed"`
	ConnectionsTotal  int64         `json:"connections_total"`
	ConnectionsActive int64         `json:"connections_active"`
	TracksTotal       int64         `json:"tracks_total"`
	TracksActive      int64         `json:"tracks_active"`
	DataTrackMessages int64         `json:"data_track_messages"`
	DataTrackBytes    int64         `json:"data_track_bytes"`
	Uptime            time.Duration `json:"uptime"`
	AckRate           float64       `json:"ack_rate"`
	ErrorRate         float64       `json:"error_rate"`
	ThroughputIn      float64       `json:"throughput_in_bytes_per_sec"`
	ThroughputOut     float64       `json:"throughput_out_bytes_per_sec"`
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	sent := m.MessagesSent()
	received := m.MessagesReceived()
	acked := m.MessagesAcked()
	errors := m.Errors()
	uptime := m.Uptime().Seconds()

	var ackRate float64
	if sent > 0 {
		ackRate = float64(acked) / float64(sent)
	}

	var errorRate float64
	if sent+received > 0 {
		errorRate = float64(errors) / float64(sent+received)
	}

	var throughputIn, throughputOut float64
	if uptime > 0 {
		throughputIn = float64(m.BytesReceived()) / uptime
		throughputOut = float64(m.BytesSent()) / uptime
	}

	return MetricsSnapshot{
		MessagesSent:      sent,
		MessagesReceived:  received,
		MessagesAcked:     acked,
		MessagesNacked:    m.MessagesNacked(),
		MessagesTimeout:   m.MessagesTimeout(),
		BytesSent:         m.BytesSent(),
		BytesReceived:     m.BytesReceived(),
		Errors:            errors,
		Retries:           m.Retries(),
		HeartbeatsSent:    m.HeartbeatsSent(),
		HeartbeatsFailed:  m.HeartbeatsFailed(),
		ConnectionsTotal:  m.ConnectionsTotal(),
		ConnectionsActive: m.ConnectionsActive(),
		TracksTotal:       m.TracksTotal(),
		TracksActive:      m.TracksActive(),
		DataTrackMessages: m.DataTrackMessages(),
		DataTrackBytes:    m.DataTrackBytes(),
		Uptime:            m.Uptime(),
		AckRate:           ackRate,
		ErrorRate:         errorRate,
		ThroughputIn:      throughputIn,
		ThroughputOut:     throughputOut,
	}
}

func (m *Metrics) Reset() {
	m.messagesSent.Store(0)
	m.messagesReceived.Store(0)
	m.messagesAcked.Store(0)
	m.messagesNacked.Store(0)
	m.messagesTimeout.Store(0)
	m.bytesSent.Store(0)
	m.bytesReceived.Store(0)
	m.errors.Store(0)
	m.retries.Store(0)
	m.heartbeatsSent.Store(0)
	m.heartbeatsFailed.Store(0)
	m.connectionsTotal.Store(0)
	m.connectionsActive.Store(0)
	m.tracksTotal.Store(0)
	m.tracksActive.Store(0)
	m.dataTrackMessages.Store(0)
	m.dataTrackBytes.Store(0)
	m.startTime = time.Now()
}

var defaultMetrics = NewMetrics()
var defaultMetricsMu sync.RWMutex

func DefaultMetrics() *Metrics {
	defaultMetricsMu.RLock()
	defer defaultMetricsMu.RUnlock()
	return defaultMetrics
}
