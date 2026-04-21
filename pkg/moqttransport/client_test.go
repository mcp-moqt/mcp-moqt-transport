package mcpmoqt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewMOQTClientTransport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []Option
		wantErr bool
	}{
		{
			name:    "default options",
			options: []Option{WithAddr("127.0.0.1:8080")},
			wantErr: false,
		},
		{
			name:    "with custom TLS config",
			options: []Option{WithAddr("127.0.0.1:8080"), WithTLSClientConfig(nil)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := NewMOQTClientTransport(tt.options...)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, transport)
				defer transport.Close()
			}
		})
	}
}

func TestMOQTClientTransport_Close(t *testing.T) {
	t.Parallel()

	transport, err := NewMOQTClientTransport(WithAddr("127.0.0.1:8080"))
	require.NoError(t, err)
	require.NotNil(t, transport)

	err = transport.Close()
	require.NoError(t, err)
}

func TestMOQTClientTransport_Connect_Timeout(t *testing.T) {
	t.Parallel()

	transport, err := NewMOQTClientTransport(WithAddr("127.0.0.1:9999"))
	require.NoError(t, err)
	defer transport.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Should timeout since no server is listening
	_, err = transport.Connect(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "deadline exceeded")
}
