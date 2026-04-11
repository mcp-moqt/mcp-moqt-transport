package mcpmoqt

import (
	"context"
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAckStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status mcpmoqt.AckStatus
	}{
		{"pending", mcpmoqt.AckStatusPending},
		{"acked", mcpmoqt.AckStatusAcked},
		{"nacked", mcpmoqt.AckStatusNacked},
		{"timeout", mcpmoqt.AckStatusTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.status)
		})
	}
}

func TestDefaultAckConfig(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()

	assert.Equal(t, 5*time.Second, config.Timeout)
	assert.Equal(t, 3, config.RetryCount)
	assert.Equal(t, 100*time.Millisecond, config.RetryInterval)
}

func TestNewAckTracker(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)

	require.NotNil(t, tracker)
	assert.Equal(t, 0, tracker.PendingCount())
	assert.False(t, tracker.IsClosed())
}

func TestAckTracker_Track(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	ackCh := tracker.Track("msg-1")

	require.NotNil(t, ackCh)
	assert.Equal(t, 1, tracker.PendingCount())
}

func TestAckTracker_Ack(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	ackCh := tracker.Track("msg-1")
	tracker.Ack("msg-1")

	select {
	case ack := <-ackCh:
		assert.Equal(t, mcpmoqt.AckStatusAcked, ack.Status)
		assert.Equal(t, "msg-1", ack.ID)
		assert.Nil(t, ack.Error)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for ack")
	}

	assert.Equal(t, 0, tracker.PendingCount())
}

func TestAckTracker_Nack(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	ackCh := tracker.Track("msg-1")
	testErr := assert.AnError
	tracker.Nack("msg-1", testErr)

	select {
	case ack := <-ackCh:
		assert.Equal(t, mcpmoqt.AckStatusNacked, ack.Status)
		assert.Equal(t, "msg-1", ack.ID)
		assert.Equal(t, testErr, ack.Error)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for nack")
	}

	assert.Equal(t, 0, tracker.PendingCount())
}

func TestAckTracker_Timeout(t *testing.T) {
	config := mcpmoqt.AckConfig{
		Timeout:       100 * time.Millisecond,
		RetryCount:    3,
		RetryInterval: 10 * time.Millisecond,
	}
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	ackCh := tracker.Track("msg-1")

	select {
	case ack := <-ackCh:
		assert.Equal(t, mcpmoqt.AckStatusTimeout, ack.Status)
		assert.Equal(t, "msg-1", ack.ID)
		assert.Equal(t, mcpmoqt.ErrTimeout, ack.Error)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for timeout")
	}

	assert.Equal(t, 0, tracker.PendingCount())
}

func TestAckTracker_Close(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)

	ackCh := tracker.Track("msg-1")
	assert.Equal(t, 1, tracker.PendingCount())

	tracker.Close()
	assert.True(t, tracker.IsClosed())

	select {
	case ack := <-ackCh:
		assert.Equal(t, mcpmoqt.AckStatusNacked, ack.Status)
		assert.Equal(t, mcpmoqt.ErrConnectionClosed, ack.Error)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for close")
	}

	assert.Equal(t, 0, tracker.PendingCount())
}

func TestAckTracker_Close_Multiple(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)

	tracker.Close()
	tracker.Close()
}

func TestAckTracker_TrackAfterClose(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	tracker.Close()

	ackCh := tracker.Track("msg-1")

	select {
	case ack := <-ackCh:
		assert.Equal(t, mcpmoqt.AckStatusNacked, ack.Status)
		assert.Equal(t, mcpmoqt.ErrConnectionClosed, ack.Error)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for ack")
	}
}

func TestAckTracker_MultipleMessages(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	ackCh1 := tracker.Track("msg-1")
	ackCh2 := tracker.Track("msg-2")
	ackCh3 := tracker.Track("msg-3")

	assert.Equal(t, 3, tracker.PendingCount())

	tracker.Ack("msg-1")
	tracker.Nack("msg-2", assert.AnError)
	tracker.Ack("msg-3")

	ack1 := <-ackCh1
	assert.Equal(t, mcpmoqt.AckStatusAcked, ack1.Status)

	ack2 := <-ackCh2
	assert.Equal(t, mcpmoqt.AckStatusNacked, ack2.Status)

	ack3 := <-ackCh3
	assert.Equal(t, mcpmoqt.AckStatusAcked, ack3.Status)

	assert.Equal(t, 0, tracker.PendingCount())
}

func TestAckTracker_AckNonExistent(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	tracker.Ack("non-existent")

	assert.Equal(t, 0, tracker.PendingCount())
}

func TestAckTracker_NackNonExistent(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	tracker.Nack("non-existent", assert.AnError)

	assert.Equal(t, 0, tracker.PendingCount())
}

func TestAckTracker_ConcurrentAccess(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	done := make(chan bool)
	count := 100

	for i := 0; i < count; i++ {
		go func(id int) {
			msgID := string(rune(id))
			tracker.Track(msgID)
			done <- true
		}(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}

	assert.Equal(t, count, tracker.PendingCount())
}

func TestAckTracker_ContextCancellation(t *testing.T) {
	config := mcpmoqt.DefaultAckConfig()
	tracker := mcpmoqt.NewAckTracker(config)
	defer tracker.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	ackCh := tracker.Track("msg-1")

	select {
	case <-ctx.Done():
		t.Log("context cancelled as expected")
	case ack := <-ackCh:
		t.Logf("received ack: %v", ack.Status)
	}
}
