package mcpmoqt

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/mengelbart/moqtransport"
)

type DataTrackHandler struct {
	sessionID string
	trackType DataTrackType
	namespace []string

	mu          sync.RWMutex
	publishers  map[string]*dataTrackPublisher
	subscribers map[string]*dataTrackSubscriber
	closed      atomic.Bool
	done        chan struct{}
}

type dataTrackPublisher struct {
	writer *TrackWriter
}

type dataTrackSubscriber struct {
	reader  *TrackReader
	handler func(*DataTrackMessage) error
	ctx     context.Context
	cancel  context.CancelFunc
}

type DataTrackConfig struct {
	SessionID string
	TrackType DataTrackType
}

func NewDataTrackHandler(sessionID string, trackType DataTrackType) *DataTrackHandler {
	return NewDataTrackHandlerWithConfig(DataTrackConfig{
		SessionID: sessionID,
		TrackType: trackType,
	})
}

func NewDataTrackHandlerWithConfig(cfg DataTrackConfig) *DataTrackHandler {
	return &DataTrackHandler{
		sessionID:   cfg.SessionID,
		trackType:   cfg.TrackType,
		namespace:   dataTrackNamespace(cfg.SessionID, cfg.TrackType),
		publishers:  make(map[string]*dataTrackPublisher),
		subscribers: make(map[string]*dataTrackSubscriber),
		done:        make(chan struct{}),
	}
}

func (h *DataTrackHandler) Namespace() []string {
	return h.namespace
}

func (h *DataTrackHandler) TrackType() DataTrackType {
	return h.trackType
}

func (h *DataTrackHandler) HandleSubscribe(rw *moqtransport.SubscribeResponseWriter, msg *moqtransport.SubscribeMessage) {
	if h.closed.Load() {
		rw.Reject(moqtransport.ErrorCodeInternal, "handler closed")
		return
	}

	if len(msg.Namespace) != len(h.namespace) {
		rw.Reject(moqtransport.ErrorCodeSubscribeTrackDoesNotExist, "invalid namespace length")
		return
	}

	for i, ns := range h.namespace {
		if msg.Namespace[i] != ns {
			rw.Reject(moqtransport.ErrorCodeSubscribeTrackDoesNotExist, "namespace mismatch")
			return
		}
	}

	if err := rw.Accept(moqtransport.WithLargestLocation(&moqtransport.Location{Group: 0, Object: 0})); err != nil {
		return
	}

	slot := NewPublisherSlot()
	slot.Set(rw)

	writer := NewTrackWriter(slot)

	h.mu.Lock()
	if existing, ok := h.publishers[msg.Track]; ok {
		_ = existing.writer.Close()
	}
	h.publishers[msg.Track] = &dataTrackPublisher{writer: writer}
	h.mu.Unlock()
}

func (h *DataTrackHandler) Publish(ctx context.Context, trackName string, message *DataTrackMessage) error {
	if h.closed.Load() {
		return ErrConnectionClosed
	}

	h.mu.RLock()
	pub, ok := h.publishers[trackName]
	h.mu.RUnlock()

	if !ok {
		return ErrNoPublisher
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return pub.writer.Write(ctx, data)
}

func (h *DataTrackHandler) PublishStream(ctx context.Context, trackName string, messages []*DataTrackMessage) error {
	if h.closed.Load() {
		return ErrConnectionClosed
	}

	h.mu.RLock()
	pub, ok := h.publishers[trackName]
	h.mu.RUnlock()

	if !ok {
		return ErrNoPublisher
	}

	objects := make([][]byte, len(messages))
	for i, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		objects[i] = data
	}

	return pub.writer.WriteStream(ctx, objects)
}

func (h *DataTrackHandler) Subscribe(ctx context.Context, trackName string, handler func(*DataTrackMessage) error) error {
	if h.closed.Load() {
		return ErrConnectionClosed
	}

	subCtx, cancel := context.WithCancel(ctx)

	h.mu.Lock()
	if existing, ok := h.subscribers[trackName]; ok {
		existing.cancel()
	}
	h.subscribers[trackName] = &dataTrackSubscriber{
		handler: handler,
		ctx:     subCtx,
		cancel:  cancel,
	}
	h.mu.Unlock()

	return nil
}

func (h *DataTrackHandler) Unsubscribe(trackName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sub, ok := h.subscribers[trackName]; ok {
		sub.cancel()
		if sub.reader != nil {
			_ = sub.reader.Close()
		}
		delete(h.subscribers, trackName)
	}

	return nil
}

func (h *DataTrackHandler) HandleData(ctx context.Context, trackName string, data []byte) error {
	if h.closed.Load() {
		return ErrConnectionClosed
	}

	h.mu.RLock()
	subscriber, ok := h.subscribers[trackName]
	h.mu.RUnlock()

	if !ok {
		return ErrNoSubscriber
	}

	var message DataTrackMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}

	select {
	case <-subscriber.ctx.Done():
		return subscriber.ctx.Err()
	default:
		return subscriber.handler(&message)
	}
}

func (h *DataTrackHandler) HandleRemoteTrack(trackName string, track *moqtransport.RemoteTrack) error {
	if h.closed.Load() {
		return ErrConnectionClosed
	}

	reader := NewTrackReader(track)

	h.mu.Lock()
	if existing, ok := h.subscribers[trackName]; ok {
		if existing.reader != nil {
			_ = existing.reader.Close()
		}
		existing.reader = reader
	}
	h.mu.Unlock()

	return nil
}

func (h *DataTrackHandler) GetTrackInfo() []DataTrackInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tracks := make([]DataTrackInfo, 0, len(h.publishers))
	for trackName := range h.publishers {
		tracks = append(tracks, DataTrackInfo{
			Type:      h.trackType,
			TrackName: trackName,
			Namespace: h.namespace,
		})
	}

	return tracks
}

func (h *DataTrackHandler) Close() error {
	if h.closed.CompareAndSwap(false, true) {
		close(h.done)

		h.mu.Lock()
		defer h.mu.Unlock()

		for _, pub := range h.publishers {
			_ = pub.writer.Close()
		}
		h.publishers = make(map[string]*dataTrackPublisher)

		for _, sub := range h.subscribers {
			sub.cancel()
			if sub.reader != nil {
				_ = sub.reader.Close()
			}
		}
		h.subscribers = make(map[string]*dataTrackSubscriber)
	}
	return nil
}

func (h *DataTrackHandler) Done() <-chan struct{} {
	return h.done
}

func (h *DataTrackHandler) IsClosed() bool {
	return h.closed.Load()
}
