package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	redislib "github.com/redis/go-redis/v9"
)

// CacheConfig defines catalog Redis cache configuration.
type CacheConfig struct {
	// Prefix stores Redis key namespace prefix.
	Prefix string
	// TTL stores cache entry time-to-live duration.
	TTL time.Duration
}

// cachedPages is the JSON-serializable envelope for catalog page lists.
type cachedPages struct {
	Pages []domain.CatalogPage `json:"pages"`
}

// cachedOffers is the JSON-serializable envelope for catalog offer lists.
type cachedOffers struct {
	Offers []domain.CatalogOffer `json:"offers"`
}

// SetCache configures optional Redis cache for catalog data.
func (service *Service) SetCache(redis *redislib.Client, cfg CacheConfig) {
	service.redis = redis
	service.cachePrefix = cfg.Prefix
	service.cacheTTL = cfg.TTL
}

// allPagesCacheKey returns the Redis key for the full page list.
func (service *Service) allPagesCacheKey() string {
	return fmt.Sprintf("%s:pages", service.cachePrefix)
}

// offersCacheKey returns the Redis key for one page offer list.
func (service *Service) offersCacheKey(pageID int) string {
	return fmt.Sprintf("%s:offers:%d", service.cachePrefix, pageID)
}

// loadCachedPages reads the page list from Redis; returns false on any miss.
func (service *Service) loadCachedPages(ctx context.Context) ([]domain.CatalogPage, bool) {
	if service.redis == nil {
		return nil, false
	}
	raw, err := service.redis.Get(ctx, service.allPagesCacheKey()).Bytes()
	if err != nil {
		return nil, false
	}
	var cached cachedPages
	if json.Unmarshal(raw, &cached) != nil {
		return nil, false
	}
	return cached.Pages, true
}

// storeCachedPages writes the page list to Redis; silent on serialization error.
func (service *Service) storeCachedPages(ctx context.Context, pages []domain.CatalogPage) {
	if service.redis == nil {
		return
	}
	raw, err := json.Marshal(cachedPages{Pages: pages})
	if err != nil {
		return
	}
	_ = service.redis.Set(ctx, service.allPagesCacheKey(), raw, service.cacheTTL)
}

// loadCachedOffers reads one page offer list from Redis; returns false on any miss.
func (service *Service) loadCachedOffers(ctx context.Context, pageID int) ([]domain.CatalogOffer, bool) {
	if service.redis == nil {
		return nil, false
	}
	raw, err := service.redis.Get(ctx, service.offersCacheKey(pageID)).Bytes()
	if err != nil {
		return nil, false
	}
	var cached cachedOffers
	if json.Unmarshal(raw, &cached) != nil {
		return nil, false
	}
	return cached.Offers, true
}

// storeCachedOffers writes one page offer list to Redis; silent on serialization error.
func (service *Service) storeCachedOffers(ctx context.Context, pageID int, offers []domain.CatalogOffer) {
	if service.redis == nil {
		return
	}
	raw, err := json.Marshal(cachedOffers{Offers: offers})
	if err != nil {
		return
	}
	_ = service.redis.Set(ctx, service.offersCacheKey(pageID), raw, service.cacheTTL)
}

// invalidatePages removes the full page list from Redis.
func (service *Service) invalidatePages(ctx context.Context) {
	if service.redis == nil {
		return
	}
	_ = service.redis.Del(ctx, service.allPagesCacheKey())
}

// invalidateOffers removes one page offer list from Redis.
func (service *Service) invalidateOffers(ctx context.Context, pageID int) {
	if service.redis == nil {
		return
	}
	_ = service.redis.Del(ctx, service.offersCacheKey(pageID))
}
