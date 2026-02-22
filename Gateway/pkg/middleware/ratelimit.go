package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/user/gateway/pkg/config"
	"github.com/user/gateway/pkg/redis"
)

func RateLimit(rdb *redis.Client, cfg config.RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			key := fmt.Sprintf("ratelimit:%s", r.RemoteAddr)

			// Simple fixed window rate limiting
			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if count == 1 {
				rdb.Expire(ctx, key, time.Second)
			}

			if count > int64(cfg.RequestsPerSecond) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
