package workers

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", config.MaxAttempts)
	}
	if config.InitialInterval != 2*time.Second {
		t.Errorf("InitialInterval = %v, want 2s", config.InitialInterval)
	}
	if config.MaxInterval != 30*time.Second {
		t.Errorf("MaxInterval = %v, want 30s", config.MaxInterval)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Multiplier = %v, want 2.0", config.Multiplier)
	}
}

func TestRetryConfig_CalculateBackoff(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 2 * time.Second},  // First attempt
		{1, 2 * time.Second},  // Second attempt
		{2, 4 * time.Second},  // Third attempt (2 * 2)
		{3, 8 * time.Second},  // Fourth attempt (4 * 2)
		{4, 16 * time.Second}, // Fifth attempt (8 * 2)
		{5, 30 * time.Second}, // Sixth attempt (capped at max)
		{6, 30 * time.Second}, // Seventh attempt (still capped)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := config.CalculateBackoff(tt.attempt)
			if result != tt.expected {
				t.Errorf("CalculateBackoff(%d) = %v, want %v", tt.attempt, result, tt.expected)
			}
		})
	}
}

func TestDefaultManagerConfig(t *testing.T) {
	config := DefaultManagerConfig()

	if config.HeartbeatInterval != 30*time.Second {
		t.Errorf("HeartbeatInterval = %v, want 30s", config.HeartbeatInterval)
	}
	if config.ZombieTimeout != 2*time.Minute {
		t.Errorf("ZombieTimeout = %v, want 2m", config.ZombieTimeout)
	}
	if config.RetryConfig.MaxAttempts != 3 {
		t.Errorf("RetryConfig.MaxAttempts = %d, want 3", config.RetryConfig.MaxAttempts)
	}
}

func TestWorkerError(t *testing.T) {
	originalErr := errors.New("something went wrong")

	t.Run("retryable error", func(t *testing.T) {
		err := NewRetryableError(originalErr)
		if err.Type != ErrorTypeRetryable {
			t.Errorf("Type = %v, want ErrorTypeRetryable", err.Type)
		}
		if err.Error() != originalErr.Error() {
			t.Errorf("Error() = %v, want %v", err.Error(), originalErr.Error())
		}
		if !errors.Is(err, originalErr) {
			t.Error("Unwrap should return original error")
		}
		if !IsRetryable(err) {
			t.Error("IsRetryable should return true for retryable errors")
		}
	})

	t.Run("fatal error", func(t *testing.T) {
		err := NewFatalError(originalErr)
		if err.Type != ErrorTypeFatal {
			t.Errorf("Type = %v, want ErrorTypeFatal", err.Type)
		}
		if err.Error() != originalErr.Error() {
			t.Errorf("Error() = %v, want %v", err.Error(), originalErr.Error())
		}
		if !errors.Is(err, originalErr) {
			t.Error("Unwrap should return original error")
		}
		if IsRetryable(err) {
			t.Error("IsRetryable should return false for fatal errors")
		}
	})

	t.Run("unknown error is retryable by default", func(t *testing.T) {
		if !IsRetryable(originalErr) {
			t.Error("IsRetryable should return true for unknown errors")
		}
	})
}

func TestErrorType_Constants(t *testing.T) {
	if ErrorTypeRetryable != 0 {
		t.Errorf("ErrorTypeRetryable = %v, want 0", ErrorTypeRetryable)
	}
	if ErrorTypeFatal != 1 {
		t.Errorf("ErrorTypeFatal = %v, want 1", ErrorTypeFatal)
	}
}

func TestRetryConfig_CustomValues(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:     5,
		InitialInterval: 1 * time.Second,
		MaxInterval:     10 * time.Second,
		Multiplier:      3.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},  // First attempt
		{1, 1 * time.Second},  // Second attempt
		{2, 3 * time.Second},  // Third attempt (1 * 3)
		{3, 9 * time.Second},  // Fourth attempt (3 * 3)
		{4, 10 * time.Second}, // Fifth attempt (capped at max)
	}

	for _, tt := range tests {
		result := config.CalculateBackoff(tt.attempt)
		if result != tt.expected {
			t.Errorf("CalculateBackoff(%d) = %v, want %v", tt.attempt, result, tt.expected)
		}
	}
}

// Note: Integration tests for Manager require running Redis
// These would test actual job processing with mock workers
