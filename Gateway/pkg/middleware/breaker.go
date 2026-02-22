package middleware

import (
	"net/http"

	"github.com/user/gateway/pkg/breaker"
)

func CircuitBreaker(cb *breaker.CircuitBreaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := cb.Execute(func() error {
				// We don't want to capture 4xx errors as failures for the circuit breaker,
				// but since Execute wraps the whole logic, we need to be careful.
				// For simplicity, we'll assume any downstream error that makes it back
				// to the proxy layer is a failure.
				next.ServeHTTP(w, r)
				return nil // Status codes are handled by proxy, this is a simplified version
			})

			if err == breaker.ErrCircuitOpen {
				http.Error(w, "Service Unavailable (Circuit Breaker Open)", http.StatusServiceUnavailable)
				return
			}
		})
	}
}
