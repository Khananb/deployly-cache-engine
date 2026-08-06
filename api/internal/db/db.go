package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Client wraps the pgxpool.Pool to provide database access
type Client struct {
	Pool *pgxpool.Pool
}

// NewClient initializes a new PostgreSQL connection pool
func NewClient(ctx context.Context) (*Client, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Default for local development via docker-compose
		connStr = "postgres://deployly_admin:password_change_me@localhost:5432/deployly_cache?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DATABASE_URL: %w", err)
	}

	// VPS-optimized pool settings
	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Client{Pool: pool}, nil
}

// Close gracefully shuts down the connection pool
func (c *Client) Close() {
	c.Pool.Close()
}
