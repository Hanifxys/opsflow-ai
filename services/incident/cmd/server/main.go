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
	"github.com/opsflow/common/database"
	"github.com/opsflow/common/httputil"
	"github.com/opsflow/common/logging"
	"github.com/opsflow/common/metrics"
	"github.com/opsflow/common/outbox"
	"github.com/opsflow/common/rabbitmq"
	"github.com/opsflow/common/telemetry"
	inchttp "github.com/opsflow/incident-service/internal/adapters/http"
	"github.com/opsflow/incident-service/internal/adapters/postgres"
	"github.com/opsflow/incident-service/internal/application"
)

func main() {
	logger := logging.New(config.GetEnv("LOG_LEVEL", "info"))
	port := config.GetEnv("INCIDENT_SERVICE_PORT", "8082")
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

	rbHost := config.GetEnv("RABBITMQ_HOST", "localhost")
	rbPort, _ := strconv.Atoi(config.GetEnv("RABBITMQ_PORT", "5672"))
	rbUser := config.GetEnv("RABBITMQ_USER", "opsflow")
	rbPass := config.GetEnv("RABBITMQ_PASSWORD", "opsflow_dev")

	rbCfg := rabbitmq.Config{
		Host:     rbHost,
		Port:     rbPort,
		User:     rbUser,
		Password: rbPass,
	}

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

		repo := postgres.NewIncidentRepository(pool)
		service := application.NewIncidentService(repo)
		handler := inchttp.NewIncidentHandler(service)

		inchttp.RegisterRoutes(mux, handler, jwtSecret)

		rabbitClient, err := rabbitmq.NewClient(rbCfg)
		if err != nil {
			logger.Warn("rabbitmq unavailable, outbox publishing paused", slog.String("error", err.Error()))
		} else {
			defer rabbitClient.Close()
			pub := outbox.NewPublisher(pool, rabbitClient, logger)
			pub.Start(2 * time.Second)
			defer pub.Stop()
			logger.Info("outbox publisher started")
		}
	}

	var handler http.Handler = mux
	handler = metrics.MetricsMiddleware("incident-service")(handler)
	handler = telemetry.TelemetryMiddleware("incident-service")(handler)
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
		logger.Info("incident-service starting", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down incident-service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
}
