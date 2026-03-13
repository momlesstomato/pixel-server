package user

import sdk "github.com/momlesstomato/pixel-sdk"

// Respected fires before a user respect event is persisted.
type Respected struct {
	sdk.BaseCancellable
	// ActorConnID stores the actor connection identifier.
	ActorConnID string
	// ActorUserID stores the actor user identifier.
	ActorUserID int
	// TargetUserID stores the target user identifier.
	TargetUserID int
}
