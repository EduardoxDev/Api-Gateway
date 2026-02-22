package proxy

import (
	"testing"
)

func TestRoundRobinBalancer(t *testing.T) {
	targets := []string{"t1", "t2", "t3"}
	lb := NewRoundRobinBalancer()

	// Initial
	if res := lb.Next(targets); res != "t1" {
		t.Errorf("Expected t1, got %s", res)
	}

	// Sequential
	if res := lb.Next(targets); res != "t2" {
		t.Errorf("Expected t2, got %s", res)
	}
	if res := lb.Next(targets); res != "t3" {
		t.Errorf("Expected t3, got %s", res)
	}

	// Wrap around
	if res := lb.Next(targets); res != "t1" {
		t.Errorf("Expected t1 after wrap, got %s", res)
	}
}

func TestRoundRobinBalancerEmpty(t *testing.T) {
	lb := NewRoundRobinBalancer()
	if res := lb.Next([]string{}); res != "" {
		t.Errorf("Expected empty string for empty targets, got %s", res)
	}
}
