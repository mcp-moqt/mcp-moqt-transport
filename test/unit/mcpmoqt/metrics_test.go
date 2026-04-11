package mcpmoqt

import (
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	require.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.MessagesSent())
	assert.Equal(t, int64(0), metrics.MessagesReceived())
	assert.Equal(t, int64(0), metrics.BytesSent())
	assert.Equal(t, int64(0), metrics.BytesReceived())
}

func TestMetrics_RecordMessageSent(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordMessageSent(100)
	assert.Equal(t, int64(1), metrics.MessagesSent())
	assert.Equal(t, int64(100), metrics.BytesSent())

	metrics.RecordMessageSent(200)
	assert.Equal(t, int64(2), metrics.MessagesSent())
	assert.Equal(t, int64(300), metrics.BytesSent())
}

func TestMetrics_RecordMessageReceived(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordMessageReceived(100)
	assert.Equal(t, int64(1), metrics.MessagesReceived())
	assert.Equal(t, int64(100), metrics.BytesReceived())

	metrics.RecordMessageReceived(200)
	assert.Equal(t, int64(2), metrics.MessagesReceived())
	assert.Equal(t, int64(300), metrics.BytesReceived())
}

func TestMetrics_RecordAck(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordAck()
	assert.Equal(t, int64(1), metrics.MessagesAcked())

	metrics.RecordAck()
	assert.Equal(t, int64(2), metrics.MessagesAcked())
}

func TestMetrics_RecordNack(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordNack()
	assert.Equal(t, int64(1), metrics.MessagesNacked())

	metrics.RecordNack()
	assert.Equal(t, int64(2), metrics.MessagesNacked())
}

func TestMetrics_RecordTimeout(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordTimeout()
	assert.Equal(t, int64(1), metrics.MessagesTimeout())

	metrics.RecordTimeout()
	assert.Equal(t, int64(2), metrics.MessagesTimeout())
}

func TestMetrics_RecordError(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordError()
	assert.Equal(t, int64(1), metrics.Errors())

	metrics.RecordError()
	assert.Equal(t, int64(2), metrics.Errors())
}

func TestMetrics_RecordRetry(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordRetry()
	assert.Equal(t, int64(1), metrics.Retries())

	metrics.RecordRetry()
	assert.Equal(t, int64(2), metrics.Retries())
}

func TestMetrics_RecordHeartbeatSent(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordHeartbeatSent()
	assert.Equal(t, int64(1), metrics.HeartbeatsSent())

	metrics.RecordHeartbeatSent()
	assert.Equal(t, int64(2), metrics.HeartbeatsSent())
}

func TestMetrics_RecordHeartbeatFailed(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordHeartbeatFailed()
	assert.Equal(t, int64(1), metrics.HeartbeatsFailed())

	metrics.RecordHeartbeatFailed()
	assert.Equal(t, int64(2), metrics.HeartbeatsFailed())
}

func TestMetrics_RecordConnectionOpened(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordConnectionOpened()
	assert.Equal(t, int64(1), metrics.ConnectionsTotal())
	assert.Equal(t, int64(1), metrics.ConnectionsActive())

	metrics.RecordConnectionOpened()
	assert.Equal(t, int64(2), metrics.ConnectionsTotal())
	assert.Equal(t, int64(2), metrics.ConnectionsActive())
}

func TestMetrics_RecordConnectionClosed(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordConnectionOpened()
	metrics.RecordConnectionOpened()
	assert.Equal(t, int64(2), metrics.ConnectionsActive())

	metrics.RecordConnectionClosed()
	assert.Equal(t, int64(1), metrics.ConnectionsActive())
	assert.Equal(t, int64(2), metrics.ConnectionsTotal())
}

func TestMetrics_RecordTrackOpened(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordTrackOpened()
	assert.Equal(t, int64(1), metrics.TracksTotal())
	assert.Equal(t, int64(1), metrics.TracksActive())

	metrics.RecordTrackOpened()
	assert.Equal(t, int64(2), metrics.TracksTotal())
	assert.Equal(t, int64(2), metrics.TracksActive())
}

func TestMetrics_RecordTrackClosed(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordTrackOpened()
	metrics.RecordTrackOpened()
	assert.Equal(t, int64(2), metrics.TracksActive())

	metrics.RecordTrackClosed()
	assert.Equal(t, int64(1), metrics.TracksActive())
	assert.Equal(t, int64(2), metrics.TracksTotal())
}

func TestMetrics_Uptime(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	uptime := metrics.Uptime()
	assert.GreaterOrEqual(t, uptime, time.Duration(0))
	assert.Less(t, uptime, time.Second)
}

