package proxy

import (
	"testing"
)

func BenchmarkRoundRobinBalancer(b *testing.B) {
	targets := []string{"t1", "t2", "t3", "t4", "t5"}
	lb := NewRoundRobinBalancer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lb.Next(targets)
	}
}
