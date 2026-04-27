package mcpmoqt

import (
	"context"
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

	// Reliability features
	ackTracker *AckTracker
	heartbeat  *Heartbeat
	retry      *Retry
	metrics    *Metrics

	// For server-side connection management
	OnClose func()
}

func newControlConn(conn moqtransport.Connection, sessionID string, recv *moqtransport.RemoteTrack, send *PublisherSlot, onClose ...func()) *controlConn {
	readCtx, cancel := context.WithCancel(conn.Context())
	var closeFunc func()
	if len(onClose) > 0 {
		closeFunc = onClose[0]
	}

	// Initialize reliability features
	ackTracker := NewAckTracker()
	heartbeat := NewHeartbeat(DefaultHeartbeatConfig())
	retry := NewRetry(DefaultRetryConfig())
	metrics := DefaultMetrics()

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
		heartbeat:  heartbeat,
		retry:      retry,
		metrics:    metrics,
		OnClose:    closeFunc,
	}

	// Start heartbeat
	heartbeat.Start(readCtx, func() error {
		// Send heartbeat message
		// This would typically be a ping message to the server
		return nil
	}, func() {
		// Handle heartbeat timeout
		c.Close()
	}, func() {
		// Handle heartbeat reconnect
		// This would typically reconnect to the server
	})

	return c
}

func (c *controlConn) SessionID() string { return c.sessionID }

func (c *controlConn) Read(ctx context.Context) (jsonrpc.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, io.EOF
	default:
	}

	combined, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-c.readCtx.Done():
			cancel()
		case <-combined.Done():
		}
	}()

	obj, err := c.recv.ReadObject(combined)
	if err != nil {
		return nil, err
	}
	return jsonrpc.DecodeMessage(obj.Payload)
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

	// Use retry mechanism for reliable delivery
	err = c.retry.Do(ctx, func(ctx context.Context) error {
		// Write the message
		if err := c.write.Write(ctx, data); err != nil {
			c.metrics.RecordError()
			return err
		}

		// Track message for acknowledgment
		// Note: This is a simplified implementation
		// In a real-world scenario, you would need to extract a message ID
		// from the JSON-RPC message and track it
		// c.ackTracker.TrackMessage(messageID)

		c.metrics.RecordMessageSent(1)
		return nil
	})

	if err != nil {
		c.metrics.RecordError()
	}

	return err
}

func (c *controlConn) Close() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		close(c.done)
		c.mu.Unlock()

		// Stop reliability features
		if c.heartbeat != nil {
			c.heartbeat.Stop()
		}

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
		// Call OnClose callback if set
		if c.OnClose != nil {
			c.OnClose()
		}
	})
	return nil
}
