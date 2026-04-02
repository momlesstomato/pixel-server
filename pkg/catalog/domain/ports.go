package domain

import "context"

// Spender is a secondary port for deducting credits and activity-point
// currencies during catalog purchase operations.
type Spender interface {
	// GetCredits resolves current credit balance for one user.
	GetCredits(ctx context.Context, userID int) (int, error)
	// AddCredits atomically adjusts credit balance by a signed amount and
	// returns the new balance. Pass a negative amount to deduct.
	AddCredits(ctx context.Context, userID int, amount int) (int, error)
	// GetCurrencyBalance resolves one activity-point balance for one user
	// and currency type identifier.
	GetCurrencyBalance(ctx context.Context, userID int, typeID int) (int, error)
	// AddCurrencyBalance atomically adjusts one activity-point balance by a
	// signed amount and returns the new balance. Pass a negative amount to deduct.
	AddCurrencyBalance(ctx context.Context, userID int, typeID int, amount int) (int, error)
}

// RecipientInfo carries the minimal user data required to validate a gift recipient.
type RecipientInfo struct {
	// UserID stores the recipient user identifier.
	UserID int
	// AllowGifts reports whether this user accepts incoming gifts.
	AllowGifts bool
}

// RecipientFinder is a secondary port for resolving gift recipients by username.
type RecipientFinder interface {
	// FindRecipientByUsername resolves recipient info by exact username.
	// Returns ErrRecipientNotFound when the username does not exist.
	FindRecipientByUsername(ctx context.Context, username string) (RecipientInfo, error)
}
