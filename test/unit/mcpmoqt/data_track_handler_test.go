package mcpmoqt

import (
	"context"
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTrackWriter 是一个模拟的 TrackWriter
type mockTrackWriter struct{}

func (m *mockTrackWriter) Write(ctx context.Context, data []byte) error {
	return nil
}

func (m *mockTrackWriter) WriteStream(ctx context.Context, objects [][]byte) error {
	return nil
}

func (m *mockTrackWriter) Close() error {
	return nil
}

func TestNewDataTrackHandler(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	require.NotNil(t, handler)
	assert.Equal(t, "session-123", handler.SessionID())
	assert.Equal(t, mcpmoqt.DataTrackResources, handler.TrackType())
	assert.False(t, handler.IsClosed())
}

func TestNewDataTrackHandlerWithConfig(t *testing.T) {
	config := mcpmoqt.DataTrackConfig{
		SessionID: "session-456",
		TrackType: mcpmoqt.DataTrackTools,
	}
	handler := mcpmoqt.NewDataTrackHandlerWithConfig(config)

	require.NotNil(t, handler)
	assert.Equal(t, "session-456", handler.SessionID())
	assert.Equal(t, mcpmoqt.DataTrackTools, handler.TrackType())
}

func TestDataTrackHandler_Namespace(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ns := handler.Namespace()
	require.Len(t, ns, 3)
	assert.Equal(t, "mcp", ns[0])
	assert.Equal(t, "session-123", ns[1])
	assert.Equal(t, "resources", ns[2])
}

func TestDataTrackHandler_TrackType(t *testing.T) {
	tests := []struct {
		name      string
		trackType mcpmoqt.DataTrackType
	}{
		{"resources", mcpmoqt.DataTrackResources},
		{"tools", mcpmoqt.DataTrackTools},
		{"notifications", mcpmoqt.DataTrackNotifications},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := mcpmoqt.NewDataTrackHandler("session-123", tt.trackType)
			assert.Equal(t, tt.trackType, handler.TrackType())
		})
	}
}

func TestDataTrackHandler_Subscribe(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()
	err := handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		return nil
	})

	require.NoError(t, err)
}

func TestDataTrackHandler_Subscribe_Replace(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()
	err := handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		return nil
	})
	require.NoError(t, err)

	err = handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		return nil
	})
	require.NoError(t, err)
}

func TestDataTrackHandler_Subscribe_Closed(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)
	handler.Close()

	ctx := context.Background()
	err := handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		return nil
	})

	require.Error(t, err)
	assert.Equal(t, mcpmoqt.ErrConnectionClosed, err)
}

func TestDataTrackHandler_PublishStream_Concurrent(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()

	// 测试错误情况：没有发布者
	err := handler.PublishStream(ctx, "track-1", []*mcpmoqt.DataTrackMessage{})
	require.Error(t, err)
	assert.Equal(t, mcpmoqt.ErrNoPublisher, err)
}

func TestDataTrackHandler_HandleData_Concurrent(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()

	// 订阅轨道
	messageCount := 0
	err := handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		messageCount++
		return nil
	})
	require.NoError(t, err)

	// 并发处理多条消息
	for i := 0; i < 10; i++ {
		data := []byte(`{"track_type": 0, "track_name": "track-1", "data": {"key": 1}}`)
		err := handler.HandleData(ctx, "track-1", data)
		require.NoError(t, err)
	}

	// 等待并发处理完成
	time.Sleep(100 * time.Millisecond)
}

func TestDataTrackHandler_Unsubscribe(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()
	err := handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		return nil
	})
	require.NoError(t, err)

	err = handler.Unsubscribe("track-1")
	require.NoError(t, err)
}

func TestDataTrackHandler_Unsubscribe_NonExistent(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	err := handler.Unsubscribe("non-existent")
	require.NoError(t, err)
}

func TestDataTrackHandler_HandleData(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	received := false
	ctx := context.Background()
	err := handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		received = true
		return nil
	})
	require.NoError(t, err)

	err = handler.HandleData(ctx, "track-1", []byte(`{"track_type":0,"track_name":"track-1","data":{}}`))
	require.NoError(t, err)
	assert.True(t, received)
}

func TestDataTrackHandler_HandleData_NoSubscriber(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()
	err := handler.HandleData(ctx, "non-existent", []byte(`{}`))

	require.Error(t, err)
	assert.Equal(t, mcpmoqt.ErrNoSubscriber, err)
}

func TestDataTrackHandler_HandleData_Closed(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)
	handler.Close()

	ctx := context.Background()
	err := handler.HandleData(ctx, "track-1", []byte(`{}`))

	require.Error(t, err)
	assert.Equal(t, mcpmoqt.ErrConnectionClosed, err)
}

