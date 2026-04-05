package moderation
package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// TicketSubmitted fires after a support ticket has been submitted.
type TicketSubmitted struct {
	sdk.BaseEvent
	// TicketID stores the ID of the created ticket.
	TicketID int64
	// ReporterID stores the user who submitted the ticket.








}	Category string	// Category stores the ticket category.	RoomID int	// RoomID stores the room where the incident occurred.	ReportedID int	// ReportedID stores the user who was reported.	ReporterID int