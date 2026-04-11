package mcpmoqt

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidRole       = errors.New("invalid transport role")
	ErrInvalidAddr       = errors.New("invalid address")
	ErrInvalidALPN       = errors.New("invalid ALPN")
	ErrConnectionClosed  = errors.New("connection closed")
	ErrSessionNotFound   = errors.New("session not found")
	ErrTrackNotFound     = errors.New("track not found")
	ErrTimeout           = errors.New("operation timed out")
	ErrTLSConfig         = errors.New("TLS configuration error")
	ErrQUICConfig        = errors.New("QUIC configuration error")
	ErrDiscoveryFailed   = errors.New("discovery failed")
	ErrSubscribeFailed   = errors.New("subscribe failed")
	ErrNoPublisher       = errors.New("no publisher available for data track")
	ErrNoSubscriber      = errors.New("no subscriber registered for data track")
	ErrInvalidDataTrack  = errors.New("invalid data track type")
	ErrAckTimeout        = errors.New("acknowledgment timeout")
	ErrHeartbeatFailed   = errors.New("heartbeat failed")
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
	ErrNotRetryable      = errors.New("error is not retryable")
)

type TransportError struct {
	Op   string
	Err  error
	Role string
}

func (e *TransportError) Error() string {
	if e.Role != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Role, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *TransportError) Unwrap() error {
	return e.Err
}

func NewTransportError(op string, err error) *TransportError {
	return &TransportError{
		Op:  op,
		Err: err,
	}
}

func NewTransportErrorWithRole(op string, role string, err error) *TransportError {
	return &TransportError{
		Op:   op,
		Err:  err,
		Role: role,
	}
}

func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

func IsConnectionClosed(err error) bool {
	return errors.Is(err, ErrConnectionClosed)
}

func IsTemporary(err error) bool {
	var tErr *TransportError
	if ok := errors.As(err, &tErr); ok {
		return tErr.Role == "temporary"
	}
	return false
}
