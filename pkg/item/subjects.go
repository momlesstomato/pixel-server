package item

// NATS subjects for the item/inventory realm.
const (
	// SubjInventoryUpdated signals item grants. Format: userID.
	SubjInventoryUpdated = "inventory.updated.%d"
)
