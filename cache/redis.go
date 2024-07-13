package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v9"
)

type RedisCache struct {
	redisClient  *redis.Client
	redisOptions *RedisOptions
}

// RedisOptions holds redis configuration parameters.
type RedisOptions struct {
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
}

func NewRedisCache(redisOpts *RedisOptions) *RedisCache {
	return &RedisCache{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     redisOpts.Addr,
			Password: redisOpts.Password,
			DB:       redisOpts.DB,
		}),
		redisOptions: redisOpts,
	}
}

func (r *RedisCache) Get(ctx context.Context, key string) (bool, bool, error) {
	val, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		// If key is not in redis
		if err == redis.Nil {
			return false, false, nil
		}
		return false, false, err
	}
	isIPAddressNear, err := strconv.ParseBool(val)
	if err != nil {
		return false, false, err
	}

	return isIPAddressNear, true, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value bool) error {
	// Redis stores false as 0 for whatever reason, so we'll store as a string and parse out in cacheGet
	return r.redisClient.Set(ctx, key, strconv.FormatBool(value), r.redisOptions.TTL).Err()
}
