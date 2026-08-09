package main

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/opsflow/auth-service/internal/adapters/bcrypt"
	authhttp "github.com/opsflow/auth-service/internal/adapters/http"
	"github.com/opsflow/auth-service/internal/adapters/postgres"
	"github.com/opsflow/auth-service/internal/application"
	"github.com/opsflow/common/config"
	"github.com/opsflow/common/database"
	"github.com/opsflow/common/httputil"
	"github.com/opsflow/common/logging"
	"github.com/opsflow/common/metrics"
	"github.com/opsflow/common/telemetry"
)

func main() {
	logger := logging.New(config.GetEnv("LOG_LEVEL", "info"))
	port := config.GetEnv("AUTH_SERVICE_PORT", "8081")
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

		userRepo := postgres.NewUserRepository(pool)
		hasher := bcrypt.New(10)
		authService := application.NewAuthService(userRepo, hasher, jwtSecret, 15*time.Minute, 7*24*time.Hour)
		authHandler := authhttp.NewAuthHandler(authService)

		authhttp.RegisterRoutes(mux, authHandler, jwtSecret)
	}

	var handler http.Handler = mux
	handler = metrics.MetricsMiddleware("auth-service")(handler)
	handler = telemetry.TelemetryMiddleware("auth-service")(handler)
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
		logger.Info("auth-service starting", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down auth-service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
}
