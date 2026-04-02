package events

import (
	"context"
	"testing"

	sdkcatalog "github.com/momlesstomato/pixel-sdk/events/catalog"
	sdkinventory "github.com/momlesstomato/pixel-sdk/events/inventory"
	sdkmessenger "github.com/momlesstomato/pixel-sdk/events/messenger"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	inventoryapplication "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	messengerconfig "github.com/momlesstomato/pixel-server/pkg/messenger/application"
	"go.uber.org/zap"
)

// Test11PluginCancelsPageCreation verifies a plugin can cancel page creation via event.
func Test11PluginCancelsPageCreation(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	dispatcher.Subscribe("test", func(event *sdkcatalog.PageCreating) {
		event.Cancel()
	})
	service, _ := catalogapplication.NewService(catalogStub{})
	service.SetEventFirer(dispatcher.Fire)
	if _, err := service.CreatePage(context.Background(), catalogPage("Blocked")); err == nil {
		t.Fatalf("expected page creation to be cancelled by plugin")
	}
}

// Test11PluginAllowsPageCreationAndReceivesAfterEvent verifies after event fires when not cancelled.
func Test11PluginAllowsPageCreationAndReceivesAfterEvent(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	var afterPageID int
	dispatcher.Subscribe("test", func(event *sdkcatalog.PageCreated) {
		afterPageID = event.PageID
	})
	service, _ := catalogapplication.NewService(catalogStub{})
	service.SetEventFirer(dispatcher.Fire)
	page, err := service.CreatePage(context.Background(), catalogPage("Allowed"))
	if err != nil {
		t.Fatalf("expected page creation success, got %v", err)
	}
	if afterPageID != page.ID {
		t.Fatalf("expected PageCreated event with ID %d, got %d", page.ID, afterPageID)
	}
}

// Test11PluginCancelsBadgeAward verifies a plugin can cancel badge award via event.
func Test11PluginCancelsBadgeAward(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	dispatcher.Subscribe("test", func(event *sdkinventory.BadgeAwarding) {
		if event.BadgeCode == "BLOCKED" {
			event.Cancel()
		}
	})
	service, _ := inventoryapplication.NewService(inventoryStub{})
	service.SetEventFirer(dispatcher.Fire)
	if _, err := service.AwardBadge(context.Background(), 1, "BLOCKED"); err == nil {
		t.Fatalf("expected badge award to be cancelled by plugin")
	}
	badge, err := service.AwardBadge(context.Background(), 1, "ALLOWED")
	if err != nil || badge.BadgeCode != "ALLOWED" {
		t.Fatalf("unexpected badge result %+v err=%v", badge, err)
	}
}

// Test11PluginReceivesFriendAddedEvent verifies plugin receives friend added after event.
func Test11PluginReceivesFriendAddedEvent(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	var receivedUserOne, receivedUserTwo int
	dispatcher.Subscribe("test", func(event *sdkmessenger.FriendAdded) {
		receivedUserOne = event.UserOneID
		receivedUserTwo = event.UserTwoID
	})
	service := newMessengerService(dispatcher)
	if err := service.AddFriendship(context.Background(), 10, 20); err != nil {
		t.Fatalf("unexpected add friendship error: %v", err)
	}
	if receivedUserOne != 10 || receivedUserTwo != 20 {
		t.Fatalf("expected FriendAdded with 10,20 got %d,%d", receivedUserOne, receivedUserTwo)
	}
}

func newMessengerService(dispatcher *coreplugin.Dispatcher) *messengerconfig.Service {
	svc, err := messengerconfig.NewService(messengerRepoStub{}, &sessionStub{}, &broadcastStub{}, messengerconfig.Config{
		MaxFriends: 100, MaxFriendsVIP: 200, FloodCooldownMs: 750, FloodViolations: 4, FloodMuteSeconds: 20,
	})
	if err != nil {
		panic(err)
	}
	svc.SetEventFirer(dispatcher.Fire)
	return svc
}
