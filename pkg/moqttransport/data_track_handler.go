package mcpmoqt

import (
	"encoding/json"
	"sync"

	"github.com/mengelbart/moqtransport"
)

type DataTrackHandler struct {
	sessionID string
	trackType DataTrackType
	
	mu         sync.Mutex
	publishers map[string]moqtransport.Publisher
	subscribers map[string]*dataTrackSubscriber
}

type dataTrackSubscriber struct {
	trackName string
	handler   func(*DataTrackMessage) error
}

func NewDataTrackHandler(sessionID string, trackType DataTrackType) *DataTrackHandler {
	return &DataTrackHandler{
		sessionID:   sessionID,
		trackType:   trackType,
		publishers:  make(map[string]moqtransport.Publisher),
		subscribers: make(map[string]*dataTrackSubscriber),
	}
}

func (h *DataTrackHandler) HandleSubscribe(rw *moqtransport.SubscribeResponseWriter, msg *moqtransport.SubscribeMessage) {
	expectedNS := dataTrackNamespace(h.sessionID, h.trackType)
	
	if len(msg.Namespace) != len(expectedNS) {
		rw.Reject(moqtransport.ErrorCodeSubscribeTrackDoesNotExist, "invalid namespace length")
		return
	}
	
	for i, ns := range expectedNS {
		if msg.Namespace[i] != ns {
			rw.Reject(moqtransport.ErrorCodeSubscribeTrackDoesNotExist, "namespace mismatch")
			return
		}
	}
	
	if err := rw.Accept(moqtransport.WithLargestLocation(&moqtransport.Location{Group: 0, Object: 0})); err != nil {
		return
	}
	
	h.mu.Lock()
	h.publishers[msg.Track] = rw
	h.mu.Unlock()
}

func (h *DataTrackHandler) Publish(trackName string, message *DataTrackMessage) error {
	h.mu.Lock()
	pub, ok := h.publishers[trackName]
	h.mu.Unlock()
	
	if !ok {
		return ErrNoPublisher
	}
	
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	sg, err := pub.OpenSubgroup(0, 0, 0)
	if err != nil {
		return err
	}
	
	_, err = sg.WriteObject(0, data)
	if err != nil {
		_ = sg.Close()
		return err
	}
	
	return sg.Close()
}

func (h *DataTrackHandler) Subscribe(trackName string, handler func(*DataTrackMessage) error) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.subscribers[trackName] = &dataTrackSubscriber{
		trackName: trackName,
		handler:   handler,
	}
	
	return nil
}

func (h *DataTrackHandler) HandleData(trackName string, data []byte) error {
	h.mu.Lock()
	subscriber, ok := h.subscribers[trackName]
	h.mu.Unlock()
	
	if !ok {
		return ErrNoSubscriber
	}
	
	var message DataTrackMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}
	
	return subscriber.handler(&message)
}

func (h *DataTrackHandler) GetTrackInfo() []DataTrackInfo {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	var tracks []DataTrackInfo
	for trackName := range h.publishers {
		tracks = append(tracks, DataTrackInfo{
			Type:      h.trackType,
			TrackName: trackName,
			Namespace: dataTrackNamespace(h.sessionID, h.trackType),
		})
	}
	
	return tracks
}
