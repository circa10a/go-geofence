package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is used to store/fetch ip proximity from redis.
type RedisCache struct {
	redisClient  *redis.Client
	redisOptions *RedisOptions
}

// RedisOptions holds redis configuration parameters.
type RedisOptions struct {
	Addr     string
	Username string
	Password string
	DB       int
	TTL      time.Duration
}

// NewRedisCache provides a new redis cache client.
func NewRedisCache(redisOpts *RedisOptions) *RedisCache {
	return &RedisCache{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     redisOpts.Addr,
			Username: redisOpts.Username,
			Password: redisOpts.Password,
			DB:       redisOpts.DB,
		}),
		redisOptions: redisOpts,
	}
}

// Get gets value from redis.
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

// Set sets k/v in redis.
func (r *RedisCache) Set(ctx context.Context, key string, value bool) error {
	// Redis stores false as 0 for whatever reason, so we'll store as a string and parse it out
	return r.redisClient.Set(ctx, key, strconv.FormatBool(value), r.redisOptions.TTL).Err()
}
