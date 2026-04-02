package events

import (
	"context"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	catalogdomain "github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	inventorydomain "github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	messengerdomain "github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

func catalogPage(caption string) catalogdomain.CatalogPage {
	return catalogdomain.CatalogPage{Caption: caption}
}

// catalogStub provides minimal catalog repository for e2e.
type catalogStub struct{}

func (catalogStub) ListPages(context.Context) ([]catalogdomain.CatalogPage, error) { return nil, nil }
func (catalogStub) FindPageByID(context.Context, int) (catalogdomain.CatalogPage, error) {
	return catalogdomain.CatalogPage{}, nil
}
func (catalogStub) CreatePage(_ context.Context, p catalogdomain.CatalogPage) (catalogdomain.CatalogPage, error) {
	p.ID = 1
	return p, nil
}
func (catalogStub) UpdatePage(context.Context, int, catalogdomain.PagePatch) (catalogdomain.CatalogPage, error) {
	return catalogdomain.CatalogPage{}, nil
}
func (catalogStub) DeletePage(context.Context, int) error { return nil }
func (catalogStub) ListOffersByPageID(context.Context, int) ([]catalogdomain.CatalogOffer, error) {
	return nil, nil
}
func (catalogStub) FindOfferByID(context.Context, int) (catalogdomain.CatalogOffer, error) {
	return catalogdomain.CatalogOffer{}, nil
}
func (catalogStub) CreateOffer(_ context.Context, o catalogdomain.CatalogOffer) (catalogdomain.CatalogOffer, error) {
	o.ID = 1
	return o, nil
}
func (catalogStub) UpdateOffer(context.Context, int, catalogdomain.OfferPatch) (catalogdomain.CatalogOffer, error) {
	return catalogdomain.CatalogOffer{}, nil
}
func (catalogStub) DeleteOffer(context.Context, int) error                         { return nil }
func (catalogStub) IncrementLimitedSells(context.Context, int) (bool, error)       { return true, nil }
func (catalogStub) FindVoucherByCode(context.Context, string) (catalogdomain.Voucher, error) {
	return catalogdomain.Voucher{}, nil
}
func (catalogStub) CreateVoucher(_ context.Context, v catalogdomain.Voucher) (catalogdomain.Voucher, error) {
	v.ID = 1
	return v, nil
}
func (catalogStub) DeleteVoucher(context.Context, int) error                       { return nil }
func (catalogStub) ListVouchers(context.Context) ([]catalogdomain.Voucher, error)  { return nil, nil }
func (catalogStub) RedeemVoucher(context.Context, int, int) error                  { return nil }
func (catalogStub) HasUserRedeemedVoucher(context.Context, int, int) (bool, error) { return false, nil }

// inventoryStub provides minimal inventory repository for e2e.
type inventoryStub struct{}

func (inventoryStub) ListCurrencyTypes(context.Context) ([]inventorydomain.ActivityCurrencyType, error) {
	return nil, nil
}
func (inventoryStub) FindCurrencyTypeByID(context.Context, int) (inventorydomain.ActivityCurrencyType, error) {
	return inventorydomain.ActivityCurrencyType{}, nil
}
func (inventoryStub) ListBadges(context.Context, int) ([]inventorydomain.Badge, error) { return nil, nil }
func (inventoryStub) AwardBadge(_ context.Context, _ int, code string) (inventorydomain.Badge, error) {
	return inventorydomain.Badge{ID: 1, BadgeCode: code}, nil
}
func (inventoryStub) RevokeBadge(context.Context, int, string) error { return nil }
func (inventoryStub) UpdateBadgeSlots(context.Context, int, []inventorydomain.BadgeSlot) error {
	return nil
}
func (inventoryStub) GetEquippedBadges(context.Context, int) ([]inventorydomain.BadgeSlot, error) {
	return nil, nil
}
func (inventoryStub) GetCredits(context.Context, int) (int, error)      { return 0, nil }
func (inventoryStub) SetCredits(context.Context, int, int) error        { return nil }
func (inventoryStub) AddCredits(context.Context, int, int) (int, error) { return 100, nil }
func (inventoryStub) GetCurrency(context.Context, int, inventorydomain.CurrencyType) (int, error) {
	return 0, nil
}
func (inventoryStub) ListCurrencies(context.Context, int) ([]inventorydomain.Currency, error) {
	return nil, nil
}
func (inventoryStub) SetCurrency(context.Context, int, inventorydomain.CurrencyType, int) error {
	return nil
}
func (inventoryStub) AddCurrency(context.Context, int, inventorydomain.CurrencyType, int) (int, error) {
	return 100, nil
}
func (inventoryStub) RecordTransaction(context.Context, inventorydomain.CurrencyTransaction) error {
	return nil
}
func (inventoryStub) ListTransactions(context.Context, int, inventorydomain.CurrencyType, int) ([]inventorydomain.CurrencyTransaction, error) {
	return nil, nil
}
func (inventoryStub) ListEffects(context.Context, int) ([]inventorydomain.Effect, error) {
	return nil, nil
}
func (inventoryStub) AwardEffect(context.Context, int, int, int, bool) (inventorydomain.Effect, error) {
	return inventorydomain.Effect{}, nil
}
func (inventoryStub) ActivateEffect(context.Context, int, int) (inventorydomain.Effect, error) {
	return inventorydomain.Effect{}, nil
}
func (inventoryStub) RemoveExpiredEffects(context.Context) ([]inventorydomain.ExpiredEffect, error) {
	return nil, nil
}

// messengerRepoStub provides minimal messenger repository for e2e.
type messengerRepoStub struct{}

func (messengerRepoStub) ListFriendships(context.Context, int) ([]messengerdomain.Friendship, error) {
	return nil, nil
}
func (messengerRepoStub) AreFriends(context.Context, int, int) (bool, error) { return false, nil }
func (messengerRepoStub) CountFriends(context.Context, int) (int, error)     { return 0, nil }
func (messengerRepoStub) AddFriendship(context.Context, int, int) error      { return nil }
func (messengerRepoStub) RemoveFriendship(context.Context, int, int) error   { return nil }
func (messengerRepoStub) SetRelationship(context.Context, int, int, messengerdomain.RelationshipType) error {
	return nil
}
func (messengerRepoStub) GetRelationship(context.Context, int, int) (messengerdomain.RelationshipType, error) {
	return 0, nil
}
func (messengerRepoStub) GetRelationshipCounts(context.Context, int) ([]messengerdomain.RelationshipCount, error) {
	return nil, nil
}
func (messengerRepoStub) CreateRequest(context.Context, int, int) (messengerdomain.FriendRequest, error) {
	return messengerdomain.FriendRequest{}, nil
}
func (messengerRepoStub) FindRequest(context.Context, int) (messengerdomain.FriendRequest, error) {
	return messengerdomain.FriendRequest{}, nil
}
func (messengerRepoStub) FindRequestByUsers(context.Context, int, int) (messengerdomain.FriendRequest, bool, error) {
	return messengerdomain.FriendRequest{}, false, nil
}
func (messengerRepoStub) ListRequests(context.Context, int) ([]messengerdomain.FriendRequest, error) {
	return nil, nil
}
func (messengerRepoStub) DeleteRequest(context.Context, int) error     { return nil }
func (messengerRepoStub) DeleteAllRequests(context.Context, int) error { return nil }
func (messengerRepoStub) SaveOfflineMessage(context.Context, int, int, string) error { return nil }
func (messengerRepoStub) GetAndDeleteOfflineMessages(context.Context, int) ([]messengerdomain.OfflineMessage, error) {
	return nil, nil
}
func (messengerRepoStub) DeleteOfflineMessagesOlderThan(context.Context, int64) error { return nil }
func (messengerRepoStub) LogMessage(context.Context, int, int, string) error          { return nil }
func (messengerRepoStub) DeleteMessageLogOlderThan(context.Context, int64) error      { return nil }
func (messengerRepoStub) SearchUsers(context.Context, string, int) ([]messengerdomain.SearchResult, error) {
	return nil, nil
}
func (messengerRepoStub) FindUserIDByUsername(context.Context, string) (int, bool, error) {
	return 0, false, nil
}
func (messengerRepoStub) FindUsersByIDs(context.Context, []int) ([]messengerdomain.SearchResult, error) {
	return nil, nil
}

// sessionStub provides minimal session registry.
type sessionStub struct{}

func (sessionStub) Register(coreconnection.Session) error                  { return nil }
func (*sessionStub) FindByConnID(string) (coreconnection.Session, bool)    { return coreconnection.Session{}, false }
func (*sessionStub) FindByUserID(int) (coreconnection.Session, bool)       { return coreconnection.Session{}, false }
func (sessionStub) Touch(string) error                                     { return nil }
func (sessionStub) Remove(string)                                          {}
func (sessionStub) ListAll() ([]coreconnection.Session, error)             { return nil, nil }

// broadcastStub provides minimal broadcaster.
type broadcastStub struct{}

func (*broadcastStub) Publish(context.Context, string, []byte) error { return nil }
func (*broadcastStub) Subscribe(context.Context, string) (<-chan []byte, coreconnection.Disposable, error) {
	return nil, coreconnection.DisposeFunc(func() error { return nil }), nil
}
