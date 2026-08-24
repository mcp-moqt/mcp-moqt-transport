package mcpmoqt

import (
	"sync"
	"sync/atomic"
	"time"
)

type AckStatus int

const (
	AckStatusPending AckStatus = iota
	AckStatusAcked
	AckStatusNacked
	AckStatusTimeout
)

type AckConfig struct {
	Timeout       time.Duration
	RetryCount    int
	RetryInterval time.Duration
}

func DefaultAckConfig() AckConfig {
	return AckConfig{
		Timeout:       5 * time.Second,
		RetryCount:    3,
		RetryInterval: 100 * time.Millisecond,
	}
}

type AckMessage struct {
	ID        string
	Status    AckStatus
	Timestamp time.Time
	Error     error
}

type AckTracker struct {
	mu      sync.RWMutex
	pending map[string]*pendingAck
	timeout time.Duration
	closed  atomic.Bool
}

type pendingAck struct {
	id        string
	timestamp time.Time
	ackCh     chan *AckMessage
	timeout   *time.Timer
}

func NewAckTracker(config AckConfig) *AckTracker {
	return &AckTracker{
		pending: make(map[string]*pendingAck),
		timeout: config.Timeout,
	}
}

func (t *AckTracker) Track(messageID string) <-chan *AckMessage {
	ackCh := make(chan *AckMessage, 1)

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed.Load() {
		ackCh <- &AckMessage{
			ID:        messageID,
			Status:    AckStatusNacked,
			Timestamp: time.Now(),
			Error:     ErrConnectionClosed,
		}
		return ackCh
	}

	p := &pendingAck{
		id:        messageID,
		timestamp: time.Now(),
		ackCh:     ackCh,
	}

	p.timeout = time.AfterFunc(t.timeout, func() {
		t.timeoutAck(messageID)
	})

	t.pending[messageID] = p

	return ackCh
}

func (t *AckTracker) Ack(messageID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if p, ok := t.pending[messageID]; ok {
		p.timeout.Stop()
		p.ackCh <- &AckMessage{
			ID:        messageID,
			Status:    AckStatusAcked,
			Timestamp: time.Now(),
		}
		close(p.ackCh)
		delete(t.pending, messageID)
	}
}

func (t *AckTracker) Nack(messageID string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if p, ok := t.pending[messageID]; ok {
		p.timeout.Stop()
		p.ackCh <- &AckMessage{
			ID:        messageID,
			Status:    AckStatusNacked,
			Timestamp: time.Now(),
			Error:     err,
		}
		close(p.ackCh)
		delete(t.pending, messageID)
	}
}

func (t *AckTracker) timeoutAck(messageID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if p, ok := t.pending[messageID]; ok {
		p.ackCh <- &AckMessage{
			ID:        messageID,
			Status:    AckStatusTimeout,
			Timestamp: time.Now(),
			Error:     ErrTimeout,
		}
		close(p.ackCh)
		delete(t.pending, messageID)
	}
}

func (t *AckTracker) Close() {
	if t.closed.CompareAndSwap(false, true) {
		t.mu.Lock()
		defer t.mu.Unlock()

		for id, p := range t.pending {
			p.timeout.Stop()
			p.ackCh <- &AckMessage{
				ID:        id,
				Status:    AckStatusNacked,
				Timestamp: time.Now(),
				Error:     ErrConnectionClosed,
			}
			close(p.ackCh)
		}
		t.pending = make(map[string]*pendingAck)
	}
}

func (t *AckTracker) PendingCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.pending)
}

func (t *AckTracker) IsClosed() bool {
	return t.closed.Load()
}
