package integration

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/require"
)

func TestDataTrack_PublishSubscribeE2E(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serverTransport, err := mcpmoqt.NewMOQTServerTransport(mcpmoqt.WithAddr(addr))
	require.NoError(t, err)
	defer serverTransport.Close()

	serverReady := make(chan struct{})
	var serverConnErr error
	go func() {
		close(serverReady)
		_, serverConnErr = serverTransport.Connect(ctx)
	}()

	<-serverReady
	time.Sleep(500 * time.Millisecond)

	clientTransport, err := mcpmoqt.NewMOQTClientTransport(mcpmoqt.WithAddr(addr))
	require.NoError(t, err)
	defer clientTransport.Close()

	conn, err := clientTransport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	require.Eventually(t, func() bool {
		return serverTransport.SessionID() != "" &&
			clientTransport.SessionID() != "" &&
			clientTransport.SessionID() == serverTransport.SessionID()
	}, 5*time.Second, 50*time.Millisecond)

	received := make(chan *mcpmoqt.DataTrackMessage, 1)
	var once sync.Once
	handler, err := clientTransport.SubscribeDataTrack(ctx, mcpmoqt.DataTrackResources, "demo", func(msg *mcpmoqt.DataTrackMessage) error {
		once.Do(func() { received <- msg })
		return nil
	})
	require.NoError(t, err)
	defer handler.Close()

	// Wait until server-side publisher slot is ready after subscribe accept.
	pub := serverTransport.DataTrackHandler(mcpmoqt.DataTrackResources)
	require.NotNil(t, pub)

	deadline := time.Now().Add(5 * time.Second)
	for {
		err = pub.Publish(ctx, "demo", &mcpmoqt.DataTrackMessage{
			TrackType: mcpmoqt.DataTrackResources,
			TrackName: "demo",
			Data:      map[string]any{"hello": "moqt"},
		})
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			require.NoError(t, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	select {
	case msg := <-received:
		require.Equal(t, "demo", msg.TrackName)
		require.Equal(t, mcpmoqt.DataTrackResources, msg.TrackType)
	case <-time.After(8 * time.Second):
		t.Fatal("timed out waiting for data-track message")
	}

	_ = serverConnErr
}

func TestDataTrack_PublishStreamAndBatch(t *testing.T) {
	h := mcpmoqt.NewDataTrackHandler("session-test", mcpmoqt.DataTrackTools)
	defer h.Close()

	slot := mcpmoqt.NewPublisherSlot()
	// Simulate accepted subscribe by injecting a publisher via HandleSubscribe is
	// network-bound; exercise batch/stream APIs against ErrNoPublisher when empty.
	err := h.PublishStream(context.Background(), "t", []*mcpmoqt.DataTrackMessage{
		{TrackType: mcpmoqt.DataTrackTools, TrackName: "t", Data: "a"},
	})
	require.ErrorIs(t, err, mcpmoqt.ErrNoPublisher)

	err = h.BatchPublish(context.Background(), "t", []*mcpmoqt.DataTrackMessage{
		{TrackType: mcpmoqt.DataTrackTools, TrackName: "t", Data: "a"},
		{TrackType: mcpmoqt.DataTrackTools, TrackName: "t", Data: "b"},
	}, 1)
	require.ErrorIs(t, err, mcpmoqt.ErrNoPublisher)

	_ = slot
}
