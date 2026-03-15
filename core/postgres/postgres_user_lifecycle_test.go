package postgres

import (
	"testing"

	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// assertUserSoftDeleteLifecycle verifies user soft-delete persistence behavior.
func assertUserSoftDeleteLifecycle(t *testing.T, database *gorm.DB) {
	t.Helper()
	user := usermodel.Record{Username: "tester"}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("expected user insert success, got %v", err)
	}
	if user.ID == 0 || user.CreatedAt.IsZero() || user.UpdatedAt.IsZero() || user.OwnerID != nil {
		t.Fatalf("expected generated id, timestamps, and nil owner for inserted user")
	}
	if err := database.Delete(&user).Error; err != nil {
		t.Fatalf("expected user soft delete success, got %v", err)
	}
	var visibleUsers int64
	if err := database.Model(&usermodel.Record{}).Where("username = ?", user.Username).Count(&visibleUsers).Error; err != nil {
		t.Fatalf("expected visible user count query success, got %v", err)
	}
	if visibleUsers != 0 {
		t.Fatalf("expected zero visible users after soft delete, got %d", visibleUsers)
	}
	var storedUser usermodel.Record
	if err := database.Unscoped().Where("id = ?", user.ID).First(&storedUser).Error; err != nil {
		t.Fatalf("expected unscoped user lookup success, got %v", err)
	}
	if !storedUser.DeletedAt.Valid {
		t.Fatalf("expected deleted_at to be set after soft delete")
	}
}
