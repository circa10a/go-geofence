package cache

import (
	"context"
)

// Cache is an interface for caching ip addresses
type Cache interface {
	Get(context.Context, string) (bool, bool, error)
	Set(context.Context, string, bool) error
}
