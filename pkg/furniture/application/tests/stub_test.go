package tests

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// repositoryStub defines deterministic furniture repository behavior.
type repositoryStub struct {
	// definition stores deterministic definition return.
	definition domain.Definition
	// item stores deterministic item return.
	item domain.Item
	// findErr stores deterministic find error.
	findErr error
	// deleteErr stores deterministic delete error.
	deleteErr error
	// transferErr stores deterministic transfer error.
	transferErr error
}

// FindDefinitionByID returns deterministic definition.
func (s repositoryStub) FindDefinitionByID(_ context.Context, _ int) (domain.Definition, error) {
	return s.definition, s.findErr
}

// FindDefinitionByName returns deterministic definition.
func (s repositoryStub) FindDefinitionByName(_ context.Context, _ string) (domain.Definition, error) {
	return s.definition, s.findErr
}

// ListDefinitions returns deterministic definition list.
func (s repositoryStub) ListDefinitions(_ context.Context) ([]domain.Definition, error) {
	return []domain.Definition{s.definition}, s.findErr
}

// CreateDefinition returns deterministic definition.
func (s repositoryStub) CreateDefinition(_ context.Context, d domain.Definition) (domain.Definition, error) {
	d.ID = 1
	return d, nil
}

// UpdateDefinition returns deterministic definition.
func (s repositoryStub) UpdateDefinition(_ context.Context, _ int, _ domain.DefinitionPatch) (domain.Definition, error) {
	return s.definition, s.findErr
}

// DeleteDefinition returns deterministic error.
func (s repositoryStub) DeleteDefinition(_ context.Context, _ int) error {
	return s.deleteErr
}

// FindItemByID returns deterministic item.
func (s repositoryStub) FindItemByID(_ context.Context, _ int) (domain.Item, error) {
	return s.item, s.findErr
}

// ListItemsByUserID returns deterministic item list.
func (s repositoryStub) ListItemsByUserID(_ context.Context, _ int) ([]domain.Item, error) {
	return []domain.Item{s.item}, s.findErr
}

// CreateItem returns deterministic item.
func (s repositoryStub) CreateItem(_ context.Context, i domain.Item) (domain.Item, error) {
	i.ID = 1
	return i, nil
}

// DeleteItem returns deterministic error.
func (s repositoryStub) DeleteItem(_ context.Context, _ int) error {
	return s.deleteErr
}

// TransferItem returns deterministic error.
func (s repositoryStub) TransferItem(_ context.Context, _ int, _ int) error {
	return s.transferErr
}

// CountItemsByUserID returns deterministic count.
func (s repositoryStub) CountItemsByUserID(_ context.Context, _ int) (int, error) {
	return 5, nil
}
