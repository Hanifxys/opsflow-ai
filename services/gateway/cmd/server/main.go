package main

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/opsflow/common/config"
	commonhttp "github.com/opsflow/common/httputil"
	"github.com/opsflow/common/logging"
	"github.com/opsflow/common/metrics"
	"github.com/opsflow/common/middleware"
	opsredis "github.com/opsflow/common/redis"
	"github.com/opsflow/common/telemetry"
	"github.com/opsflow/gateway/internal/proxy"
	searchhttp "github.com/opsflow/gateway/internal/search"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := logging.New(config.GetEnv("LOG_LEVEL", "info"))
	port := config.GetEnv("GATEWAY_PORT", "8080")
	jwtSecret := config.GetEnv("JWT_SECRET", "opsflow-dev-secret-key-change-in-prod")

	authPort := config.GetEnv("AUTH_SERVICE_PORT", "8081")
	incidentPort := config.GetEnv("INCIDENT_SERVICE_PORT", "8082")
	registryPort := config.GetEnv("REGISTRY_SERVICE_PORT", "8083")
	aiPort := config.GetEnv("AI_GATEWAY_PORT", "8084")

	redisHost := config.GetEnv("REDIS_HOST", "localhost")
	redisPort, _ := strconv.Atoi(config.GetEnv("REDIS_PORT", "6379"))
	redisPass := config.GetEnv("REDIS_PASSWORD", "")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	routes := []proxy.Route{
		{Prefix: "/api/v1/auth", Upstream: "http://localhost:" + authPort, Public: false},
		{Prefix: "/api/v1/incidents", Upstream: "http://localhost:" + incidentPort, Public: false},
		{Prefix: "/api/v1/services", Upstream: "http://localhost:" + registryPort, Public: false},
		{Prefix: "/api/v1/ai", Upstream: "http://localhost:" + aiPort, Public: false},
	}

	mainMux := http.NewServeMux()

	// Infrastructure probes & Prometheus metrics (unprotected)
	readyHandler, _ := commonhttp.ReadyHandler()
	mainMux.HandleFunc("/health", commonhttp.HealthHandler())
	mainMux.HandleFunc("/ready", readyHandler)
	mainMux.Handle("/metrics", metrics.Handler())

	// Unified Search API
	searchHandler := searchhttp.NewSearchHandler(authPort, incidentPort, registryPort)
	mainMux.Handle("/api/v1/search", http.HandlerFunc(searchHandler.Search))

	// Reverse proxy handler
	proxyHandler, err := proxy.NewGatewayHandler(routes, jwtSecret, nil)
	if err != nil {
		logger.Error("failed to create gateway handler", slog.String("error", err.Error()))
		return
	}
	mainMux.Handle("/api/v1/", proxyHandler)

	// Optional Redis client connection
	var rdb *redis.Client
	rdbClient, err := opsredis.NewClient(ctx, opsredis.Config{
		Host:     redisHost,
		Port:     redisPort,
		Password: redisPass,
	})
	if err != nil {
		logger.Warn("redis unavailable, using in-memory fallback rate limiter", slog.String("error", err.Error()))
	} else {
		rdb = rdbClient
		defer rdb.Close()
		logger.Info("redis connected for API rate limiting", slog.String("addr", rdbClient.Options().Addr))
	}

	// Middlewares
	var handler http.Handler = mainMux
	handler = metrics.MetricsMiddleware("gateway")(handler)
	handler = telemetry.TelemetryMiddleware("gateway")(handler)
	handler = middleware.RedisRateLimitMiddleware(rdb, 120, middleware.RateLimitMiddleware(120))(handler)
	handler = middleware.CORSMiddleware([]string{"*"})(handler)
	handler = commonhttp.RequestIDMiddleware(handler)
	handler = commonhttp.LoggingMiddleware(logger)(handler)
	handler = commonhttp.RecoverMiddleware(logger)(handler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("api-gateway starting", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down api-gateway")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
}
