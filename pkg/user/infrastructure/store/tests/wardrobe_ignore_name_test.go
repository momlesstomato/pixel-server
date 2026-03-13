package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// TestRepositoryWardrobeIgnoreAndNameFlows verifies wardrobe, ignore, and rename repository behavior.
func TestRepositoryWardrobeIgnoreAndNameFlows(t *testing.T) {
	repository := openRepository(t)
	owner, _ := repository.Create(context.Background(), "owner")
	target, _ := repository.Create(context.Background(), "target")
	slot := domain.WardrobeSlot{SlotID: 1, Figure: "hd-180-1", Gender: "F"}
	if err := repository.SaveWardrobeSlot(context.Background(), owner.ID, slot); err != nil {
		t.Fatalf("expected wardrobe save success, got %v", err)
	}
	slots, err := repository.LoadWardrobe(context.Background(), owner.ID)
	if err != nil || len(slots) != 1 || slots[0].Figure != "hd-180-1" {
		t.Fatalf("unexpected wardrobe payload %+v err=%v", slots, err)
	}
	if _, err := repository.IgnoreUserByUsername(context.Background(), owner.ID, "target"); err != nil {
		t.Fatalf("expected ignore success, got %v", err)
	}
	ignored, err := repository.ListIgnoredUsernames(context.Background(), owner.ID)
	if err != nil || len(ignored) != 1 || ignored[0] != "target" {
		t.Fatalf("unexpected ignored list %+v err=%v", ignored, err)
	}
	if _, err := repository.UnignoreUserByUsername(context.Background(), owner.ID, "target"); err != nil {
		t.Fatalf("expected unignore success, got %v", err)
	}
	available, err := repository.IsUsernameAvailable(context.Background(), "owner", owner.ID)
	if err != nil || !available {
		t.Fatalf("expected owner name to be available for same user, got available=%v err=%v", available, err)
	}
	available, err = repository.IsUsernameAvailable(context.Background(), "target", owner.ID)
	if err != nil || available {
		t.Fatalf("expected target name to be unavailable, got available=%v err=%v", available, err)
	}
	if _, err := repository.ChangeUsername(context.Background(), owner.ID, "new-owner", false); !errors.Is(err, domain.ErrNameChangeNotAllowed) {
		t.Fatalf("expected rename guard error, got %v", err)
	}
	changed, err := repository.ChangeUsername(context.Background(), owner.ID, "new-owner", true)
	if err != nil || changed.Username != "new-owner" {
		t.Fatalf("unexpected changed user payload %+v err=%v", changed, err)
	}
	if err := repository.IgnoreUserByID(context.Background(), owner.ID, target.ID); err != nil {
		t.Fatalf("expected ignore by id success, got %v", err)
	}
}
