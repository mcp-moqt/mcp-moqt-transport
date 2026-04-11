package mcpmoqt

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	Multiplier      float64
	Jitter          bool
	RetryableErrors []error
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

type RetryableFunc func() error

type Retry struct {
	config RetryConfig
	rand   *rand.Rand
	mu     sync.Mutex
}

func NewRetry(config RetryConfig) *Retry {
	return &Retry{
		config: config,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *Retry) Do(ctx context.Context, fn RetryableFunc) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !r.IsRetryable(err) {
			return err
		}

		if attempt >= r.config.MaxAttempts {
			break
		}

		delay := r.CalculateDelay(attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}

func (r *Retry) DoWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !r.IsRetryable(err) {
			return nil, err
		}

		if attempt >= r.config.MaxAttempts {
			break
		}

		delay := r.CalculateDelay(attempt)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, lastErr
}

func (r *Retry) CalculateDelay(attempt int) time.Duration {
	delay := float64(r.config.InitialDelay)
	for i := 1; i < attempt; i++ {
		delay *= r.config.Multiplier
	}

	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	if r.config.Jitter {
		r.mu.Lock()
		jitter := delay * 0.1 * r.rand.Float64()
		r.mu.Unlock()
		delay += jitter
	}

	return time.Duration(delay)
}

func (r *Retry) IsRetryable(err error) bool {
	if len(r.config.RetryableErrors) == 0 {
		return true
	}

	for _, retryableErr := range r.config.RetryableErrors {
		if err == retryableErr {
			return true
		}
	}

	return false
}

type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

func NewRetryableError(err error, retryable bool) *RetryableError {
	return &RetryableError{Err: err, Retryable: retryable}
}
