package cache

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const (
	deleteExpiredCacheItemsInternal = 10 * time.Minute
)

// MemoryCache is used to store/fetch ip proximity from an in-memory cache.
type MemoryCache struct {
	memoryClient  *gocache.Cache
	memoryOptions *MemoryOptions
}

// MemoryOptions holds in-memory cache configuration parameters.
type MemoryOptions struct {
	TTL time.Duration
}

// NewRedisCache provides a new in-memory cache client.
func NewMemoryCache(memoryOptions *MemoryOptions) *MemoryCache {
	return &MemoryCache{
		memoryClient:  gocache.New(memoryOptions.TTL, deleteExpiredCacheItemsInternal),
		memoryOptions: memoryOptions,
	}
}

// Get gets value from the in-memory cache.
func (m *MemoryCache) Get(ctx context.Context, key string) (bool, bool, error) {
	if isIPAddressNear, found := m.memoryClient.Get(key); found {
		return isIPAddressNear.(bool), found, nil
	}
	return false, false, nil
}

// Set sets k/v in the in-memory cache.
func (m *MemoryCache) Set(ctx context.Context, key string, value bool) error {
	m.memoryClient.Set(key, value, m.memoryOptions.TTL)
	return nil
}
