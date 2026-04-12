package mcpmoqt

import (
	"context"
	"fmt"

	"github.com/mengelbart/moqtransport"
	"github.com/mengelbart/moqtransport/quicmoq"
)

type MOQTServerTransport struct {
	cfg       *transportConfig
	sessionID string
}

func NewMOQTServerTransport(opts ...Option) (*MOQTServerTransport, error) {
	cfg, err := applyOptions(roleServer, opts)
	if err != nil {
		return nil, err
	}

	if cfg.tlsServer == nil {
		tlsCfg, err := defaultTLSConfig(roleServer, cfg.alpn)
		if err != nil {
			return nil, err
		}
		cfg.tlsServer = tlsCfg
	}

	return &MOQTServerTransport{cfg: cfg, sessionID: generateSessionID()}, nil
}

func (t *MOQTServerTransport) Connect(ctx context.Context) (Connection, error) {
	quicConn, ln, _, err := listenAndAcceptMOQT(ctx, t.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to listen and accept QUIC connection: %w", err)
	}

	conn := quicmoq.NewServer(quicConn)

	sendSlot := NewPublisherSlot()

	session := &moqtransport.Session{
		InitialMaxRequestID: DefaultMaxRequestID,
		Handler: &discoveryHandler{
			sessionID: t.sessionID,
		},
		SubscribeHandler: &subscribeHandler{
			sessionID: t.sessionID,
			sendTrack: trackServerToClient,
			sendSlot:  sendSlot,
		},
	}

	if err := runSession(ctx, session, conn); err != nil {
		ln.Close()
		return nil, fmt.Errorf("failed to run MOQT session: %w", err)
	}

	recv, err := session.Subscribe(ctx, controlNamespace(t.sessionID), trackClientToServer)
	if err != nil {
		ln.Close()
		return nil, fmt.Errorf("failed to subscribe to client-to-server control track: %w", err)
	}

	return newControlConn(conn, t.sessionID, recv, sendSlot), nil
}
