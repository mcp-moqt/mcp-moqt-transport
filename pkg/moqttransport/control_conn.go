package mcpmoqt

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/mengelbart/moqtransport"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	controlNS0 = "mcp"
	controlNS2 = "control"

	trackClientToServer = "client-to-server"
	trackServerToClient = "server-to-client"

	heartbeatMethod = "notifications/moqt/heartbeat"
)

func controlNamespace(sessionID string) []string {
	return []string{controlNS0, sessionID, controlNS2}
}

type controlConn struct {
	conn      moqtransport.Connection
	sessionID string

	recv  *moqtransport.RemoteTrack
	send  *PublisherSlot
	write *TrackWriter

	readCtx    context.Context
	cancelRead context.CancelFunc

	closed    bool
	mu        sync.Mutex
	closeOnce sync.Once
	done      chan struct{}

	ackTracker *AckTracker
	heartbeat  *Heartbeat
	retry      *Retry
	metrics    *Metrics

	OnClose func()
}

func newControlConn(conn moqtransport.Connection, sessionID string, recv *moqtransport.RemoteTrack, send *PublisherSlot, onClose ...func()) *controlConn {
	readCtx, cancel := context.WithCancel(conn.Context())
	var closeFunc func()
	if len(onClose) > 0 {
		closeFunc = onClose[0]
	}

	ackTracker := NewAckTracker(DefaultAckConfig())
	retry := NewRetry(DefaultRetryConfig())
	metrics := DefaultMetrics()
	metrics.RecordConnectionOpened()

	c := &controlConn{
		conn:       conn,
		sessionID:  sessionID,
		recv:       recv,
		send:       send,
		write:      NewTrackWriter(send),
		readCtx:    readCtx,
		cancelRead: cancel,
		done:       make(chan struct{}),
		ackTracker: ackTracker,
		retry:      retry,
		metrics:    metrics,
		OnClose:    closeFunc,
	}

	heartbeatConfig := DefaultHeartbeatConfig()
	heartbeatConfig.OnHeartbeat = func() error {
		c.metrics.RecordHeartbeatSent()
		payload := []byte(`{"jsonrpc":"2.0","method":"` + heartbeatMethod + `"}`)
		return c.write.Write(c.readCtx, payload)
	}
	heartbeatConfig.OnTimeout = func() {
		c.metrics.RecordHeartbeatFailed()
		if c.heartbeat != nil && c.heartbeat.Failures() >= heartbeatConfig.MaxFailures {
			_ = c.Close()
		}
	}
	c.heartbeat = NewHeartbeat(heartbeatConfig)
	_ = c.heartbeat.Start(readCtx)

	return c
}

func (c *controlConn) SessionID() string { return c.sessionID }

// MetricsSnapshot returns a point-in-time metrics snapshot for this connection.
func (c *controlConn) MetricsSnapshot() MetricsSnapshot {
	return c.metrics.Snapshot()
}

func (c *controlConn) Read(ctx context.Context) (jsonrpc.Message, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.done:
			return nil, io.EOF
		default:
		}

		combined, cancel := context.WithCancel(ctx)
		go func() {
			select {
			case <-c.readCtx.Done():
				cancel()
			case <-combined.Done():
			}
		}()

		obj, err := c.recv.ReadObject(combined)
		cancel()
		if err != nil {
			return nil, err
		}

		msg, err := jsonrpc.DecodeMessage(obj.Payload)
		if err != nil {
			c.metrics.RecordError()
			return nil, err
		}

		c.metrics.RecordMessageReceived(len(obj.Payload))
		if c.heartbeat != nil {
			c.heartbeat.RecordPong()
		}
		c.handleInboundAck(msg)

		if req, ok := msg.(*jsonrpc.Request); ok && req.Method == heartbeatMethod {
			// Internal keepalive; do not surface to MCP layer.
			continue
		}
		return msg, nil
	}
}

func (c *controlConn) Write(ctx context.Context, msg jsonrpc.Message) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return mcp.ErrConnectionClosed
	}
	c.mu.Unlock()

	data, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return err
	}

	msgID := extractMessageID(msg)
	attempt := 0
	err = c.retry.Do(ctx, func() error {
		if attempt > 0 {
			c.metrics.RecordRetry()
		}
		attempt++

		if err := c.write.Write(ctx, data); err != nil {
			c.metrics.RecordError()
			return err
		}

		c.metrics.RecordMessageSent(len(data))
		if msgID != "" {
			c.trackAckAsync(msgID)
		}
		return nil
	})

	if err != nil {
		c.metrics.RecordError()
	}
	return err
}

func (c *controlConn) trackAckAsync(msgID string) {
	ackCh := c.ackTracker.Track(msgID)
	go func() {
		select {
		case ack, ok := <-ackCh:
			if !ok || ack == nil {
				return
			}
			switch ack.Status {
			case AckStatusAcked:
				c.metrics.RecordAck()
			case AckStatusNacked:
				c.metrics.RecordNack()
			case AckStatusTimeout:
				c.metrics.RecordTimeout()
			}
		case <-c.done:
		}
	}()
}

func (c *controlConn) handleInboundAck(msg jsonrpc.Message) {
	switch m := msg.(type) {
	case *jsonrpc.Response:
		if !m.ID.IsValid() {
			return
		}
		id := fmt.Sprint(m.ID.Raw())
		if m.Error != nil {
			c.ackTracker.Nack(id, m.Error)
			c.metrics.RecordNack()
			return
		}
		c.ackTracker.Ack(id)
		c.metrics.RecordAck()
	}
}

func extractMessageID(msg jsonrpc.Message) string {
	switch m := msg.(type) {
	case *jsonrpc.Request:
		if m.ID.IsValid() {
			return fmt.Sprint(m.ID.Raw())
		}
	case *jsonrpc.Response:
		if m.ID.IsValid() {
			return fmt.Sprint(m.ID.Raw())
		}
	}
	return ""
}

func (c *controlConn) Close() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		close(c.done)
		c.mu.Unlock()

		if c.heartbeat != nil {
			c.heartbeat.Stop()
		}
		if c.ackTracker != nil {
			c.ackTracker.Close()
		}
		c.metrics.RecordConnectionClosed()

		if c.cancelRead != nil {
			c.cancelRead()
		}
		if c.recv != nil {
			_ = c.recv.Close()
		}
		if c.write != nil {
			_ = c.write.Close()
		}
		if c.conn != nil {
			_ = c.conn.CloseWithError(0, "")
		}
		if c.OnClose != nil {
			c.OnClose()
		}
	})
	return nil
}
