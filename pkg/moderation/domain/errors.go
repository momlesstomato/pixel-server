package domain

import "errors"

// ErrActionNotFound indicates the moderation action was not found.
var ErrActionNotFound = errors.New("moderation action not found")

// ErrCannotDeleteHotelAction indicates hotel actions cannot be deleted.
var ErrCannotDeleteHotelAction = errors.New("hotel actions cannot be deleted")

// ErrAlreadyInactive indicates the action is already deactivated.
var ErrAlreadyInactive = errors.New("action is already inactive")

// ErrInvalidScope indicates an invalid action scope value.
var ErrInvalidScope = errors.New("invalid action scope")

// ErrMissingTarget indicates the target user is required.
var ErrMissingTarget = errors.New("target user is required")

// ErrTicketNotFound indicates the support ticket was not found.
var ErrTicketNotFound = errors.New("support ticket not found")

// ErrWordFilterNotFound indicates the word filter entry was not found.
var ErrWordFilterNotFound = errors.New("word filter entry not found")

// ErrPresetNotFound indicates the moderation preset was not found.
var ErrPresetNotFound = errors.New("moderation preset not found")

// ErrTicketAlreadyClosed indicates the ticket is already closed.
var ErrTicketAlreadyClosed = errors.New("ticket is already closed")
