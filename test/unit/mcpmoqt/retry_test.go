package mcpmoqt

import (
	"context"
	"errors"
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := mcpmoqt.DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.True(t, config.Jitter)
}

func TestNewRetry(t *testing.T) {
	config := mcpmoqt.DefaultRetryConfig()
	retry := mcpmoqt.NewRetry(config)

	require.NotNil(t, retry)
}

func TestRetry_Do_Success(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	attempts := 0
	err := retry.Do(context.Background(), func() error {
		attempts++
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestRetry_Do_SuccessAfterRetry(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	attempts := 0
	err := retry.Do(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetry_Do_MaxAttemptsExceeded(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	attempts := 0
	testErr := errors.New("persistent error")
	err := retry.Do(context.Background(), func() error {
		attempts++
		return testErr
	})

	require.Error(t, err)
	assert.Equal(t, testErr, err)
	assert.Equal(t, 3, attempts)
}

func TestRetry_Do_ContextCancellation(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	attempts := 0
	err := retry.Do(ctx, func() error {
		attempts++
		return errors.New("error")
	})

	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Equal(t, 1, attempts)
}

func TestRetry_DoWithResult_Success(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	attempts := 0
	result, err := retry.DoWithResult(context.Background(), func() (interface{}, error) {
		attempts++
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 1, attempts)
}

func TestRetry_DoWithResult_SuccessAfterRetry(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	attempts := 0
	result, err := retry.DoWithResult(context.Background(), func() (interface{}, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("temporary error")
		}
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, attempts)
}

func TestRetry_DoWithResult_MaxAttemptsExceeded(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	attempts := 0
	testErr := errors.New("persistent error")
	result, err := retry.DoWithResult(context.Background(), func() (interface{}, error) {
		attempts++
		return nil, testErr
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, testErr, err)
	assert.Equal(t, 3, attempts)
}

func TestRetry_CalculateDelay(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 800 * time.Millisecond},
		{5, 1 * time.Second},
		{6, 1 * time.Second},
	}

	for _, tt := range tests {
		delay := retry.CalculateDelay(tt.attempt)
		assert.Equal(t, tt.expected, delay, "attempt %d", tt.attempt)
	}
}

func TestRetry_IsRetryable_AllErrors(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        1 * time.Second,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: nil,
	}
	retry := mcpmoqt.NewRetry(config)

	assert.True(t, retry.IsRetryable(errors.New("any error")))
}

func TestRetry_IsRetryable_SpecificErrors(t *testing.T) {
	customErr := errors.New("custom error")
	config := mcpmoqt.RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        1 * time.Second,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []error{customErr},
	}
	retry := mcpmoqt.NewRetry(config)

	assert.True(t, retry.IsRetryable(customErr))
	assert.False(t, retry.IsRetryable(errors.New("other error")))
}

func TestRetryableError(t *testing.T) {
	innerErr := errors.New("inner error")
	retryableErr := mcpmoqt.NewRetryableError(innerErr, true)

	assert.Equal(t, innerErr.Error(), retryableErr.Error())
	assert.Equal(t, innerErr, errors.Unwrap(retryableErr))
	assert.True(t, retryableErr.Retryable)
}

func TestRetryableError_NotRetryable(t *testing.T) {
	innerErr := errors.New("inner error")
	retryableErr := mcpmoqt.NewRetryableError(innerErr, false)

	assert.False(t, retryableErr.Retryable)
}

func TestRetry_DoWithNonRetryableError(t *testing.T) {
	customErr := errors.New("retryable error")
	nonRetryableErr := errors.New("non-retryable error")
	config := mcpmoqt.RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        1 * time.Second,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []error{customErr},
	}
	retry := mcpmoqt.NewRetry(config)

	attempts := 0
	err := retry.Do(context.Background(), func() error {
		attempts++
		return nonRetryableErr
	})

	require.Error(t, err)
	assert.Equal(t, nonRetryableErr, err)
	assert.Equal(t, 1, attempts)
}

func TestRetry_ConcurrentRetries(t *testing.T) {
	config := mcpmoqt.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	retry := mcpmoqt.NewRetry(config)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			attempts := 0
			_ = retry.Do(context.Background(), func() error {
				attempts++
				if attempts < 2 {
					return errors.New("error")
				}
				return nil
			})
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
