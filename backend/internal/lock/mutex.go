package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Mutex struct {
	client *redis.Client
	key    string
	ttl    time.Duration
}

func New(client *redis.Client, key string, ttl time.Duration) *Mutex {
	return &Mutex{client: client, key: "lock:" + key, ttl: ttl}
}

func (m *Mutex) Lock(ctx context.Context) (bool, error) {
	ok, err := m.client.SetNX(ctx, m.key, "1", m.ttl).Result()
	return ok, err
}

func (m *Mutex) Unlock(ctx context.Context) error {
	return m.client.Del(ctx, m.key).Err()
}

func WithLock(ctx context.Context, client *redis.Client, key string, ttl time.Duration, fn func() error) error {
	m := New(client, key, ttl)
	acquired, err := m.Lock(ctx)
	if err != nil {
		return err
	}
	if !acquired {
		return fmt.Errorf("resource locked")
	}
	defer m.Unlock(ctx)
	return fn()
}
