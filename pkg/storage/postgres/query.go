package postgres

import (
	"context"
	"errors"

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
	value, err := mapper(querier.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return value, interfaces.ErrNotFound
	}
	return value, err
}
