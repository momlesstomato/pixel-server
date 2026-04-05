package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// TicketSubmitting fires before a support ticket is submitted.
type TicketSubmitting struct {
	sdk.BaseCancellable
	// ReporterID stores the user submitting the ticket.
	ReporterID int
	// ReportedID stores the user being reported.
	ReportedID int
	// RoomID stores the room where the incident occurred.
	RoomID int
	// Category stores the ticket category.
	Category string
	// Message stores the ticket description.
	Message string
}
