package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConfig holds options for connecting to PostgreSQL.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// DSN returns a PostgreSQL connection string.
func (c PostgresConfig) DSN() string {
	ssl := c.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, ssl)
}

// NewPostgresPool initializes a pgxpool connection pool.
func NewPostgresPool(ctx context.Context, cfg PostgresConfig) (*pgxpool.Pool, error) {
	connConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("invalid postgres config: %w", err)
	}

	connConfig.MaxConns = 25
	connConfig.MinConns = 5
	connConfig.MaxConnLifetime = 1 * time.Hour
	connConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create postgres pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping postgres: %w", err)
	}

	return pool, nil
}
