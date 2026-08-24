package mcpmoqt

import (
	"errors"
	"testing"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransportError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *mcpmoqt.TransportError
		expected string
	}{
		{
			name: "with role",
			err: mcpmoqt.NewTransportErrorWithRole(
				"connect",
				"server",
				mcpmoqt.ErrInvalidAddr,
			),
			expected: "connect: server: invalid address",
		},
		{
			name: "without role",
			err: mcpmoqt.NewTransportError(
				"connect",
				mcpmoqt.ErrInvalidAddr,
			),
			expected: "connect: invalid address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestTransportError_Unwrap(t *testing.T) {
	innerErr := mcpmoqt.ErrInvalidAddr
	err := mcpmoqt.NewTransportError("connect", innerErr)

	assert.True(t, errors.Is(err, innerErr))
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "timeout error",
			err:      mcpmoqt.ErrTimeout,
			expected: true,
		},
		{
			name:     "not timeout error",
			err:      mcpmoqt.ErrInvalidAddr,
			expected: false,
		},
		{
			name:     "wrapped timeout error",
			err:      mcpmoqt.NewTransportError("op", mcpmoqt.ErrTimeout),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, mcpmoqt.IsTimeout(tt.err))
		})
	}
}

func TestIsConnectionClosed(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "connection closed error",
			err:      mcpmoqt.ErrConnectionClosed,
			expected: true,
		},
		{
			name:     "not connection closed error",
			err:      mcpmoqt.ErrInvalidAddr,
			expected: false,
		},
		{
			name:     "wrapped connection closed error",
			err:      mcpmoqt.NewTransportError("op", mcpmoqt.ErrConnectionClosed),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, mcpmoqt.IsConnectionClosed(tt.err))
		})
	}
}

func TestNewTransportError(t *testing.T) {
	err := mcpmoqt.NewTransportError("test-op", mcpmoqt.ErrInvalidAddr)
	require.NotNil(t, err)
	assert.Equal(t, "test-op", err.Op)
	assert.Equal(t, mcpmoqt.ErrInvalidAddr, err.Err)
	assert.Empty(t, err.Role)
}

func TestNewTransportErrorWithRole(t *testing.T) {
	err := mcpmoqt.NewTransportErrorWithRole("test-op", "server", mcpmoqt.ErrInvalidAddr)
	require.NotNil(t, err)
	assert.Equal(t, "test-op", err.Op)
	assert.Equal(t, "server", err.Role)
	assert.Equal(t, mcpmoqt.ErrInvalidAddr, err.Err)
}
