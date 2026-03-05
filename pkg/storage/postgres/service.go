package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service owns a PostgreSQL connection pool.
type Service struct {
	pool *pgxpool.Pool
}

// New creates a Service from Config.
func New(ctx context.Context, cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConns = cfg.MaxConns
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return &Service{pool: pool}, nil
}

// Ping checks database connectivity.
func (s *Service) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Close releases the connection pool.
func (s *Service) Close() {
	s.pool.Close()
}

// Pool returns the underlying pgx connection pool.
func (s *Service) Pool() *pgxpool.Pool {
	return s.pool
}
