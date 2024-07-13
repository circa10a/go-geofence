package cache

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const (
	deleteExpiredCacheItemsInternal = 10 * time.Minute
)

type MemoryCache struct {
	memoryClient  *gocache.Cache
	memoryOptions *MemoryOptions
}

// MemoryOptions holds in memory cache configuration parameters.
type MemoryOptions struct {
	TTL time.Duration
}

func NewMemoryCache(memoryOptions *MemoryOptions) *MemoryCache {
	return &MemoryCache{
		memoryClient:  gocache.New(memoryOptions.TTL, deleteExpiredCacheItemsInternal),
		memoryOptions: memoryOptions,
	}
}

func (m *MemoryCache) Get(ctx context.Context, key string) (bool, bool, error) {
	if isIPAddressNear, found := m.memoryClient.Get(key); found {
		return isIPAddressNear.(bool), found, nil
	} else {
		return false, false, nil
	}
}

func (m *MemoryCache) Set(ctx context.Context, key string, value bool) error {
	m.memoryClient.Set(key, value, m.memoryOptions.TTL)
	return nil
}
