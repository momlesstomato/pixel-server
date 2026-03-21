package domain

import "time"

// TradeLog defines one completed trade audit row.
type TradeLog struct {
	// ID stores stable trade log identifier.
	ID int
	// UserOneID stores the first trading user identifier.
	UserOneID int
	// UserTwoID stores the second trading user identifier.
	UserTwoID int
	// TradedAt stores the trade completion timestamp.
	TradedAt time.Time
}

// TradeLogItem defines one item exchanged in a trade.
type TradeLogItem struct {
	// ID stores stable trade log item identifier.
	ID int
	// TradeID stores the owning trade log identifier.
	TradeID int
	// ItemID stores the traded item instance identifier.
	ItemID int
	// UserID stores the item source owner identifier.
	UserID int
	// DefinitionID stores the item definition identifier.
	DefinitionID int
}
