package mcpmoqt

import "fmt"

// Role describes whether the transport acts as a server or client.
type Role int

const (
	RoleServer Role = iota
	RoleClient
)

// NewMoqTransport is the unified constructor for MCP over MOQT.
func NewMoqTransport(role Role, opts ...Option) (Transport, error) {
	switch role {
	case RoleServer:
		return NewMOQTServerTransport(opts...)
	case RoleClient:
		return NewMOQTClientTransport(opts...)
	default:
		return nil, fmt.Errorf("unknown role: %d", role)
	}
}
