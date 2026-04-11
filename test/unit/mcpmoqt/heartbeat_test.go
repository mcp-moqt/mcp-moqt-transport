package mcpmoqt

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultHeartbeatConfig(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()

	assert.Equal(t, 30*time.Second, config.Interval)
	assert.Equal(t, 10*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxFailures)
}

func TestNewHeartbeat(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()
	hb := mcpmoqt.NewHeartbeat(config)

	require.NotNil(t, hb)
	assert.False(t, hb.IsRunning())
	assert.False(t, hb.IsHealthy())
}

func TestHeartbeat_Start(t *testing.T) {
	config := mcpmoqt.HeartbeatConfig{
		Interval:    100 * time.Millisecond,
		Timeout:     50 * time.Millisecond,
		MaxFailures: 3,
	}
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	err := hb.Start(ctx)
	require.NoError(t, err)
	assert.True(t, hb.IsRunning())

	hb.Stop()
	assert.False(t, hb.IsRunning())
}

func TestHeartbeat_StartTwice(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	err := hb.Start(ctx)
	require.NoError(t, err)

	err = hb.Start(ctx)
	require.NoError(t, err)

	hb.Stop()
}

func TestHeartbeat_Stop(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	hb.Start(ctx)
	hb.Stop()

	assert.False(t, hb.IsRunning())
}

func TestHeartbeat_StopTwice(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	hb.Start(ctx)
	hb.Stop()
	hb.Stop()

	assert.False(t, hb.IsRunning())
}

func TestHeartbeat_OnHeartbeat(t *testing.T) {
	var count atomic.Int32
	config := mcpmoqt.HeartbeatConfig{
		Interval:    50 * time.Millisecond,
		Timeout:     10 * time.Millisecond,
		MaxFailures: 3,
		OnHeartbeat: func() error {
			count.Add(1)
			return nil
		},
	}
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	err := hb.Start(ctx)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	hb.Stop()

	assert.GreaterOrEqual(t, count.Load(), int32(2))
}

func TestHeartbeat_OnTimeout(t *testing.T) {
	var timeoutCount atomic.Int32
	config := mcpmoqt.HeartbeatConfig{
		Interval:    50 * time.Millisecond,
		Timeout:     10 * time.Millisecond,
		MaxFailures: 3,
		OnHeartbeat: func() error {
			return assert.AnError
		},
		OnTimeout: func() {
			timeoutCount.Add(1)
		},
	}
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	err := hb.Start(ctx)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	hb.Stop()

	assert.GreaterOrEqual(t, timeoutCount.Load(), int32(1))
}

func TestHeartbeat_OnReconnect(t *testing.T) {
	var reconnectCount atomic.Int32
	config := mcpmoqt.HeartbeatConfig{
		Interval:    50 * time.Millisecond,
		Timeout:     10 * time.Millisecond,
		MaxFailures: 2,
		OnHeartbeat: func() error {
			return assert.AnError
		},
		OnReconnect: func() {
			reconnectCount.Add(1)
		},
	}
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	err := hb.Start(ctx)
	require.NoError(t, err)

	time.Sleep(300 * time.Millisecond)
	hb.Stop()

	assert.GreaterOrEqual(t, reconnectCount.Load(), int32(1))
}

func TestHeartbeat_RecordPong(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	hb.Start(ctx)
	defer hb.Stop()

	before := hb.LastPong()
	time.Sleep(10 * time.Millisecond)
	hb.RecordPong()
	after := hb.LastPong()

	assert.True(t, after.After(before))
	assert.Equal(t, 0, hb.Failures())
}

func TestHeartbeat_LastPong(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()
	hb := mcpmoqt.NewHeartbeat(config)

	lastPong := hb.LastPong()
	assert.False(t, lastPong.IsZero())
}

func TestHeartbeat_Failures(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()
	hb := mcpmoqt.NewHeartbeat(config)

	assert.Equal(t, 0, hb.Failures())
}

func TestHeartbeat_IsHealthy(t *testing.T) {
	config := mcpmoqt.HeartbeatConfig{
		Interval:    100 * time.Millisecond,
		Timeout:     50 * time.Millisecond,
		MaxFailures: 3,
		OnHeartbeat: func() error {
			return nil
		},
	}
	hb := mcpmoqt.NewHeartbeat(config)

	assert.False(t, hb.IsHealthy())

	ctx := context.Background()
	hb.Start(ctx)
	defer hb.Stop()

	time.Sleep(50 * time.Millisecond)
	assert.True(t, hb.IsHealthy())
}

func TestHeartbeat_IsHealthy_AfterStop(t *testing.T) {
	config := mcpmoqt.DefaultHeartbeatConfig()
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	hb.Start(ctx)
	hb.Stop()

	assert.False(t, hb.IsHealthy())
}

func TestHeartbeat_ContextCancellation(t *testing.T) {
	config := mcpmoqt.HeartbeatConfig{
		Interval:    50 * time.Millisecond,
		Timeout:     10 * time.Millisecond,
		MaxFailures: 3,
		OnHeartbeat: func() error {
			return nil
		},
	}
	hb := mcpmoqt.NewHeartbeat(config)

	ctx, cancel := context.WithCancel(context.Background())
	err := hb.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	assert.False(t, hb.IsRunning())
}

func TestHeartbeat_NoHeartbeatCallback(t *testing.T) {
	config := mcpmoqt.HeartbeatConfig{
		Interval:    50 * time.Millisecond,
		Timeout:     10 * time.Millisecond,
		MaxFailures: 3,
	}
	hb := mcpmoqt.NewHeartbeat(config)

	ctx := context.Background()
	err := hb.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	hb.Stop()

	assert.False(t, hb.IsRunning())
}

func TestHeartbeat_ConcurrentStartStop(t *testing.T) {
	config := mcpmoqt.HeartbeatConfig{
		Interval:    10 * time.Millisecond,
		Timeout:     5 * time.Millisecond,
		MaxFailures: 3,
		OnHeartbeat: func() error {
			return nil
		},
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			hb := mcpmoqt.NewHeartbeat(config)
			ctx := context.Background()
			hb.Start(ctx)
			time.Sleep(20 * time.Millisecond)
			hb.Stop()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
