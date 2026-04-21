package mcpmoqt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestControlNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sessionID string
		want      []string
	}{
		{
			name:      "empty session ID",
			sessionID: "",
			want:      []string{"mcp", "", "control"},
		},
		{
			name:      "non-empty session ID",
			sessionID: "test-session",
			want:      []string{"mcp", "test-session", "control"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controlNamespace(tt.sessionID)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestNewControlConn(t *testing.T) {
	t.Parallel()

	// This test would require mocking the moqtransport.Connection and RemoteTrack
	// For now, we'll just test that the function doesn't panic
	// In a real test, we would use mocks to test the full functionality
	slot := NewPublisherSlot()
	require.NotNil(t, slot)

	// We can't create a real controlConn without a real moqtransport.Connection
	// So we'll just test that the PublisherSlot works with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := slot.Get(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "deadline exceeded")
}
