package postgresstore

import (
	"context"
	"errors"
	"testing"

	usermodel "github.com/momlesstomato/pixel-server/core/postgres/model/user"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestNewRepositoryRejectsNilDatabase verifies constructor validation behavior.
func TestNewRepositoryRejectsNilDatabase(t *testing.T) {
	if _, err := NewRepository(nil); err == nil {
		t.Fatalf("expected nil database validation failure")
	}
}

// TestRepositoryCreateAndFindByID verifies persisted user read/write behavior.
func TestRepositoryCreateAndFindByID(t *testing.T) {
	database := openDatabase(t)
	repository, _ := NewRepository(database)
	created, err := repository.Create(context.Background(), "tester")
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if created.ID == 0 || created.Username != "tester" {
		t.Fatalf("unexpected created user payload")
	}
	loaded, err := repository.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected find success, got %v", err)
	}
	if loaded.ID != created.ID || loaded.Username != created.Username {
		t.Fatalf("unexpected loaded user payload")
	}
}

// TestRepositoryFindByIDHandlesNotFoundAndSoftDelete verifies soft-delete visibility behavior.
func TestRepositoryFindByIDHandlesNotFoundAndSoftDelete(t *testing.T) {
	database := openDatabase(t)
	repository, _ := NewRepository(database)
	if _, err := repository.FindByID(context.Background(), 99); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found error, got %v", err)
	}
	record := usermodel.Record{Username: "to-delete"}
	if err := database.Create(&record).Error; err != nil {
		t.Fatalf("expected seed user create success, got %v", err)
	}
	if err := database.Delete(&record).Error; err != nil {
		t.Fatalf("expected user soft delete success, got %v", err)
	}
	if _, err := repository.FindByID(context.Background(), int(record.ID)); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found for soft-deleted row, got %v", err)
	}
}

// TestRepositoryDeleteByIDSoftDeletesRecord verifies repository soft-delete behavior.
func TestRepositoryDeleteByIDSoftDeletesRecord(t *testing.T) {
	database := openDatabase(t)
	repository, _ := NewRepository(database)
	created, err := repository.Create(context.Background(), "delete-me")
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if err := repository.DeleteByID(context.Background(), created.ID); err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
	if _, err := repository.FindByID(context.Background(), created.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found after soft delete, got %v", err)
	}
	var record usermodel.Record
	if err := database.Unscoped().First(&record, created.ID).Error; err != nil {
		t.Fatalf("expected unscoped deleted record lookup success, got %v", err)
	}
	if !record.DeletedAt.Valid {
		t.Fatalf("expected deleted_at to be set")
	}
	if err := repository.DeleteByID(context.Background(), created.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found when deleting already deleted row, got %v", err)
	}
}

// openDatabase creates a sqlite database with migrated user schema.
func openDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&usermodel.Record{}); err != nil {
		t.Fatalf("expected sqlite migration success, got %v", err)
	}
	return database
}
