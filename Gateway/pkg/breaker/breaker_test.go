package breaker

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	// Closed state
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// First failure
	_ = cb.Execute(func() error { return errors.New("fail") })
	if cb.state != StateClosed {
		t.Errorf("Expected state to be Closed after 1 failure, got %v", cb.state)
	}

	// Second failure -> Open
	_ = cb.Execute(func() error { return errors.New("fail") })
	if cb.state != StateOpen {
		t.Errorf("Expected state to be Open after 2 failures, got %v", cb.state)
	}

	// Should be blocked
	err = cb.Execute(func() error { return nil })
	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}

	// Wait for timeout -> Half-Open (allows one request)
	time.Sleep(150 * time.Millisecond)
	err = cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("Expected nil error after timeout, got %v", err)
	}
	if cb.state != StateClosed {
		t.Errorf("Expected state to be Closed after success in Half-Open, got %v", cb.state)
	}
}
