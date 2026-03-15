package application

import (
	"context"
	"sort"

	corepermission "github.com/momlesstomato/pixel-server/core/permission"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
)

// ResolveAccess resolves user effective group and merged permissions.
func (service *Service) ResolveAccess(ctx context.Context, userID int) (permissiondomain.Access, error) {
	groupIDs, err := service.repository.ListUserGroupIDs(ctx, userID)
	if err != nil {
		return permissiondomain.Access{}, err
	}
	if len(groupIDs) == 0 {
		defaultGroup, defaultErr := service.repository.FindDefaultGroup(ctx)
		if defaultErr != nil {
			return permissiondomain.Access{}, defaultErr
		}
		groupIDs = []int{defaultGroup.ID}
	}
	return service.resolveAccessFromGroups(ctx, userID, groupIDs)
}

// ResolvePerks resolves known client perk codes for one access snapshot.
func (service *Service) ResolvePerks(access permissiondomain.Access) []permissiondomain.PerkGrant {
	output := make([]permissiondomain.PerkGrant, 0, len(knownPerks))
	for _, perk := range knownPerks {
		allowed := corepermission.Resolve(access.Permissions, perk.Permission)
		output = append(output, permissiondomain.PerkGrant{Code: perk.Code, ErrorMessage: perk.DeniedMessage, IsAllowed: allowed})
	}
	return output
}

// resolveAccessFromGroups resolves one access snapshot from explicit group identifiers.
func (service *Service) resolveAccessFromGroups(ctx context.Context, userID int, groupIDs []int) (permissiondomain.Access, error) {
	permissions := map[string]struct{}{}
	snapshots := make([]cachedGroup, 0, len(groupIDs))
	for _, groupID := range dedupeGroupIDs(groupIDs) {
		snapshot, err := service.loadGroupSnapshot(ctx, groupID)
		if err != nil {
			return permissiondomain.Access{}, err
		}
		snapshots = append(snapshots, snapshot)
		for _, permission := range snapshot.Permissions {
			permissions[permission] = struct{}{}
		}
	}
	sort.SliceStable(snapshots, func(left int, right int) bool {
		if snapshots[left].Group.Priority == snapshots[right].Group.Priority {
			return snapshots[left].Group.ID < snapshots[right].Group.ID
		}
		return snapshots[left].Group.Priority > snapshots[right].Group.Priority
	})
	primary := snapshots[0].Group
	if corepermission.Resolve(permissions, service.ambassadorPermission) {
		primary.IsAmbassador = true
	}
	normalizedIDs := make([]int, 0, len(snapshots))
	for _, snapshot := range snapshots {
		normalizedIDs = append(normalizedIDs, snapshot.Group.ID)
	}
	return permissiondomain.Access{UserID: userID, PrimaryGroup: primary, GroupIDs: normalizedIDs, Permissions: permissions}, nil
}

// loadGroupSnapshot resolves one group and permissions using cache.
func (service *Service) loadGroupSnapshot(ctx context.Context, groupID int) (cachedGroup, error) {
	if cached, ok := service.loadCachedGroup(ctx, groupID); ok {
		return cached, nil
	}
	group, err := service.repository.FindGroupByID(ctx, groupID)
	if err != nil {
		return cachedGroup{}, err
	}
	permissions, err := service.repository.ListGroupPermissions(ctx, groupID)
	if err != nil {
		return cachedGroup{}, err
	}
	snapshot := cachedGroup{Group: group, Permissions: permissions}
	service.storeCachedGroup(ctx, groupID, snapshot)
	return snapshot, nil
}

// dedupeGroupIDs returns sorted deduplicated group identifiers.
func dedupeGroupIDs(groupIDs []int) []int {
	seen := map[int]struct{}{}
	output := make([]int, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		if groupID <= 0 {
			continue
		}
		if _, exists := seen[groupID]; exists {
			continue
		}
		seen[groupID] = struct{}{}
		output = append(output, groupID)
	}
	sort.Ints(output)
	return output
}
