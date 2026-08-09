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
	"github.com/opsflow/common/rabbitmq"
	"github.com/opsflow/common/telemetry"
	notifamqp "github.com/opsflow/notification-service/internal/adapters/amqp"
	"github.com/opsflow/notification-service/internal/adapters/postgres"
	"github.com/opsflow/notification-service/internal/application"
)

func main() {
	logger := logging.New(config.GetEnv("LOG_LEVEL", "info"))
	port := config.GetEnv("NOTIFICATION_SERVICE_PORT", "8085")

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

		repo := postgres.NewNotificationRepository(pool)
		service := application.NewNotificationService(repo)

		rabbitClient, err := rabbitmq.NewClient(rbCfg)
		if err != nil {
			logger.Warn("rabbitmq connection unavailable, worker paused", slog.String("error", err.Error()))
			setReady(false)
		} else {
			defer rabbitClient.Close()
			consumer := notifamqp.NewNotificationConsumer(rabbitClient, service, logger)
			if err := consumer.Start(); err != nil {
				logger.Error("failed to start notification consumer", slog.String("error", err.Error()))
			} else {
				defer consumer.Stop()
				logger.Info("notification worker consumer running")
			}
		}
	}

	var handler http.Handler = mux
	handler = metrics.MetricsMiddleware("notification-service")(handler)
	handler = telemetry.TelemetryMiddleware("notification-service")(handler)
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
		logger.Info("notification-service starting", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down notification-service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
}
