package store

import (
	"context"
	"strings"

	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListGroupPermissions returns one group's granted permissions.
func (repository *Repository) ListGroupPermissions(ctx context.Context, groupID int) ([]string, error) {
	var rows []permissionmodel.Grant
	if err := repository.database.WithContext(ctx).Where("group_id = ?", groupID).Order("permission ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	permissions := make([]string, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, row.Permission)
	}
	return permissions, nil
}

// AddGroupPermissions adds permission grants to one group.
func (repository *Repository) AddGroupPermissions(ctx context.Context, groupID int, permissions []string) error {
	if len(permissions) == 0 {
		return nil
	}
	rows := make([]permissionmodel.Grant, 0, len(permissions))
	for _, permission := range permissions {
		rows = append(rows, permissionmodel.Grant{GroupID: uint(groupID), Permission: strings.TrimSpace(permission)})
	}
	return repository.database.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

// RemoveGroupPermission removes one permission grant from one group.
func (repository *Repository) RemoveGroupPermission(ctx context.Context, groupID int, permission string) error {
	return repository.database.WithContext(ctx).Where("group_id = ? AND permission = ?", groupID, strings.TrimSpace(permission)).Delete(&permissionmodel.Grant{}).Error
}

// ListUserGroupIDs resolves assigned group identifiers for one user.
func (repository *Repository) ListUserGroupIDs(ctx context.Context, userID int) ([]int, error) {
	var rows []permissionmodel.Assignment
	if err := repository.database.WithContext(ctx).Where("user_id = ?", userID).Order("group_id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	groupIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		groupIDs = append(groupIDs, int(row.GroupID))
	}
	return groupIDs, nil
}

// ReplaceUserGroups replaces group assignments for one user.
func (repository *Repository) ReplaceUserGroups(ctx context.Context, userID int, groupIDs []int) error {
	return repository.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Where("user_id = ?", userID).Delete(&permissionmodel.Assignment{}).Error; err != nil {
			return err
		}
		if len(groupIDs) == 0 {
			return nil
		}
		rows := make([]permissionmodel.Assignment, 0, len(groupIDs))
		for _, groupID := range groupIDs {
			rows = append(rows, permissionmodel.Assignment{UserID: uint(userID), GroupID: uint(groupID)})
		}
		return transaction.Create(&rows).Error
	})
}
