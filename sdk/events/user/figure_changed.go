package user

import sdk "github.com/momlesstomato/pixel-sdk"

// FigureChanged fires before a user figure update is persisted.
type FigureChanged struct {
	sdk.BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the user identifier.
	UserID int
	// OldFigure stores previous figure value.
	OldFigure string
	// NewFigure stores requested figure value.
	NewFigure string
	// Gender stores requested gender value.
	Gender string
}
