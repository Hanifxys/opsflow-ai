package main

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	aihttp "github.com/opsflow/ai-gateway/internal/adapters/http"
	"github.com/opsflow/ai-gateway/internal/adapters/postgres"
	"github.com/opsflow/ai-gateway/internal/adapters/providers"
	"github.com/opsflow/ai-gateway/internal/adapters/tools"
	"github.com/opsflow/ai-gateway/internal/application"
	"github.com/opsflow/common/config"
	"github.com/opsflow/common/database"
	"github.com/opsflow/common/httputil"
	"github.com/opsflow/common/logging"
	"github.com/opsflow/common/metrics"
	"github.com/opsflow/common/telemetry"
)

func main() {
	logger := logging.New(config.GetEnv("LOG_LEVEL", "info"))
	port := config.GetEnv("AI_GATEWAY_PORT", "8084")
	jwtSecret := config.GetEnv("JWT_SECRET", "opsflow-dev-secret-key-change-in-prod")

	dbHost := config.GetEnv("POSTGRES_HOST", "localhost")
	dbPort, _ := strconv.Atoi(config.GetEnv("POSTGRES_PORT", "5432"))
	dbUser := config.GetEnv("POSTGRES_USER", "opsflow")
	dbPass := config.GetEnv("POSTGRES_PASSWORD", "opsflow_dev")
	dbName := config.GetEnv("POSTGRES_DB", "opsflow")

	dbCfg := database.PostgresConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPass,
		Database: dbName,
	}

	ollamaURL := config.GetEnv("OLLAMA_BASE_URL", "http://localhost:11434")
	cloudKey := config.GetEnv("OPENAI_API_KEY", "")
	cloudURL := config.GetEnv("OPENAI_BASE_URL", "https://api.openai.com/v1")
	defaultProvider := config.GetEnv("LLM_DEFAULT_PROVIDER", "mock")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()

	readyHandler, setReady := httputil.ReadyHandler()
	mux.HandleFunc("/health", httputil.HealthHandler())
	mux.HandleFunc("/ready", readyHandler)
	mux.Handle("/metrics", metrics.Handler())

	pool, err := database.NewPostgresPool(ctx, dbCfg)
	if err != nil {
		logger.Warn("database connection unavailable, running in degraded mode", slog.String("error", err.Error()))
		setReady(false)
	} else {
		defer pool.Close()

		migrationsDir := config.GetEnv("MIGRATIONS_DIR", "migrations")
		if err := database.RunMigrations(migrationsDir, dbCfg.DSN()); err != nil {
			logger.Warn("migrations warning", slog.String("error", err.Error()))
		}

		repo := postgres.NewAIRepository(pool)

		// Setup LLM Providers & Router
		mockP := providers.NewMockProvider()
		ollamaP := providers.NewOllamaProvider(ollamaURL)
		cloudP := providers.NewCloudProvider(cloudKey, cloudURL)

		router := application.NewModelRouter(defaultProvider, mockP, ollamaP, cloudP)
		toolReg := tools.NewToolRegistry()

		service := application.NewAIService(repo, router, toolReg)
		handler := aihttp.NewAIHandler(service)

		aihttp.RegisterRoutes(mux, handler, jwtSecret)
	}

	var handler http.Handler = mux
	handler = metrics.MetricsMiddleware("ai-gateway")(handler)
	handler = telemetry.TelemetryMiddleware("ai-gateway")(handler)
	handler = httputil.RequestIDMiddleware(handler)
	handler = httputil.LoggingMiddleware(logger)(handler)
	handler = httputil.RecoverMiddleware(logger)(handler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("ai-gateway starting", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down ai-gateway")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
}
