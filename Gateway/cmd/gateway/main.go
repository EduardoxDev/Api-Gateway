package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/user/gateway/pkg/breaker"
	"github.com/user/gateway/pkg/config"
	"github.com/user/gateway/pkg/logger"
	"github.com/user/gateway/pkg/middleware"
	"github.com/user/gateway/pkg/proxy"
	"github.com/user/gateway/pkg/redis"
)

func main() {
	// Initialize logger
	_ = logger.Init()
	slog.Info("Starting API Gateway...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize Redis
	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		slog.Warn("Redis connection failed, rate limiting will be disabled", "error", err)
	}

	// Initialize Circuit Breaker
	cb := breaker.NewCircuitBreaker(5, 30*time.Second)

	// Create request multiplexer
	mux := http.NewServeMux()

	// Initialize Load Balancer
	lb := proxy.NewRoundRobinBalancer()

	// Setup routes
	for _, route := range cfg.Gateway.Routes {
		if len(route.Targets) == 0 {
			slog.Error("No targets specified for route", "path", route.Path)
			continue
		}

		// Pre-create proxies for each target
		proxies := make([]*proxy.Proxy, len(route.Targets))
		for i, target := range route.Targets {
			p, err := proxy.NewProxy(target)
			if err != nil {
				slog.Error("Failed to create proxy for target", "target", target, "error", err)
				continue
			}
			proxies[i] = p
		}

		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if method is allowed
			methodAllowed := false
			for _, m := range route.Methods {
				if m == r.Method {
					methodAllowed = true
					break
				}
			}
			if !methodAllowed {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Load Balancing
			target := lb.Next(route.Targets)
			// Find the corresponding proxy
			var selectedProxy *proxy.Proxy
			for _, p := range proxies {
				if p != nil && strings.Contains(target, p.TargetHost()) { // Added TargetHost helper
					selectedProxy = p
					break
				}
			}

			if selectedProxy != nil {
				selectedProxy.ServeHTTP(w, r, route.Path)
			} else {
				http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			}
		})

		// Apply middleware in REVERSE order (innermost first)
		for i := len(route.Middleware) - 1; i >= 0; i-- {
			switch route.Middleware[i] {
			case "auth":
				handler = middleware.Auth(cfg.Auth)(handler)
			case "ratelimit":
				if rdb != nil {
					handler = middleware.RateLimit(rdb, cfg.RateLimit)(handler)
				}
			}
		}

		// Always apply common middleware
		handler = middleware.CircuitBreaker(cb)(handler)
		handler = middleware.Logger(handler)

		mux.Handle(route.Path+"/", handler)
		slog.Info("Registered route", "path", route.Path, "targets", route.Targets)
	}

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Server configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	go func() {
		slog.Info("Gateway server listening", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	slog.Info("Shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}

	slog.Info("Gateway stopped gracefully")
}
