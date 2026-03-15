package store

import (
	"context"
	"errors"
	"strings"

	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	"gorm.io/gorm"
)

// ListGroups returns all groups sorted by priority descending and id ascending.
func (repository *Repository) ListGroups(ctx context.Context) ([]permissiondomain.Group, error) {
	var rows []permissionmodel.Group
	if err := repository.database.WithContext(ctx).Order("priority DESC").Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	converted := make([]permissiondomain.Group, 0, len(rows))
	for _, row := range rows {
		converted = append(converted, mapGroup(row))
	}
	return converted, nil
}

// FindGroupByID resolves one group by identifier.
func (repository *Repository) FindGroupByID(ctx context.Context, groupID int) (permissiondomain.Group, error) {
	var row permissionmodel.Group
	if err := repository.database.WithContext(ctx).First(&row, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return permissiondomain.Group{}, permissiondomain.ErrGroupNotFound
		}
		return permissiondomain.Group{}, err
	}
	return mapGroup(row), nil
}

// FindGroupByName resolves one group by name.
func (repository *Repository) FindGroupByName(ctx context.Context, name string) (permissiondomain.Group, error) {
	var row permissionmodel.Group
	if err := repository.database.WithContext(ctx).Where("name = ?", strings.TrimSpace(strings.ToLower(name))).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return permissiondomain.Group{}, permissiondomain.ErrGroupNotFound
		}
		return permissiondomain.Group{}, err
	}
	return mapGroup(row), nil
}

// CreateGroup creates one new permission group.
func (repository *Repository) CreateGroup(ctx context.Context, group permissiondomain.Group) (permissiondomain.Group, error) {
	row := permissionmodel.Group{
		Name: group.Name, DisplayName: group.DisplayName, Priority: group.Priority,
		ClubLevel: group.ClubLevel, SecurityLevel: group.SecurityLevel,
		IsAmbassador: group.IsAmbassador, IsDefault: group.IsDefault,
	}
	if err := repository.database.WithContext(ctx).Create(&row).Error; err != nil {
		return permissiondomain.Group{}, err
	}
	return mapGroup(row), nil
}

// UpdateGroup updates mutable attributes of one group.
func (repository *Repository) UpdateGroup(ctx context.Context, groupID int, patch permissiondomain.GroupPatch) (permissiondomain.Group, error) {
	updates := map[string]any{}
	if patch.DisplayName != nil {
		updates["display_name"] = strings.TrimSpace(*patch.DisplayName)
	}
	if patch.Priority != nil {
		updates["priority"] = *patch.Priority
	}
	if patch.ClubLevel != nil {
		updates["club_level"] = *patch.ClubLevel
	}
	if patch.SecurityLevel != nil {
		updates["security_level"] = *patch.SecurityLevel
	}
	if patch.IsAmbassador != nil {
		updates["is_ambassador"] = *patch.IsAmbassador
	}
	if patch.IsDefault != nil {
		updates["is_default"] = *patch.IsDefault
	}
	if len(updates) > 0 {
		result := repository.database.WithContext(ctx).Model(&permissionmodel.Group{}).Where("id = ?", groupID).Updates(updates)
		if result.Error != nil {
			return permissiondomain.Group{}, result.Error
		}
		if result.RowsAffected == 0 {
			return permissiondomain.Group{}, permissiondomain.ErrGroupNotFound
		}
	}
	return repository.FindGroupByID(ctx, groupID)
}
