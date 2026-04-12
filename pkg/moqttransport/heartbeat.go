package mcpmoqt

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type HeartbeatConfig struct {
	Interval    time.Duration
	Timeout     time.Duration
	MaxFailures int
	OnHeartbeat func() error
	OnTimeout   func()
	OnReconnect func()
}

func DefaultHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		MaxFailures: 3,
	}
}

type Heartbeat struct {
	config   HeartbeatConfig
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	running  atomic.Bool
	failures atomic.Int32
	lastPong atomic.Int64
}

func NewHeartbeat(config HeartbeatConfig) *Heartbeat {
	return &Heartbeat{config: config}
}

func (h *Heartbeat) Start(ctx context.Context) error {
	if !h.running.CompareAndSwap(false, true) {
		return nil
	}

	h.ctx, h.cancel = context.WithCancel(ctx)
	h.failures.Store(0)
	h.lastPong.Store(time.Now().UnixNano())

	go h.loop()

	return nil
}

func (h *Heartbeat) loop() {
	ticker := time.NewTicker(h.config.Interval)
	defer ticker.Stop()
	defer h.running.Store(false)

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.sendHeartbeat()
		}
	}
}

func (h *Heartbeat) sendHeartbeat() {
	if h.config.OnHeartbeat == nil {
		return
	}

	ctx, cancel := context.WithTimeout(h.ctx, h.config.Timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- h.config.OnHeartbeat()
	}()

	select {
	case <-h.ctx.Done():
		<-done
		h.handleFailure(ErrTimeout)
	case <-ctx.Done():
		<-done
		h.handleFailure(ErrTimeout)
	case err := <-done:
		if err != nil {
			h.handleFailure(err)
		} else {
			h.handleSuccess()
		}
	}
}

func (h *Heartbeat) handleSuccess() {
	h.failures.Store(0)
	h.lastPong.Store(time.Now().UnixNano())
}

func (h *Heartbeat) handleFailure(err error) {
	failures := h.failures.Add(1)

	if h.config.OnTimeout != nil {
		h.config.OnTimeout()
	}

	if failures >= int32(h.config.MaxFailures) {
		if h.config.OnReconnect != nil {
			h.config.OnReconnect()
		}
		h.failures.Store(0)
	}
}

func (h *Heartbeat) Stop() {
	if h.running.CompareAndSwap(true, false) {
		if h.cancel != nil {
			h.cancel()
		}
	}
}

func (h *Heartbeat) RecordPong() {
	h.lastPong.Store(time.Now().UnixNano())
	h.failures.Store(0)
}

func (h *Heartbeat) LastPong() time.Time {
	return time.Unix(0, h.lastPong.Load())
}

func (h *Heartbeat) Failures() int {
	return int(h.failures.Load())
}

func (h *Heartbeat) IsRunning() bool {
	return h.running.Load()
}

func (h *Heartbeat) IsHealthy() bool {
	if !h.running.Load() {
		return false
	}

	lastPong := h.LastPong()
	elapsed := time.Since(lastPong)
	return elapsed < h.config.Interval*time.Duration(h.config.MaxFailures)
}
