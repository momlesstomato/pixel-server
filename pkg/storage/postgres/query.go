package postgres

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"pixelsv/pkg/storage/interfaces"
)

// FetchOne maps a single query row into a typed value.
func FetchOne[T any](
	ctx context.Context,
	querier interfaces.RowQuerier,
	query string,
	args []any,
	mapper func(scanner interfaces.RowScanner) (T, error),
) (T, error) {
	if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
		log.Printf("level=debug component=postgres query=%q args=%v", query, args)
	}
	value, err := mapper(querier.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return value, interfaces.ErrNotFound
	}
	return value, err
}
