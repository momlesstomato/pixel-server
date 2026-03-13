package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestNewRepositoryRejectsNilDatabase verifies constructor validation behavior.
func TestNewRepositoryRejectsNilDatabase(t *testing.T) {
	if _, err := userstore.NewRepository(nil); err == nil {
		t.Fatalf("expected nil database validation failure")
	}
}

// TestRepositoryCreateFindDeleteAndLogin verifies CRUD and login persistence behavior.
func TestRepositoryCreateFindDeleteAndLogin(t *testing.T) {
	database := openCRUDDatabase(t)
	repository, _ := userstore.NewRepository(database)
	created, err := repository.Create(context.Background(), "tester")
	if err != nil || created.ID == 0 || created.Username != "tester" {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
	loaded, err := repository.FindByID(context.Background(), created.ID)
	if err != nil || loaded.ID != created.ID {
		t.Fatalf("unexpected find result %+v err=%v", loaded, err)
	}
	if err := repository.DeleteByID(context.Background(), created.ID); err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
	if _, err := repository.FindByID(context.Background(), created.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found after delete, got %v", err)
	}
	loggedAt := time.Date(2026, time.March, 12, 8, 30, 0, 0, time.UTC)
	first, err := repository.RecordLogin(context.Background(), 7, "pixel-server", loggedAt)
	if err != nil || !first {
		t.Fatalf("unexpected first login result first=%v err=%v", first, err)
	}
	first, err = repository.RecordLogin(context.Background(), 7, "pixel-server", loggedAt.Add(2*time.Hour))
	if err != nil || first {
		t.Fatalf("unexpected same-day login result first=%v err=%v", first, err)
	}
}

// TestRepositoryFindByIDNotFoundAndSoftDelete verifies missing and soft-deleted lookup behavior.
func TestRepositoryFindByIDNotFoundAndSoftDelete(t *testing.T) {
	database := openCRUDDatabase(t)
	repository, _ := userstore.NewRepository(database)
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

// openCRUDDatabase creates sqlite database with repository schemas.
func openCRUDDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&usermodel.Record{}, &usermodel.LoginEvent{}, &usermodel.Settings{}, &usermodel.Respect{}, &usermodel.WardrobeSlot{}, &usermodel.Ignore{}); err != nil {
		t.Fatalf("expected sqlite migration success, got %v", err)
	}
	return database
}
