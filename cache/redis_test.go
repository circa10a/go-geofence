package cache

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisCache(t *testing.T) {
	tests := []struct {
		input *RedisOptions
	}{
		{
			input: &RedisOptions{
				Addr:     "testaddr",
				Username: "testusername",
				Password: "testpassword",
				DB:       1,
			},
		},
		{
			input: &RedisOptions{
				Addr: "localhost:6379", // If empty, client auto sets this
			},
		},
	}
	for _, test := range tests {
		actual := NewRedisCache(test.input)
		actualRedisClientOpts := actual.redisClient.Options()
		assert.NotNil(t, actual)

		assert.Equal(t, actual.redisOptions, actual.redisOptions)

		assert.Equal(t, actualRedisClientOpts.Addr, test.input.Addr)
		assert.Equal(t, actualRedisClientOpts.Username, test.input.Username)
		assert.Equal(t, actualRedisClientOpts.Password, test.input.Password)
		assert.Equal(t, actualRedisClientOpts.DB, test.input.DB)
	}
}
func TestRedisGet(t *testing.T) {
	tests := []struct {
		key   string
		value bool

		exists bool
	}{
		// Happy path. Set key, get value
		// Ensure true is returned
		{
			key:    "testkey1",
			value:  true,
			exists: true,
		},
		// Don't set key, ensure exists value is returned properly
		{
			key:    "testkey2",
			value:  false,
			exists: false,
		},
		// Ensure false is returned
		{
			key:    "testkey3",
			value:  false,
			exists: true,
		},
	}
	for _, test := range tests {
		ttl := time.Second * 5
		client := NewRedisCache(&RedisOptions{TTL: ttl})
		db, mock := redismock.NewClientMock()

		// Overide real client with mock client
		client.redisClient = db
		setVal := strconv.FormatBool(test.value)

		// Set key if expected to be present
		if test.exists {
			mock.ExpectSet(test.key, setVal, ttl).SetVal(setVal)
			mock.ExpectGet(test.key).SetVal(setVal)

			err := client.Set(context.TODO(), test.key, test.value)
			assert.NoError(t, err)
		} else {
			mock.ExpectGet(test.key).RedisNil()
		}

		val, exists, err := client.Get(context.TODO(), test.key)

		assert.NoError(t, err)
		assert.Equal(t, test.value, val)
		assert.Equal(t, test.exists, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestRedisSet(t *testing.T) {
	tests := []struct {
		key   string
		value bool
	}{
		{
			key:   "testkey1",
			value: true,
		},
		{
			key:   "testkey2",
			value: false,
		},
		// Ensure false is set correctly
		{
			key:   "testkey3",
			value: false,
		},
	}
	for _, test := range tests {
		ttl := time.Second * 5
		client := NewRedisCache(&RedisOptions{TTL: ttl})
		db, mock := redismock.NewClientMock()

		// Overide real client with mock client
		client.redisClient = db
		setVal := strconv.FormatBool(test.value)

		mock.ExpectSet(test.key, setVal, ttl).SetVal(setVal)

		err := client.Set(context.TODO(), test.key, test.value)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