func TestDataTrackHandler_HandleData_InvalidJSON(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()
	err := handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		return nil
	})
	require.NoError(t, err)

	err = handler.HandleData(ctx, "track-1", []byte(`invalid json`))
	require.Error(t, err)
}

func TestDataTrackHandler_Publish_NoPublisher(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()
	msg := &mcpmoqt.DataTrackMessage{
		TrackType: mcpmoqt.DataTrackResources,
		TrackName: "track-1",
		Data:      map[string]interface{}{"key": "value"},
	}

	err := handler.Publish(ctx, "track-1", msg)
	require.Error(t, err)
	assert.Equal(t, mcpmoqt.ErrNoPublisher, err)
}

func TestDataTrackHandler_Publish_Closed(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)
	handler.Close()

	ctx := context.Background()
	msg := &mcpmoqt.DataTrackMessage{
		TrackType: mcpmoqt.DataTrackResources,
		TrackName: "track-1",
		Data:      map[string]interface{}{"key": "value"},
	}

	err := handler.Publish(ctx, "track-1", msg)
	require.Error(t, err)
	assert.Equal(t, mcpmoqt.ErrConnectionClosed, err)
}

func TestDataTrackHandler_PublishStream_NoPublisher(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()
	msgs := []*mcpmoqt.DataTrackMessage{
		{TrackType: mcpmoqt.DataTrackResources, TrackName: "track-1", Data: map[string]interface{}{"key": "value"}},
	}

	err := handler.PublishStream(ctx, "track-1", msgs)
	require.Error(t, err)
	assert.Equal(t, mcpmoqt.ErrNoPublisher, err)
}

func TestDataTrackHandler_PublishStream_Closed(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)
	handler.Close()

	ctx := context.Background()
	msgs := []*mcpmoqt.DataTrackMessage{
		{TrackType: mcpmoqt.DataTrackResources, TrackName: "track-1", Data: map[string]interface{}{"key": "value"}},
	}

	err := handler.PublishStream(ctx, "track-1", msgs)
	require.Error(t, err)
	assert.Equal(t, mcpmoqt.ErrConnectionClosed, err)
}

func TestDataTrackHandler_GetTrackInfo(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	info := handler.GetTrackInfo()
	assert.Empty(t, info)
}

func TestDataTrackHandler_Close(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	err := handler.Close()
	require.NoError(t, err)
	assert.True(t, handler.IsClosed())
}

func TestDataTrackHandler_Close_Multiple(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	err := handler.Close()
	require.NoError(t, err)

	err = handler.Close()
	require.NoError(t, err)
}

func TestDataTrackHandler_Done(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	done := handler.Done()
	require.NotNil(t, done)

	select {
	case <-done:
		t.Fatal("done channel should not be closed yet")
	default:
	}

	handler.Close()

	select {
	case <-done:
	default:
		t.Fatal("done channel should be closed")
	}
}

func TestDataTrackHandler_IsClosed(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	assert.False(t, handler.IsClosed())
	handler.Close()
	assert.True(t, handler.IsClosed())
}

func TestDataTrackHandler_ConcurrentSubscribe(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	done := make(chan bool)
	count := 100

	for i := 0; i < count; i++ {
		go func(id int) {
			ctx := context.Background()
			trackName := string(rune(id))
			_ = handler.Subscribe(ctx, trackName, func(msg *mcpmoqt.DataTrackMessage) error {
				return nil
			})
			done <- true
		}(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}
}

func TestDataTrackHandler_ConcurrentClose(t *testing.T) {
	handler := mcpmoqt.NewDataTrackHandler("session-123", mcpmoqt.DataTrackResources)

	ctx := context.Background()
	_ = handler.Subscribe(ctx, "track-1", func(msg *mcpmoqt.DataTrackMessage) error {
		return nil
	})

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			handler.Close()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	assert.True(t, handler.IsClosed())
}

func TestDataTrackMessage(t *testing.T) {
	msg := &mcpmoqt.DataTrackMessage{
		TrackType: mcpmoqt.DataTrackResources,
		TrackName: "test-track",
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	assert.Equal(t, mcpmoqt.DataTrackResources, msg.TrackType)
	assert.Equal(t, "test-track", msg.TrackName)
	assert.NotNil(t, msg.Data)
}

func TestDataTrackInfo(t *testing.T) {
	info := mcpmoqt.DataTrackInfo{
		Type:      mcpmoqt.DataTrackResources,
		TrackName: "test-track",
		Namespace: []string{"mcp", "session-123", "resources"},
	}

	assert.Equal(t, mcpmoqt.DataTrackResources, info.Type)
	assert.Equal(t, "test-track", info.TrackName)
	assert.Len(t, info.Namespace, 3)
}
