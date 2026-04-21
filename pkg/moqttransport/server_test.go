package mcpmoqt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMOQTServerTransport(t *testing.T) {
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
			options: []Option{WithAddr("127.0.0.1:8080"), WithTLSServerConfig(nil)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := NewMOQTServerTransport(tt.options...)
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

func TestMOQTServerTransport_Close(t *testing.T) {
	t.Parallel()

	transport, err := NewMOQTServerTransport(WithAddr("127.0.0.1:8080"))
	require.NoError(t, err)
	require.NotNil(t, transport)

	err = transport.Close()
	require.NoError(t, err)
}
