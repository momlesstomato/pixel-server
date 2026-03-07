package messaging

const (
	// DisconnectReasonGeneric indicates unspecified disconnect reason.
	DisconnectReasonGeneric int32 = 0
	// DisconnectReasonBanned indicates user is banned.
	DisconnectReasonBanned int32 = 1
	// DisconnectReasonConcurrentLogin indicates a newer session replaced this one.
	DisconnectReasonConcurrentLogin int32 = 2
	// DisconnectReasonHotelClosed indicates hotel is closed.
	DisconnectReasonHotelClosed int32 = 3
	// DisconnectReasonIdleTimeout indicates keepalive/ping timeout.
	DisconnectReasonIdleTimeout int32 = 4
	// DisconnectReasonMaintenance indicates maintenance mode.
	DisconnectReasonMaintenance int32 = 5
)
