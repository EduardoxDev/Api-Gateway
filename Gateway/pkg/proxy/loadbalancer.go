package proxy

import (
	"sync/atomic"
)

type LoadBalancer interface {
	Next(targets []string) string
}

type RoundRobinBalancer struct {
	current uint32
}

func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

func (rr *RoundRobinBalancer) Next(targets []string) string {
	if len(targets) == 0 {
		return ""
	}
	n := atomic.AddUint32(&rr.current, 1)
	return targets[(int(n)-1)%len(targets)]
}
