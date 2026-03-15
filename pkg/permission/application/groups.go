package application

import (
	"context"
	"fmt"
	"strings"

	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
)

// GroupDetails defines one group plus granted permissions payload.
type GroupDetails struct {
	// Group stores group attributes.
	Group permissiondomain.Group
	// Permissions stores granted permissions.
	Permissions []string
}

// CreateGroupInput defines create-group payload.
type CreateGroupInput struct {
	// Name stores group name.
	Name string
	// DisplayName stores display name.
	DisplayName string
	// Priority stores group priority.
	Priority int
	// ClubLevel stores group club-level attribute.
	ClubLevel int
	// SecurityLevel stores group security-level attribute.
	SecurityLevel int
	// IsAmbassador stores group ambassador attribute.
	IsAmbassador bool
	// IsDefault stores default-group marker.
	IsDefault bool
}

// ListGroups returns all groups with granted permissions.
func (service *Service) ListGroups(ctx context.Context) ([]GroupDetails, error) {
	groups, err := service.repository.ListGroups(ctx)
	if err != nil {
		return nil, err
	}
	output := make([]GroupDetails, 0, len(groups))
	for _, group := range groups {
		value, loadErr := service.groupDetails(ctx, group.ID)
		if loadErr != nil {
			return nil, loadErr
		}
		output = append(output, value)
	}
	return output, nil
}

// GetGroup resolves one group with granted permissions.
func (service *Service) GetGroup(ctx context.Context, groupID int) (GroupDetails, error) {
	if groupID <= 0 {
		return GroupDetails{}, fmt.Errorf("group id must be positive")
	}
	return service.groupDetails(ctx, groupID)
}

// GetGroupByName resolves one group by name and returns resulting details.
func (service *Service) GetGroupByName(ctx context.Context, name string) (GroupDetails, error) {
	group, err := service.repository.FindGroupByName(ctx, name)
	if err != nil {
		return GroupDetails{}, err
	}
	return service.groupDetails(ctx, group.ID)
}

// CreateGroup creates one group and returns resulting details.
func (service *Service) CreateGroup(ctx context.Context, input CreateGroupInput) (GroupDetails, error) {
	name, err := permissiondomain.ValidateGroupName(input.Name)
	if err != nil {
		return GroupDetails{}, err
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = name
	}
	group, err := service.repository.CreateGroup(ctx, permissiondomain.Group{
		Name: name, DisplayName: displayName, Priority: input.Priority,
		ClubLevel: input.ClubLevel, SecurityLevel: input.SecurityLevel,
		IsAmbassador: input.IsAmbassador, IsDefault: input.IsDefault,
	})
	if err != nil {
		return GroupDetails{}, err
	}
	if input.IsDefault {
		if switchErr := service.repository.SwitchDefaultGroup(ctx, group.ID); switchErr != nil {
			return GroupDetails{}, switchErr
		}
	}
	return service.groupDetails(ctx, group.ID)
}

// DeleteGroup deletes one group.
func (service *Service) DeleteGroup(ctx context.Context, groupID int) error {
	if groupID <= 0 {
		return fmt.Errorf("group id must be positive")
	}
	return service.repository.DeleteGroup(ctx, groupID)
}

// groupDetails resolves group with permissions and cache population.
func (service *Service) groupDetails(ctx context.Context, groupID int) (GroupDetails, error) {
	snapshot, err := service.loadGroupSnapshot(ctx, groupID)
	if err != nil {
		return GroupDetails{}, err
	}
	return GroupDetails{Group: snapshot.Group, Permissions: snapshot.Permissions}, nil
}
