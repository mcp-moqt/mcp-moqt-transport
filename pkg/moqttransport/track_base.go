package mcpmoqt

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/mengelbart/moqtransport"
)

type PublisherSlot struct {
	mu      sync.Mutex
	pub     moqtransport.Publisher
	ready   chan struct{}
	cleared bool
}

func NewPublisherSlot() *PublisherSlot {
	return &PublisherSlot{ready: make(chan struct{})}
}

func (s *PublisherSlot) Set(pub moqtransport.Publisher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pub != nil || s.cleared {
		return
	}
	s.pub = pub
	close(s.ready)
}

func (s *PublisherSlot) Get(ctx context.Context) (moqtransport.Publisher, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.ready:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cleared {
		return nil, ErrConnectionClosed
	}
	return s.pub, nil
}

func (s *PublisherSlot) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleared = true
	s.pub = nil
}

type TrackWriter struct {
	slot      *PublisherSlot
	nextGroup atomic.Uint64
	closed    atomic.Bool
}

func NewTrackWriter(slot *PublisherSlot) *TrackWriter {
	return &TrackWriter{slot: slot}
}

func (w *TrackWriter) Write(ctx context.Context, data []byte) error {
	if w.closed.Load() {
		return ErrConnectionClosed
	}

	pub, err := w.slot.Get(ctx)
	if err != nil {
		return err
	}

	groupID := w.nextGroup.Add(1) - 1
	sg, err := pub.OpenSubgroup(groupID, 0, 0)
	if err != nil {
		return err
	}

	if _, err := sg.WriteObject(0, data); err != nil {
		_ = sg.Close()
		return err
	}

	return sg.Close()
}

func (w *TrackWriter) WriteStream(ctx context.Context, objects [][]byte) error {
	if w.closed.Load() {
		return ErrConnectionClosed
	}

	pub, err := w.slot.Get(ctx)
	if err != nil {
		return err
	}

	groupID := w.nextGroup.Add(1) - 1
	sg, err := pub.OpenSubgroup(groupID, 0, 0)
	if err != nil {
		return err
	}

	for i, obj := range objects {
		if _, err := sg.WriteObject(uint64(i), obj); err != nil {
			_ = sg.Close()
			return err
		}
	}

	return sg.Close()
}

func (w *TrackWriter) Close() error {
	w.closed.Store(true)
	w.slot.Clear()
	return nil
}

type TrackReader struct {
	track  *moqtransport.RemoteTrack
	closed atomic.Bool
	done   chan struct{}
}

func NewTrackReader(track *moqtransport.RemoteTrack) *TrackReader {
	return &TrackReader{
		track: track,
		done:  make(chan struct{}),
	}
}

func (r *TrackReader) Read(ctx context.Context) ([]byte, error) {
	if r.closed.Load() {
		return nil, ErrConnectionClosed
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.done:
		return nil, ErrConnectionClosed
	default:
	}

	obj, err := r.track.ReadObject(ctx)
	if err != nil {
		return nil, err
	}

	return obj.Payload, nil
}

func (r *TrackReader) Close() error {
	if r.closed.CompareAndSwap(false, true) {
		close(r.done)
		if r.track != nil {
			return r.track.Close()
		}
	}
	return nil
}

type DataBuffer struct {
	pool sync.Pool
}

func NewDataBuffer() *DataBuffer {
	return &DataBuffer{
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 4096)
				return &buf
			},
		},
	}
}

func (b *DataBuffer) Get() *[]byte {
	return b.pool.Get().(*[]byte)
}

func (b *DataBuffer) Put(buf *[]byte) {
	*buf = (*buf)[:0]
	b.pool.Put(buf)
}

var defaultBufferPool = NewDataBuffer()
