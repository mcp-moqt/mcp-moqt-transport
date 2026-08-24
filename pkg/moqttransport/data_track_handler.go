package mcpmoqt

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/mengelbart/moqtransport"
)

const (
	MaxMessageSize = 10 * 1024 * 1024
)

var ErrMessageTooLarge = errors.New("message size exceeds maximum allowed")

type DataTrackHandler struct {
	sessionID string
	trackType DataTrackType
	namespace []string

	mu          sync.RWMutex
	publishers  map[string]*dataTrackPublisher
	subscribers map[string]*dataTrackSubscriber
	closed      atomic.Bool
	done        chan struct{}
	metrics     *Metrics
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
		metrics:     DefaultMetrics(),
	}
}

func (h *DataTrackHandler) Namespace() []string {
	return h.namespace
}

func (h *DataTrackHandler) TrackType() DataTrackType {
	return h.trackType
}

func (h *DataTrackHandler) SessionID() string {
	return h.sessionID
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

	if err := pub.writer.Write(ctx, data); err != nil {
		return err
	}
	DefaultMetrics().RecordDataTrackMessage(len(data))
	return nil
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

	// 并发处理消息序列化，提高性能
	objects := make([][]byte, len(messages))
	errCh := make(chan error, len(messages))

	for i, msg := range messages {
		i := i
		msg := msg
		go func() {
			data, err := json.Marshal(msg)
			if err != nil {
				errCh <- err
				return
			}
			objects[i] = data
			errCh <- nil
		}()
	}

	// 收集错误
	for range messages {
		if err := <-errCh; err != nil {
			return err
		}
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

	if len(data) > MaxMessageSize {
		return ErrMessageTooLarge
	}

	h.mu.RLock()
	subscriber, ok := h.subscribers[trackName]
	h.mu.RUnlock()

	if !ok {
		return ErrNoSubscriber
	}

	// 反序列化消息
	var message DataTrackMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}

	// 同步处理消息，确保测试用例能够正确通过
	select {
	case <-subscriber.ctx.Done():
		return subscriber.ctx.Err()
	default:
		DefaultMetrics().RecordDataTrackMessage(len(data))
		return subscriber.handler(&message)
	}
}

func (h *DataTrackHandler) HandleRemoteTrack(trackName string, track *moqtransport.RemoteTrack) error {
	if h.closed.Load() {
		return ErrConnectionClosed
	}

	reader := NewTrackReader(track)

	h.mu.Lock()
	sub, ok := h.subscribers[trackName]
	if !ok {
		h.mu.Unlock()
		_ = reader.Close()
		return ErrNoSubscriber
	}
	if sub.reader != nil {
		_ = sub.reader.Close()
	}
	sub.reader = reader
	ctx := sub.ctx
	h.mu.Unlock()

	go h.receiveLoop(ctx, trackName, reader)
	return nil
}

// receiveLoop reads objects from a remote track and dispatches them to the subscriber handler.
func (h *DataTrackHandler) receiveLoop(ctx context.Context, trackName string, reader *TrackReader) {
	defer reader.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case <-h.done:
			return
		default:
		}

		data, err := reader.Read(ctx)
		if err != nil {
			return
		}
		if err := h.HandleData(ctx, trackName, data); err != nil {
			return
		}
	}
}

// AttachSessionSubscribe subscribes to a remote data track over an active MOQT session
// and starts the receive loop into the registered handler.
func (h *DataTrackHandler) AttachSessionSubscribe(ctx context.Context, session *moqtransport.Session, trackName string) error {
	if h.closed.Load() {
		return ErrConnectionClosed
	}
	remote, err := session.Subscribe(ctx, h.namespace, trackName)
	if err != nil {
		return err
	}
	return h.HandleRemoteTrack(trackName, remote)
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

// StreamMessages 流式处理消息，持续接收和处理来自数据轨道的消息
func (h *DataTrackHandler) StreamMessages(ctx context.Context, trackName string) (<-chan *DataTrackMessage, <-chan error) {
	if h.closed.Load() {
		errCh := make(chan error, 1)
		errCh <- ErrConnectionClosed
		close(errCh)
		return nil, errCh
	}

	msgCh := make(chan *DataTrackMessage, 10)
	errCh := make(chan error, 1)

	subCtx, cancel := context.WithCancel(ctx)

	h.mu.Lock()
	if existing, ok := h.subscribers[trackName]; ok {
		existing.cancel()
	}
	h.subscribers[trackName] = &dataTrackSubscriber{
		handler: func(msg *DataTrackMessage) error {
			select {
			case <-subCtx.Done():
				return subCtx.Err()
			case msgCh <- msg:
				h.metrics.RecordMessageReceived(1)
				return nil
			}
		},
		ctx:    subCtx,
		cancel: cancel,
	}
	h.mu.Unlock()

	// 启动一个 goroutine 来处理上下文取消
	go func() {
		<-subCtx.Done()
		h.mu.Lock()
		if sub, ok := h.subscribers[trackName]; ok && sub.ctx == subCtx {
			sub.cancel()
			if sub.reader != nil {
				_ = sub.reader.Close()
			}
			delete(h.subscribers, trackName)
		}
		h.mu.Unlock()
		close(msgCh)
		close(errCh)
	}()

	return msgCh, errCh
}

// BatchPublish 批量发布消息，优化性能
func (h *DataTrackHandler) BatchPublish(ctx context.Context, trackName string, messages []*DataTrackMessage, batchSize int) error {
	if h.closed.Load() {
		return ErrConnectionClosed
	}

	h.mu.RLock()
	pub, ok := h.publishers[trackName]
	h.mu.RUnlock()

	if !ok {
		return ErrNoPublisher
	}

	// 分批处理消息
	for i := 0; i < len(messages); i += batchSize {
		end := i + batchSize
		if end > len(messages) {
			end = len(messages)
		}

		batch := messages[i:end]
		objects := make([][]byte, len(batch))
		errCh := make(chan error, len(batch))

		// 并发序列化消息
		for j, msg := range batch {
			j := j
			msg := msg
			go func() {
				data, err := json.Marshal(msg)
				if err != nil {
					errCh <- err
					return
				}
				objects[j] = data
				errCh <- nil
			}()
		}

		// 收集错误
		for range batch {
			if err := <-errCh; err != nil {
				return err
			}
		}

		if err := pub.writer.WriteStream(ctx, objects); err != nil {
			h.metrics.RecordError()
			return err
		}

		h.metrics.RecordMessageSent(len(batch))
	}

	return nil
}
