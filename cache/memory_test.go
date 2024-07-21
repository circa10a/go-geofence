package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryCache(t *testing.T) {
	tests := []struct {
		input *MemoryOptions
	}{
		{
			input: &MemoryOptions{
				TTL: 0,
			},
		},
		{
			input: &MemoryOptions{
				TTL: time.Second * 1,
			},
		},
	}
	for _, test := range tests {
		actual := NewMemoryCache(test.input)
		assert.NotNil(t, actual)

		assert.Equal(t, actual.memoryOptions, test.input)
	}
}

func TestMemoryGetAndSet(t *testing.T) {
	tests := []struct {
		input struct {
			key   string
			value bool
		}
		exists   bool
		expected bool
	}{
		// Ensure key not present
		{
			input: struct {
				key   string
				value bool
			}{
				key:   "testkey1",
				value: true,
			},
			exists:   false,
			expected: false,
		},
		// Ensure value was set to true
		{
			input: struct {
				key   string
				value bool
			}{
				key:   "testkey2",
				value: true,
			},
			exists:   true,
			expected: true,
		},
		// Ensure value was set to false
		{
			input: struct {
				key   string
				value bool
			}{
				key:   "testkey3",
				value: false,
			},
			exists:   true,
			expected: false,
		},
	}
	for _, test := range tests {
		client := NewMemoryCache(&MemoryOptions{})
		assert.NotNil(t, client)

		if test.exists {
			client.Set(context.TODO(), test.input.key, test.input.value)
		}

		val, exists, err := client.Get(context.TODO(), test.input.key)
		assert.Equal(t, test.expected, val)
		assert.Equal(t, test.exists, exists)
		assert.NoError(t, err)
	}
}
