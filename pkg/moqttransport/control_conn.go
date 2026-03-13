package mcpmoqt

import (
	"context"
	"io"
	"sync"
	"sync/atomic"

	"github.com/mengelbart/moqtransport"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Draft: draft-jennings-mcp-over-moqt-00 §2.2.1
// Control tracks: namespace "mcp/<session-id>/control", tracks:
//   - client-to-server: for messages from client to server
//   - server-to-client: for messages from server to client
//
// This implementation maps MCP JSON-RPC to MOQT objects on these tracks.
// NOTE: This implementation targets moqtransport (draft-11, moq-00).
// It is not wire-compatible with draft-16 stream/datagram encodings.
const (
	// controlNS0 is the first part of the control track namespace.
	controlNS0 = "mcp"
	// controlNS2 is the third part of the control track namespace.
	controlNS2 = "control"

	// trackClientToServer is the track name for client-to-server messages.
	trackClientToServer = "client-to-server"
	// trackServerToClient is the track name for server-to-client messages.
	trackServerToClient = "server-to-client"
)

// controlNamespace constructs the control track namespace for a given session ID.
// Format: ["mcp", sessionID, "control"]
func controlNamespace(sessionID string) []string {
	return []string{controlNS0, sessionID, controlNS2}
}

// publisherSlot is a thread-safe slot for a MOQT publisher.
// It allows setting the publisher once and retrieving it with context support.
type publisherSlot struct {
	mu    sync.Mutex
	pub   moqtransport.Publisher
	ready chan struct{}
}

// newPublisherSlot creates a new publisher slot.
func newPublisherSlot() *publisherSlot {
	return &publisherSlot{ready: make(chan struct{})}
}

// set sets the publisher in the slot if it's not already set.
func (s *publisherSlot) set(pub moqtransport.Publisher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pub != nil {
		return
	}
	s.pub = pub
	close(s.ready)
}

// get returns the publisher from the slot, waiting until it's ready or the context is canceled.
func (s *publisherSlot) get(ctx context.Context) (moqtransport.Publisher, error) {
	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	case <-s.ready:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pub, nil
}

// controlConn implements mcp.Connection over MOQT control tracks.
// It handles reading and writing MCP JSON-RPC messages over MOQT.
type controlConn struct {
	// conn is the underlying MOQT connection.
	conn moqtransport.Connection
	// sessionID is the MCP session ID.
	sessionID string

	// recv is the remote track for receiving messages.
	recv *moqtransport.RemoteTrack
	// send is the publisher slot for sending messages.
	send *publisherSlot

	// nextGroup is the next group ID to use for MOQT objects.
	nextGroup atomic.Uint64

	// readCtx is the context for read operations.
	readCtx context.Context
	// cancelRead cancels the read context.
	cancelRead context.CancelFunc

	// closed indicates if the connection is closed.
	closed bool
	// mu protects the closed field.
	mu sync.Mutex
	// closeOnce ensures Close is called only once.
	closeOnce sync.Once
	// done is closed when the connection is closed.
	done chan struct{}
}

// newControlConn creates a new control connection.
func newControlConn(conn moqtransport.Connection, sessionID string, recv *moqtransport.RemoteTrack, send *publisherSlot) *controlConn {
	readCtx, cancel := context.WithCancel(conn.Context())
	return &controlConn{
		conn:       conn,
		sessionID:  sessionID,
		recv:       recv,
		send:       send,
		readCtx:    readCtx,
		cancelRead: cancel,
		done:       make(chan struct{}),
	}
}

// SessionID returns the MCP session ID.
func (c *controlConn) SessionID() string { return c.sessionID }

// Read reads a JSON-RPC message from the control track.
// It blocks until a message is received, the context is canceled, or the connection is closed.
func (c *controlConn) Read(ctx context.Context) (jsonrpc.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, io.EOF
	default:
	}

	// Ensure that Close unblocks a Read that is waiting in ReadObject.
	combined, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-c.readCtx.Done():
			cancel()
		case <-combined.Done():
		}
	}()

	// Read MOQT object from the remote track
	obj, err := c.recv.ReadObject(combined)
	if err != nil {
		return nil, err
	}
	// Decode JSON-RPC message from the object payload
	return jsonrpc.DecodeMessage(obj.Payload)
}

// Write writes a JSON-RPC message to the control track.
// It blocks until the message is sent or the context is canceled.
func (c *controlConn) Write(ctx context.Context, msg jsonrpc.Message) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return mcp.ErrConnectionClosed
	}
	c.mu.Unlock()

	// Get the publisher from the slot
	pub, err := c.send.get(ctx)
	if err != nil {
		return err
	}
	// Encode the JSON-RPC message
	data, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return err
	}

	// Create a new subgroup for the message
	groupID := c.nextGroup.Add(1) - 1
	sg, err := pub.OpenSubgroup(groupID, 0, 0)
	if err != nil {
		return err
	}
	// Write the message as a MOQT object
	if _, err := sg.WriteObject(0, data); err != nil {
		_ = sg.Close()
		return err
	}
	// Close the subgroup
	return sg.Close()
}

// Close closes the control connection.
// It cancels all pending operations and closes the underlying MOQT connection.
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
		if c.conn != nil {
			_ = c.conn.CloseWithError(0, "")
		}
	})
	return nil
}
