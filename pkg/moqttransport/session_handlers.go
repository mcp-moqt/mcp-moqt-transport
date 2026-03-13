package mcpmoqt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mengelbart/moqtransport"
)

// Draft: draft-jennings-mcp-over-moqt-00 §2.2.1
// Control tracks: namespace "mcp/<session-id>/control", tracks:
//   - client-to-server
//   - server-to-client
//
// This implementation maps MCP JSON-RPC to MOQT objects on these tracks.
// NOTE: This implementation targets moqtransport (draft-11, moq-00).
// It is not wire-compatible with draft-16 stream/datagram encodings.
// subscribeHandler accepts subscriptions to the local send track and exposes a Publisher via slot.
type subscribeHandler struct {
	sessionID string
	sendTrack string
	sendSlot  *publisherSlot
}

func (h *subscribeHandler) HandleSubscribe(rw *moqtransport.SubscribeResponseWriter, msg *moqtransport.SubscribeMessage) {
	// Expected: namespace (mcp, <session-id>, control), track == our send track.
	if len(msg.Namespace) != 3 || msg.Namespace[0] != controlNS0 || msg.Namespace[2] != controlNS2 {
		rw.Reject(moqtransport.ErrorCodeSubscribeTrackDoesNotExist, "unknown namespace")
		return
	}
	// If h.sessionID is empty, accept first matching subscription and bind it.
	// This avoids ordering races between discovery and the peer subscribing.
	if h.sessionID != "" && msg.Namespace[1] != h.sessionID {
		rw.Reject(moqtransport.ErrorCodeSubscribeTrackDoesNotExist, "unknown session")
		return
	}
	if msg.Track != h.sendTrack {
		rw.Reject(moqtransport.ErrorCodeSubscribeTrackDoesNotExist, "unknown track")
		return
	}
	if err := rw.Accept(moqtransport.WithLargestLocation(&moqtransport.Location{Group: 0, Object: 0})); err != nil {
		return
	}
	if h.sessionID == "" {
		h.sessionID = msg.Namespace[1]
	}
	h.sendSlot.set(rw)
}

// Draft: draft-jennings-mcp-over-moqt-00 §2.2.x (discovery)
// Discovery is implemented via FETCH on "mcp/discovery/sessions".
// Response contains session_id and available control tracks.
// TODO (Draft: draft-jennings-mcp-over-moqt-00 §2.3/§2.4):
// Reserve tracks for resources/tools/notifications. Not implemented in v0.2.0.
// discoveryHandler implements the server-side FETCH "mcp/discovery" "sessions".
type discoveryHandler struct {
	sessionID string

	mu sync.Mutex
}

func (h *discoveryHandler) Handle(rw moqtransport.ResponseWriter, msg *moqtransport.Message) {
	if msg.Method != moqtransport.MessageFetch {
		rw.Reject(0, "unsupported")
		return
	}
	if len(msg.Namespace) == 2 && msg.Namespace[0] == "mcp" && msg.Namespace[1] == "discovery" && msg.Track == "sessions" {
		h.handleDiscoveryFetch(rw)
		return
	}
	rw.Reject(0, "unknown fetch track")
}

func (h *discoveryHandler) handleDiscoveryFetch(rw moqtransport.ResponseWriter) {
	fetchPublisher, ok := rw.(moqtransport.FetchPublisher)
	if !ok {
		rw.Reject(0, "internal error: not a fetch publisher")
		return
	}
	if err := rw.Accept(); err != nil {
		return
	}

	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]any{
			"session_id": h.sessionID,
			"server_info": map[string]any{
				"name":             "mcp-moqt-transport",
				"version":          "0.1.2",
				"protocol_version": "2025-06-18",
			},
			"available_tracks": map[string]any{
				"control": map[string]string{
					"client_to_server": fmt.Sprintf("mcp/%s/control/%s", h.sessionID, trackClientToServer),
					"server_to_client": fmt.Sprintf("mcp/%s/control/%s", h.sessionID, trackServerToClient),
				},
			},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	stream, err := fetchPublisher.FetchStream()
	if err != nil {
		return
	}
	defer stream.Close()
	_, _ = stream.WriteObject(0, 0, 0, 1, data)
}

// noOpHandler rejects everything (used on the client, which doesn't serve discovery).
type noOpHandler struct{}

func (noOpHandler) Handle(rw moqtransport.ResponseWriter, _ *moqtransport.Message) {
	_ = rw.Reject(0, "unsupported")
}

func runSession(ctx context.Context, s *moqtransport.Session, conn moqtransport.Connection) error {
	// moqtransport.Session.Run is async-ish (it returns after handshake setup),
	// but we still call it synchronously so any immediate error propagates.
	if err := s.Run(conn); err != nil {
		return err
	}
	// Close the session when ctx is done.
	go func() {
		<-ctx.Done()
		_ = s.Close()
	}()
	return nil
}
