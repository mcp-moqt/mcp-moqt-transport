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
}

func newControlConn(conn moqtransport.Connection, sessionID string, recv *moqtransport.RemoteTrack, send *PublisherSlot) *controlConn {
	readCtx, cancel := context.WithCancel(conn.Context())
	return &controlConn{
		conn:       conn,
		sessionID:  sessionID,
		recv:       recv,
		send:       send,
		write:      NewTrackWriter(send),
		readCtx:    readCtx,
		cancelRead: cancel,
		done:       make(chan struct{}),
	}
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

	return c.write.Write(ctx, data)
}

func (c *controlConn) Close() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		close(c.done)
		c.mu.Unlock()
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
	})
	return nil
}
