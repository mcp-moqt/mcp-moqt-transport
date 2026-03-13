package mcpmoqt

import (
	"context"
	"fmt"
	"net"

	"github.com/quic-go/quic-go"
)

// listenAndAcceptMOQT listens for QUIC connections and accepts the first one.
func listenAndAcceptMOQT(ctx context.Context, cfg *transportConfig) (*quic.Conn, *quic.Listener, net.Addr, error) {
	// Create QUIC listener
	ln, err := quic.ListenAddr(cfg.addr, cfg.tlsServer, cfg.quicConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to listen: %w", err)
	}

	// Accept connection
	conn, err := ln.Accept(ctx)
	if err != nil {
		ln.Close()
		return nil, nil, nil, fmt.Errorf("failed to accept: %w", err)
	}

	return conn, ln, ln.Addr(), nil
}

// dialMOQT dials a QUIC connection and creates a MOQT connection.
func dialMOQT(ctx context.Context, cfg *transportConfig) (*quic.Conn, net.Addr, error) {
	// Dial QUIC connection
	conn, err := quic.DialAddr(ctx, cfg.addr, cfg.tlsClient, cfg.quicConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial: %w", err)
	}

	return conn, conn.RemoteAddr(), nil
}
