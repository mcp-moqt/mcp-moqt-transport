package mcpmoqt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mengelbart/moqtransport"
)

type subscribeHandler struct {
	sessionID string
	sendTrack string
	sendSlot  *PublisherSlot
}

func (h *subscribeHandler) HandleSubscribe(rw *moqtransport.SubscribeResponseWriter, msg *moqtransport.SubscribeMessage) {
	if len(msg.Namespace) != 3 || msg.Namespace[0] != controlNS0 || msg.Namespace[2] != controlNS2 {
		rw.Reject(moqtransport.ErrorCodeSubscribeTrackDoesNotExist, "unknown namespace")
		return
	}
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
	h.sendSlot.Set(rw)
}

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
				"version":          "0.7.1",
				"protocol_version": "2025-06-18",
			},
			"available_tracks": map[string]any{
				"control": map[string]string{
					"client_to_server": fmt.Sprintf("mcp/%s/control/%s", h.sessionID, trackClientToServer),
					"server_to_client": fmt.Sprintf("mcp/%s/control/%s", h.sessionID, trackServerToClient),
				},
				"data": map[string]any{
					"resources": map[string]string{
						"namespace":   fmt.Sprintf("mcp/%s/resources", h.sessionID),
						"description": "Track for resource data transfer",
					},
					"tools": map[string]string{
						"namespace":   fmt.Sprintf("mcp/%s/tools", h.sessionID),
						"description": "Track for tool invocation and results",
					},
					"notifications": map[string]string{
						"namespace":   fmt.Sprintf("mcp/%s/notifications", h.sessionID),
						"description": "Track for notification messages",
					},
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

type noOpHandler struct{}

func (noOpHandler) Handle(rw moqtransport.ResponseWriter, _ *moqtransport.Message) {
	_ = rw.Reject(0, "unsupported")
}

func runSession(ctx context.Context, s *moqtransport.Session, conn moqtransport.Connection) error {
	if err := s.Run(conn); err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = s.Close()
	}()
	return nil
}
