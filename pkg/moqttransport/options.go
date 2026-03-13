package mcpmoqt

import (
	"crypto/tls"
	"errors"

	"github.com/mcp-moqt/mcp-moqt-transport/pkg/config"
	"github.com/quic-go/quic-go"
)

// transportRole defines the role of the transport (server or client).
type transportRole int

const (
	// roleServer indicates the transport is acting as a server.
	roleServer transportRole = iota + 1
	// roleClient indicates the transport is acting as a client.
	roleClient
)

// transportConfig contains configuration options for the MOQT transport.
type transportConfig struct {
	// role is the role of the transport (server or client).
	role transportRole

	// addr is the address to listen on (server) or connect to (client).
	addr string

	// tlsServer is the TLS config used for QUIC listening (server side).
	tlsServer *tls.Config
	// tlsClient is the TLS config used for QUIC dialing (client side).
	tlsClient *tls.Config

	// quicConfig is the QUIC config for both client and server.
	quicConfig *quic.Config
	// alpn is the ALPN protocol for TLS.
	alpn []string
}

// Option is a function that configures the transport.
type Option func(*transportConfig) error

// defaultALPN returns the default ALPN protocol for MOQT.
// Defaults to "moq-00" for draft-11 compatibility.
func defaultALPN() []string { return []string{"moq-00"} }

// WithAddr sets the address to listen on (server) or connect to (client).
// Example: "127.0.0.1:8080"
func WithAddr(addr string) Option {
	return func(c *transportConfig) error {
		c.addr = addr
		return nil
	}
}

// WithQUICConfig sets the QUIC config for both client and server.
// Use this to customize QUIC behavior, such as enabling datagrams or setting timeouts.
func WithQUICConfig(cfg *quic.Config) Option {
	return func(c *transportConfig) error {
		c.quicConfig = cfg
		return nil
	}
}

// WithALPN overrides the default ALPN ("moq-00").
// Use this to specify a different ALPN protocol for TLS.
func WithALPN(alpn ...string) Option {
	return func(c *transportConfig) error {
		if len(alpn) == 0 {
			return errors.New("alpn must not be empty")
		}
		c.alpn = append([]string(nil), alpn...)
		return nil
	}
}

// WithTLSServerConfig sets the TLS config used for QUIC listening (server side).
// If not provided, a default self-signed certificate is used for local development.
func WithTLSServerConfig(cfg *tls.Config) Option {
	return func(c *transportConfig) error {
		c.tlsServer = cfg
		return nil
	}
}

// WithTLSClientConfig sets the TLS config used for QUIC dialing (client side).
// If not provided, a default config with InsecureSkipVerify enabled is used for local development.
func WithTLSClientConfig(cfg *tls.Config) Option {
	return func(c *transportConfig) error {
		c.tlsClient = cfg
		return nil
	}
}

// WithConfig loads options from a config.Config object.
func WithConfig(cfg *config.Config) Option {
	return func(c *transportConfig) error {
		if err := cfg.Validate(); err != nil {
			return err
		}
		
		c.addr = cfg.Addr
		c.alpn = cfg.ALPN
		c.quicConfig = &quic.Config{EnableDatagrams: cfg.EnableDatagrams}
		
		return nil
	}
}

// WithConfigFromEnv loads options from environment variables.
func WithConfigFromEnv() Option {
	return WithConfig(config.LoadFromEnv())
}

// applyOptions applies the provided options to the transport config.
// It sets default values for any options not provided.
//
// Defaults:
// - addr: "localhost:0" (random port)
// - ALPN: ["moq-00"]
// - QUIC: Datagrams enabled
func applyOptions(role transportRole, opts []Option) (*transportConfig, error) {
	// Defaults: addr localhost:0, ALPN moq-00, QUIC datagrams enabled.
	cfg := &transportConfig{
		role:       role,
		addr:       "localhost:0",
		quicConfig: &quic.Config{EnableDatagrams: true},
		alpn:       defaultALPN(),
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	if cfg.addr == "" {
		return nil, errors.New("addr must not be empty")
	}
	return cfg, nil
}
