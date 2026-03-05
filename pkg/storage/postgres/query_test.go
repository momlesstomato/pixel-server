package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"pixelsv/pkg/storage/interfaces"
)

type stubQuerier struct {
	row interfaces.RowScanner
}

func (s stubQuerier) QueryRow(context.Context, string, ...any) interfaces.RowScanner {
	return s.row
}

type stubScanner struct {
	scan func(dest ...any) error
}

func (s stubScanner) Scan(dest ...any) error {
	return s.scan(dest...)
}

// TestFetchOneNotFound maps pgx.ErrNoRows to interfaces.ErrNotFound.
func TestFetchOneNotFound(t *testing.T) {
	querier := stubQuerier{row: stubScanner{scan: func(...any) error { return pgx.ErrNoRows }}}
	_, err := FetchOne(context.Background(), querier, "select 1", nil, func(s interfaces.RowScanner) (int64, error) {
		var id int64
		return id, s.Scan(&id)
	})
	if !errors.Is(err, interfaces.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestFetchOneSuccess checks normal mapping path.
func TestFetchOneSuccess(t *testing.T) {
	row := stubScanner{scan: func(dest ...any) error {
		*dest[0].(*int64) = 9
		return nil
	}}
	querier := stubQuerier{row: row}
	value, err := FetchOne(context.Background(), querier, "select id", nil, func(s interfaces.RowScanner) (int64, error) {
		var id int64
		return id, s.Scan(&id)
	})
	if err != nil || value != 9 {
		t.Fatalf("unexpected result: value=%d err=%v", value, err)
	}
}