func TestMetrics_Snapshot(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordMessageSent(100)
	metrics.RecordMessageReceived(50)
	metrics.RecordAck()
	metrics.RecordNack()
	metrics.RecordError()
	metrics.RecordConnectionOpened()

	snapshot := metrics.Snapshot()

	assert.Equal(t, int64(1), snapshot.MessagesSent)
	assert.Equal(t, int64(1), snapshot.MessagesReceived)
	assert.Equal(t, int64(100), snapshot.BytesSent)
	assert.Equal(t, int64(50), snapshot.BytesReceived)
	assert.Equal(t, int64(1), snapshot.MessagesAcked)
	assert.Equal(t, int64(1), snapshot.MessagesNacked)
	assert.Equal(t, int64(1), snapshot.Errors)
	assert.Equal(t, int64(1), snapshot.ConnectionsTotal)
	assert.Equal(t, int64(1), snapshot.ConnectionsActive)
}

func TestMetrics_Snapshot_AckRate(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	snapshot := metrics.Snapshot()
	assert.Equal(t, float64(0), snapshot.AckRate)

	metrics.RecordMessageSent(100)
	metrics.RecordMessageSent(100)
	metrics.RecordAck()

	snapshot = metrics.Snapshot()
	assert.Equal(t, 0.5, snapshot.AckRate)
}

func TestMetrics_Snapshot_ErrorRate(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	snapshot := metrics.Snapshot()
	assert.Equal(t, float64(0), snapshot.ErrorRate)

	metrics.RecordMessageSent(100)
	metrics.RecordMessageReceived(100)
	metrics.RecordError()

	snapshot = metrics.Snapshot()
	assert.Equal(t, 0.5, snapshot.ErrorRate)
}

func TestMetrics_Snapshot_Throughput(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	time.Sleep(10 * time.Millisecond)

	metrics.RecordMessageSent(1000)
	metrics.RecordMessageReceived(500)

	snapshot := metrics.Snapshot()

	assert.GreaterOrEqual(t, snapshot.ThroughputOut, float64(0))
	assert.GreaterOrEqual(t, snapshot.ThroughputIn, float64(0))
}

func TestMetrics_Reset(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordMessageSent(100)
	metrics.RecordMessageReceived(50)
	metrics.RecordAck()
	metrics.RecordError()
	metrics.RecordConnectionOpened()

	metrics.Reset()

	assert.Equal(t, int64(0), metrics.MessagesSent())
	assert.Equal(t, int64(0), metrics.MessagesReceived())
	assert.Equal(t, int64(0), metrics.BytesSent())
	assert.Equal(t, int64(0), metrics.BytesReceived())
	assert.Equal(t, int64(0), metrics.MessagesAcked())
	assert.Equal(t, int64(0), metrics.Errors())
	assert.Equal(t, int64(0), metrics.ConnectionsTotal())
	assert.Equal(t, int64(0), metrics.ConnectionsActive())
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	done := make(chan bool)
	count := 100

	for i := 0; i < count; i++ {
		go func() {
			metrics.RecordMessageSent(10)
			metrics.RecordMessageReceived(10)
			metrics.RecordAck()
			done <- true
		}()
	}

	for i := 0; i < count; i++ {
		<-done
	}

	assert.Equal(t, int64(count), metrics.MessagesSent())
	assert.Equal(t, int64(count), metrics.MessagesReceived())
	assert.Equal(t, int64(count), metrics.MessagesAcked())
}

func TestDefaultMetrics(t *testing.T) {
	metrics := mcpmoqt.DefaultMetrics()

	require.NotNil(t, metrics)
}

func TestMetrics_AllCounters(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordMessageSent(100)
	metrics.RecordMessageReceived(50)
	metrics.RecordAck()
	metrics.RecordNack()
	metrics.RecordTimeout()
	metrics.RecordError()
	metrics.RecordRetry()
	metrics.RecordHeartbeatSent()
	metrics.RecordHeartbeatFailed()
	metrics.RecordConnectionOpened()
	metrics.RecordTrackOpened()

	assert.Equal(t, int64(1), metrics.MessagesSent())
	assert.Equal(t, int64(1), metrics.MessagesReceived())
	assert.Equal(t, int64(1), metrics.MessagesAcked())
	assert.Equal(t, int64(1), metrics.MessagesNacked())
	assert.Equal(t, int64(1), metrics.MessagesTimeout())
	assert.Equal(t, int64(1), metrics.Errors())
	assert.Equal(t, int64(1), metrics.Retries())
	assert.Equal(t, int64(1), metrics.HeartbeatsSent())
	assert.Equal(t, int64(1), metrics.HeartbeatsFailed())
	assert.Equal(t, int64(1), metrics.ConnectionsTotal())
	assert.Equal(t, int64(1), metrics.ConnectionsActive())
	assert.Equal(t, int64(1), metrics.TracksTotal())
	assert.Equal(t, int64(1), metrics.TracksActive())
}

func TestMetrics_NegativeActiveConnections(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordConnectionClosed()
	assert.Equal(t, int64(-1), metrics.ConnectionsActive())
}

func TestMetrics_NegativeActiveTracks(t *testing.T) {
	metrics := mcpmoqt.NewMetrics()

	metrics.RecordTrackClosed()
	assert.Equal(t, int64(-1), metrics.TracksActive())
}
