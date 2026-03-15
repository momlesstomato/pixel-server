package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
	corepermission "github.com/momlesstomato/pixel-server/core/permission"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	redislib "github.com/redis/go-redis/v9"
)

// LiveUpdater defines live packet update behavior after assignment changes.
type LiveUpdater interface {
	// PushAccessUpdate publishes updated access packets for one user.
	PushAccessUpdate(context.Context, permissiondomain.Access, []permissiondomain.PerkGrant) error
}

// Config defines permission service cache and runtime behavior.
type Config struct {
	// CachePrefix stores Redis key prefix for cached group snapshots.
	CachePrefix string
	// CacheTTL stores Redis cache TTL for group snapshots.
	CacheTTL time.Duration
	// AmbassadorPermission stores dotted permission string that grants ambassador state.
	AmbassadorPermission string
}

// Service defines permission and group application use-cases.
type Service struct {
	// repository stores permission persistence behavior.
	repository permissiondomain.Repository
	// redis stores optional cache client.
	redis *redislib.Client
	// cachePrefix stores Redis group cache prefix.
	cachePrefix string
	// cacheTTL stores Redis group cache TTL.
	cacheTTL time.Duration
	// ambassadorPermission stores permission that grants ambassador flag.
	ambassadorPermission string
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
	// liveUpdater stores optional live packet publish behavior.
	liveUpdater LiveUpdater
}

// NewService creates one permission service.
func NewService(repository permissiondomain.Repository, redis *redislib.Client, config Config) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("permission repository is required")
	}
	cachePrefix := strings.TrimSpace(config.CachePrefix)
	if cachePrefix == "" {
		cachePrefix = "perm:group"
	}
	cacheTTL := config.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}
	ambassadorPermission := strings.TrimSpace(strings.ToLower(config.AmbassadorPermission))
	if ambassadorPermission == "" {
		ambassadorPermission = "role.ambassador"
	}
	return &Service{
		repository: repository, redis: redis, cachePrefix: cachePrefix, cacheTTL: cacheTTL,
		ambassadorPermission: ambassadorPermission,
	}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// SetLiveUpdater configures optional live packet publish behavior.
func (service *Service) SetLiveUpdater(liveUpdater LiveUpdater) {
	service.liveUpdater = liveUpdater
}

// HasPermission resolves whether one user has one permission.
func (service *Service) HasPermission(ctx context.Context, userID int, permission string) (bool, error) {
	access, err := service.ResolveAccess(ctx, userID)
	if err != nil {
		return false, err
	}
	return corepermission.Resolve(access.Permissions, permission), nil
}

// EffectiveGroup resolves one user's effective group as plugin group info.
func (service *Service) EffectiveGroup(ctx context.Context, userID int) (sdk.GroupInfo, bool, error) {
	access, err := service.ResolveAccess(ctx, userID)
	if err != nil {
		return sdk.GroupInfo{}, false, err
	}
	group := access.PrimaryGroup
	return sdk.GroupInfo{
		ID: group.ID, Name: group.Name, ClubLevel: group.ClubLevel,
		SecurityLevel: group.SecurityLevel, IsAmbassador: group.IsAmbassador,
	}, group.ID > 0, nil
}

// cachedGroup defines one Redis-cached group snapshot payload.
type cachedGroup struct {
	// Group stores cached group attributes.
	Group permissiondomain.Group `json:"group"`
	// Permissions stores cached permission grants.
	Permissions []string `json:"permissions"`
}

// groupCacheKey returns Redis cache key for one group identifier.
func (service *Service) groupCacheKey(groupID int) string {
	return fmt.Sprintf("%s:%d", service.cachePrefix, groupID)
}

// loadCachedGroup resolves one group snapshot from cache.
func (service *Service) loadCachedGroup(ctx context.Context, groupID int) (cachedGroup, bool) {
	if service.redis == nil {
		return cachedGroup{}, false
	}
	payload, err := service.redis.Get(ctx, service.groupCacheKey(groupID)).Bytes()
	if err != nil {
		return cachedGroup{}, false
	}
	var cached cachedGroup
	if json.Unmarshal(payload, &cached) != nil {
		return cachedGroup{}, false
	}
	return cached, true
}

// storeCachedGroup stores one group snapshot in Redis.
func (service *Service) storeCachedGroup(ctx context.Context, groupID int, value cachedGroup) {
	if service.redis == nil {
		return
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return
	}
	service.redis.Set(ctx, service.groupCacheKey(groupID), payload, service.cacheTTL)
}
