package tests

import (
	"context"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	messengerapplication "github.com/momlesstomato/pixel-server/pkg/messenger/application"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// repositoryStub provides deterministic repository behavior for tests.
type repositoryStub struct {
	friendships   []domain.Friendship
	areFriends    bool
	friendCount   int
	relationship  domain.RelationshipType
	relCounts     []domain.RelationshipCount
	request       domain.FriendRequest
	requestFound  bool
	requests      []domain.FriendRequest
	offline       []domain.OfflineMessage
	searchResults []domain.SearchResult
	userID        int
	usernameFound bool
	callErr       error
}

// ListFriendships returns stub friendships.
func (s repositoryStub) ListFriendships(_ context.Context, _ int) ([]domain.Friendship, error) {
	return s.friendships, s.callErr
}

// AreFriends returns stub result.
func (s repositoryStub) AreFriends(_ context.Context, _, _ int) (bool, error) {
	return s.areFriends, s.callErr
}

// CountFriends returns stub count.
func (s repositoryStub) CountFriends(_ context.Context, _ int) (int, error) {
	return s.friendCount, s.callErr
}

// AddFriendship returns stub error.
func (s repositoryStub) AddFriendship(_ context.Context, _, _ int) error { return s.callErr }

// RemoveFriendship returns stub error.
func (s repositoryStub) RemoveFriendship(_ context.Context, _, _ int) error { return s.callErr }

// SetRelationship returns stub error.
func (s repositoryStub) SetRelationship(_ context.Context, _, _ int, _ domain.RelationshipType) error {
	return s.callErr
}

// GetRelationship returns stub relationship.
func (s repositoryStub) GetRelationship(_ context.Context, _, _ int) (domain.RelationshipType, error) {
	return s.relationship, s.callErr
}

// GetRelationshipCounts returns stub counts.
func (s repositoryStub) GetRelationshipCounts(_ context.Context, _ int) ([]domain.RelationshipCount, error) {
	return s.relCounts, s.callErr
}

// CreateRequest returns stub request.
func (s repositoryStub) CreateRequest(_ context.Context, from, to int) (domain.FriendRequest, error) {
	return domain.FriendRequest{ID: 1, FromUserID: from, ToUserID: to}, s.callErr
}

// FindRequest returns stub request.
func (s repositoryStub) FindRequest(_ context.Context, _ int) (domain.FriendRequest, error) {
	return s.request, s.callErr
}

// FindRequestByUsers returns stub result.
func (s repositoryStub) FindRequestByUsers(_ context.Context, _, _ int) (domain.FriendRequest, bool, error) {
	return s.request, s.requestFound, s.callErr
}

// ListRequests returns stub requests.
func (s repositoryStub) ListRequests(_ context.Context, _ int) ([]domain.FriendRequest, error) {
	return s.requests, s.callErr
}

// DeleteRequest returns stub error.
func (s repositoryStub) DeleteRequest(_ context.Context, _ int) error { return s.callErr }

// DeleteAllRequests returns stub error.
func (s repositoryStub) DeleteAllRequests(_ context.Context, _ int) error { return s.callErr }

// SaveOfflineMessage returns stub error.
func (s repositoryStub) SaveOfflineMessage(_ context.Context, _, _ int, _ string) error {
	return s.callErr
}

// GetAndDeleteOfflineMessages returns stub messages.
func (s repositoryStub) GetAndDeleteOfflineMessages(_ context.Context, _ int) ([]domain.OfflineMessage, error) {
	return s.offline, s.callErr
}

// DeleteOfflineMessagesOlderThan returns stub error.
func (s repositoryStub) DeleteOfflineMessagesOlderThan(_ context.Context, _ int64) error {
	return s.callErr
}

// SearchUsers returns stub results.
func (s repositoryStub) SearchUsers(_ context.Context, _ string, _ int) ([]domain.SearchResult, error) {
	return s.searchResults, s.callErr
}

// FindUserIDByUsername returns stub user id result.
func (s repositoryStub) FindUserIDByUsername(_ context.Context, _ string) (int, bool, error) {
	return s.userID, s.usernameFound, s.callErr
}

// FindUsersByIDs returns stub search results for a set of user identifiers.
func (s repositoryStub) FindUsersByIDs(_ context.Context, _ []int) ([]domain.SearchResult, error) {
	return s.searchResults, s.callErr
}

// broadcasterStub records published messages.
type broadcasterStub struct {
	published [][]byte
	callErr   error
}

// Publish records the payload.
func (b *broadcasterStub) Publish(_ context.Context, _ string, payload []byte) error {
	b.published = append(b.published, payload)
	return b.callErr
}

// Subscribe returns a no-op channel.
func (b *broadcasterStub) Subscribe(_ context.Context, _ string) (<-chan []byte, coreconnection.Disposable, error) {
	ch := make(chan []byte)
	close(ch)
	return ch, coreconnection.DisposeFunc(func() error { return nil }), nil
}

// sessionRegistryStub resolves sessions by connID.
type sessionRegistryStub struct {
	byConnID map[string]coreconnection.Session
}

// Register does nothing.
func (s *sessionRegistryStub) Register(_ coreconnection.Session) error { return nil }

// FindByUserID always returns not found.
func (s *sessionRegistryStub) FindByUserID(_ int) (coreconnection.Session, bool) {
	return coreconnection.Session{}, false
}

// FindByConnID looks up by connection id.
func (s *sessionRegistryStub) FindByConnID(connID string) (coreconnection.Session, bool) {
	if s.byConnID != nil {
		sess, ok := s.byConnID[connID]
		return sess, ok
	}
	return coreconnection.Session{}, false
}

// Touch does nothing.
func (s *sessionRegistryStub) Touch(_ string) error { return nil }

// Remove does nothing.
func (s *sessionRegistryStub) Remove(_ string) {}

// ListAll returns nil.
func (s *sessionRegistryStub) ListAll() ([]coreconnection.Session, error) { return nil, nil }

// countingRepositoryStub wraps repositoryStub with a per-call CountFriends callback.
type countingRepositoryStub struct {
	repositoryStub
	countByCall func(userID int) int
}

// CountFriends delegates to countByCall when set, otherwise falls back to repositoryStub.
func (c countingRepositoryStub) CountFriends(_ context.Context, userID int) (int, error) {
	if c.countByCall != nil {
		return c.countByCall(userID), c.repositoryStub.callErr
	}
	return c.repositoryStub.friendCount, c.repositoryStub.callErr
}

// permissionCheckerStub provides deterministic permission resolution for tests.
type permissionCheckerStub struct {
	grants map[string]bool
}

// HasPermission reports whether the permission is present in the grants map.
func (p *permissionCheckerStub) HasPermission(_ context.Context, _ int, permission string) (bool, error) {
	return p.grants[permission], nil
}

// newTestService creates a test messenger service from stubs.
func newTestService(repo domain.Repository, sessions coreconnection.SessionRegistry, bus broadcast.Broadcaster) *messengerapplication.Service {
	return newTestServiceWithChecker(repo, sessions, bus, nil)
}

// newTestServiceWithChecker creates a test messenger service with optional permission checker.
func newTestServiceWithChecker(repo domain.Repository, sessions coreconnection.SessionRegistry, bus broadcast.Broadcaster, checker domain.PermissionChecker) *messengerapplication.Service {
	service, err := messengerapplication.NewService(repo, sessions, bus, messengerapplication.Config{
		MaxFriends:       2,
		MaxFriendsVIP:    5,
		FloodCooldownMs:  750,
		FloodViolations:  4,
		FloodMuteSeconds: 20,
	})
	if err != nil {
		panic("newTestServiceWithChecker: " + err.Error())
	}
	var firer func(sdk.Event)
	service.SetEventFirer(firer)
	if checker != nil {
		service.SetPermissionChecker(checker)
	}
	return service
}

// epoch returns a fixed timestamp for tests.
func epoch() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
