package application

import (
	"context"
	"fmt"
	"strings"

	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
)

// UpdateGroup updates one group and returns resulting details.
func (service *Service) UpdateGroup(ctx context.Context, groupID int, patch permissiondomain.GroupPatch) (GroupDetails, error) {
	if groupID <= 0 {
		return GroupDetails{}, fmt.Errorf("group id must be positive")
	}
	if patch.DisplayName != nil {
		value := strings.TrimSpace(*patch.DisplayName)
		patch.DisplayName = &value
	}
	if patch.IsDefault != nil && *patch.IsDefault {
		if err := service.repository.SwitchDefaultGroup(ctx, groupID); err != nil {
			return GroupDetails{}, err
		}
	}
	if patch.IsDefault != nil && !*patch.IsDefault {
		current, err := service.repository.FindGroupByID(ctx, groupID)
		if err != nil {
			return GroupDetails{}, err
		}
		if current.IsDefault {
			return GroupDetails{}, permissiondomain.ErrDefaultGroupRequired
		}
	}
	updated, err := service.repository.UpdateGroup(ctx, groupID, patch)
	if err != nil {
		return GroupDetails{}, err
	}
	service.invalidateGroupCache(ctx, groupID)
	return service.groupDetails(ctx, updated.ID)
}

// AddPermissions adds one or many permissions to one group.
func (service *Service) AddPermissions(ctx context.Context, groupID int, permissions []string) (GroupDetails, error) {
	if groupID <= 0 {
		return GroupDetails{}, fmt.Errorf("group id must be positive")
	}
	if _, err := service.repository.FindGroupByID(ctx, groupID); err != nil {
		return GroupDetails{}, err
	}
	normalized := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		value, err := permissiondomain.ValidatePermission(permission)
		if err != nil {
			return GroupDetails{}, err
		}
		normalized = append(normalized, value)
	}
	if err := service.repository.AddGroupPermissions(ctx, groupID, normalized); err != nil {
		return GroupDetails{}, err
	}
	service.invalidateGroupCache(ctx, groupID)
	return service.groupDetails(ctx, groupID)
}

// RemovePermission removes one permission from one group.
func (service *Service) RemovePermission(ctx context.Context, groupID int, permission string) (GroupDetails, error) {
	if groupID <= 0 {
		return GroupDetails{}, fmt.Errorf("group id must be positive")
	}
	if _, err := service.repository.FindGroupByID(ctx, groupID); err != nil {
		return GroupDetails{}, err
	}
	normalized, err := permissiondomain.ValidatePermission(permission)
	if err != nil {
		return GroupDetails{}, err
	}
	if err := service.repository.RemoveGroupPermission(ctx, groupID, normalized); err != nil {
		return GroupDetails{}, err
	}
	service.invalidateGroupCache(ctx, groupID)
	return service.groupDetails(ctx, groupID)
}

// invalidateGroupCache deletes one cached group snapshot.
func (service *Service) invalidateGroupCache(ctx context.Context, groupID int) {
	if service.redis == nil || groupID <= 0 {
		return
	}
	service.redis.Del(ctx, service.groupCacheKey(groupID))
}
