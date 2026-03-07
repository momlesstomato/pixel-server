package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
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
	if debugLoggingEnabled() {
		poolCfg.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   postgresTraceLogger{},
			LogLevel: tracelog.LogLevelDebug,
		}
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return &Service{pool: pool}, nil
}

// debugLoggingEnabled reports whether runtime debug logging is enabled.
func debugLoggingEnabled() bool {
	return strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug")
}

// postgresTraceLogger bridges pgx tracelog output into stdout console logs.
type postgresTraceLogger struct{}

// Log writes one pgx trace event.
func (l postgresTraceLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	log.Printf("level=debug component=postgres msg=%q data=%v", msg, data)
	_ = ctx
	_ = level
	_ = l
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
