package store

import (
	"context"
	"errors"

	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	"gorm.io/gorm"
)

// DeleteGroup deletes one group when constraints allow.
func (repository *Repository) DeleteGroup(ctx context.Context, groupID int) error {
	row, err := repository.FindGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if row.IsDefault {
		return permissiondomain.ErrCannotDeleteDefaultGroup
	}
	count, countErr := repository.CountGroupUsers(ctx, groupID)
	if countErr != nil {
		return countErr
	}
	if count > 0 {
		return permissiondomain.ErrGroupInUse
	}
	if err := repository.database.WithContext(ctx).Where("group_id = ?", groupID).Delete(&permissionmodel.Grant{}).Error; err != nil {
		return err
	}
	result := repository.database.WithContext(ctx).Delete(&permissionmodel.Group{}, groupID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return permissiondomain.ErrGroupNotFound
	}
	return nil
}

// CountGroupUsers counts users assigned to one group.
func (repository *Repository) CountGroupUsers(ctx context.Context, groupID int) (int64, error) {
	var count int64
	if err := repository.database.WithContext(ctx).Model(&permissionmodel.Assignment{}).Where("group_id = ?", groupID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindDefaultGroup resolves the active default group.
func (repository *Repository) FindDefaultGroup(ctx context.Context) (permissiondomain.Group, error) {
	var row permissionmodel.Group
	if err := repository.database.WithContext(ctx).Where("is_default = ?", true).Order("priority DESC").Order("id ASC").First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return permissiondomain.Group{}, permissiondomain.ErrDefaultGroupRequired
		}
		return permissiondomain.Group{}, err
	}
	return mapGroup(row), nil
}

// SwitchDefaultGroup marks one group as default and unmarks the previous default group.
func (repository *Repository) SwitchDefaultGroup(ctx context.Context, groupID int) error {
	return repository.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Model(&permissionmodel.Group{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		result := transaction.Model(&permissionmodel.Group{}).Where("id = ?", groupID).Update("is_default", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return permissiondomain.ErrGroupNotFound
		}
		return nil
	})
}
