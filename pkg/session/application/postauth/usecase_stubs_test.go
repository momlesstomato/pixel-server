package postauth

import (
	"context"
	"time"

	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// transportStub captures sent packet identifiers.
type transportStub struct {
	// sent stores sent packet identifiers.
	sent []uint16
	// closed stores closed connection identifiers.
	closed []string
}

// Send records sent packet identifiers.
func (stub *transportStub) Send(_ string, packetID uint16, _ []byte) error {
	stub.sent = append(stub.sent, packetID)
	return nil
}

// Close records closed connection identifiers.
func (stub *transportStub) Close(connID string, _ int, _ string) error {
	stub.closed = append(stub.closed, connID)
	return nil
}

// statusStub provides deterministic status responses.
type statusStub struct{ status statusdomain.HotelStatus }

// Current returns deterministic hotel status.
func (stub statusStub) Current(context.Context) (statusdomain.HotelStatus, error) {
	return stub.status, nil
}

// loginStub provides deterministic login stamp responses.
type loginStub struct{ first bool }

// RecordLogin returns deterministic first-login marker.
func (stub loginStub) RecordLogin(context.Context, int, string, time.Time) (bool, error) {
	return stub.first, nil
}

// profileStub provides deterministic user profile reads.
type profileStub struct{}

// FindByID returns deterministic user payload.
func (profileStub) FindByID(context.Context, int) (userdomain.User, error) {
	return userdomain.User{ID: 7, Username: "tester", Figure: "hd-180-1", Gender: "M", Motto: "hello", HomeRoomID: -1, NoobnessLevel: 2}, nil
}

// LoadSettings returns deterministic user settings payload.
func (profileStub) LoadSettings(context.Context, int) (userdomain.Settings, error) {
	return userdomain.Settings{VolumeSystem: 100, VolumeFurni: 100, VolumeTrax: 100, RoomInvites: true, CameraFollow: true}, nil
}

// RemainingRespects returns deterministic remaining respects payload.
func (profileStub) RemainingRespects(context.Context, int, userdomain.RespectTargetType, time.Time) (int, error) {
	return 3, nil
}

// ListIgnoredUsernames returns deterministic ignored usernames payload.
func (profileStub) ListIgnoredUsernames(context.Context, int) ([]string, error) {
	return []string{}, nil
}

// accessStub provides deterministic permission access reads.
type accessStub struct{}

// ResolveAccess returns deterministic access payload.
func (accessStub) ResolveAccess(context.Context, int) (permissiondomain.Access, error) {
	return permissiondomain.Access{PrimaryGroup: permissiondomain.Group{ID: 1, ClubLevel: 0, SecurityLevel: 0, IsAmbassador: false}}, nil
}

// ResolvePerks returns deterministic perk grants payload.
func (accessStub) ResolvePerks(permissiondomain.Access) []permissiondomain.PerkGrant {
	return []permissiondomain.PerkGrant{{Code: "SAFE_CHAT", IsAllowed: true}}
}

// equalIDs compares packet identifier slices.
func equalIDs(left []uint16, right []uint16) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
