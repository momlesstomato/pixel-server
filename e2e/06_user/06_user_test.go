package user

import (
	"context"
	"errors"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test06UserCreateAndLookupFlow verifies end-to-end user realm persistence flow.
func Test06UserCreateAndLookupFlow(t *testing.T) {
	database := openDatabase(t)
	repository, err := userstore.NewRepository(database)
	if err != nil {
		t.Fatalf("expected repository creation success, got %v", err)
	}
	service, err := application.NewService(repository)
	if err != nil {
		t.Fatalf("expected service creation success, got %v", err)
	}
	created, err := service.Create(context.Background(), "  e2e-user  ")
	if err != nil {
		t.Fatalf("expected user creation success, got %v", err)
	}
	if created.ID <= 0 || created.Username != "e2e-user" {
		t.Fatalf("unexpected created user payload")
	}
	loaded, err := service.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected user lookup success, got %v", err)
	}
	if loaded.ID != created.ID || loaded.Username != created.Username {
		t.Fatalf("unexpected loaded user payload")
	}
}

// Test06UserSoftDeleteFlow verifies end-to-end user soft-delete behavior.
func Test06UserSoftDeleteFlow(t *testing.T) {
	database := openDatabase(t)
	repository, err := userstore.NewRepository(database)
	if err != nil {
		t.Fatalf("expected repository creation success, got %v", err)
	}
	service, err := application.NewService(repository)
	if err != nil {
		t.Fatalf("expected service creation success, got %v", err)
	}
	created, err := service.Create(context.Background(), "e2e-delete")
	if err != nil {
		t.Fatalf("expected user creation success, got %v", err)
	}
	if err := service.DeleteByID(context.Background(), created.ID); err != nil {
		t.Fatalf("expected user soft-delete success, got %v", err)
	}
	if _, err := service.FindByID(context.Background(), created.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found after soft delete, got %v", err)
	}
	if err := service.DeleteByID(context.Background(), created.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found when deleting already deleted user, got %v", err)
	}
}

// openDatabase creates one sqlite database with user schema.
func openDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&usermodel.Record{}); err != nil {
		t.Fatalf("expected user migration success, got %v", err)
	}
	return database
}
