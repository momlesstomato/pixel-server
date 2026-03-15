package domain

import "errors"

// ErrGroupNotFound defines missing group lookup behavior.
var ErrGroupNotFound = errors.New("permission group not found")

// ErrDefaultGroupRequired defines missing default group behavior.
var ErrDefaultGroupRequired = errors.New("default permission group is required")

// ErrGroupInUse defines delete-protected group behavior.
var ErrGroupInUse = errors.New("permission group has assigned users")

// ErrCannotDeleteDefaultGroup defines protected default-group delete behavior.
var ErrCannotDeleteDefaultGroup = errors.New("default permission group cannot be deleted")
