package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// TicketClosed fires after a support ticket has been closed.
type TicketClosed struct {
	sdk.BaseEvent
	// TicketID stores the ticket that was closed.
	TicketID int64
	// ClosedBy stores the staff member who closed the ticket.
	ClosedBy int
}
