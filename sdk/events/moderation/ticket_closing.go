package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// TicketClosing fires before a support ticket is closed.
type TicketClosing struct {
	sdk.BaseCancellable
	// TicketID stores the ticket being closed.
	TicketID int64
	// ClosedBy stores the staff member closing the ticket.
	ClosedBy int
}
