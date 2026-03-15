package application

import (
	"context"
	"fmt"

	sdkpermission "github.com/momlesstomato/pixel-sdk/events/permission"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
)

// ReplaceUserGroups replaces group assignments for one user and returns resolved access.
func (service *Service) ReplaceUserGroups(ctx context.Context, userID int, groupIDs []int) (permissiondomain.Access, error) {
	if userID <= 0 {
		return permissiondomain.Access{}, fmt.Errorf("user id must be positive")
	}
	oldAccess, err := service.ResolveAccess(ctx, userID)
	if err != nil {
		return permissiondomain.Access{}, err
	}
	normalized := dedupeGroupIDs(groupIDs)
	if len(normalized) == 0 {
		defaultGroup, defaultErr := service.repository.FindDefaultGroup(ctx)
		if defaultErr != nil {
			return permissiondomain.Access{}, defaultErr
		}
		normalized = []int{defaultGroup.ID}
	}
	nextAccess, err := service.resolveAccessFromGroups(ctx, userID, normalized)
	if err != nil {
		return permissiondomain.Access{}, err
	}
	if service.fire != nil {
		event := &sdkpermission.UserGroupChanged{
			UserID: userID, OldGroupID: oldAccess.PrimaryGroup.ID, NewGroupID: nextAccess.PrimaryGroup.ID,
			OldGroupIDs: oldAccess.GroupIDs, NewGroupIDs: nextAccess.GroupIDs,
		}
		service.fire(event)
		if event.Cancelled() {
			return permissiondomain.Access{}, fmt.Errorf("user group change cancelled by plugin")
		}
	}
	if err := service.repository.ReplaceUserGroups(ctx, userID, normalized); err != nil {
		return permissiondomain.Access{}, err
	}
	updated, err := service.ResolveAccess(ctx, userID)
	if err != nil {
		return permissiondomain.Access{}, err
	}
	if service.liveUpdater != nil {
		if liveErr := service.liveUpdater.PushAccessUpdate(ctx, updated, service.ResolvePerks(updated)); liveErr != nil {
			return permissiondomain.Access{}, liveErr
		}
	}
	return updated, nil
}
