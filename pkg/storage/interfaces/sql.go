package interfaces

import "context"

// RowScanner scans query results into destination pointers.
type RowScanner interface {
	// Scan copies current row columns into destination pointers.
	Scan(dest ...any) error
}

// RowQuerier executes single-row queries.
type RowQuerier interface {
	// QueryRow executes a query expected to return one row.
	QueryRow(ctx context.Context, query string, args ...any) RowScanner
}
