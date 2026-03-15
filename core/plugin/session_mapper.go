package plugin

import (
	sdk "github.com/momlesstomato/pixel-sdk"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
)

// mapSessionInfo converts a core session to SDK session info.
func mapSessionInfo(value coreconnection.Session) sdk.SessionInfo {
	return sdk.SessionInfo{ConnID: value.ConnID, UserID: value.UserID, MachineID: value.MachineID, InstanceID: value.InstanceID}
}
