package redisstore

import (
	"context"
	"fmt"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	Session   *redis.Client
	Cart      *redis.Client
	RateLimit *redis.Client
}

func Connect(cfg *config.Config) (*Store, error) {
	session := newClient(cfg, cfg.RedisSessionDB)
	cart := newClient(cfg, cfg.RedisCartDB)
	rate := newClient(cfg, cfg.RedisRateLimitDB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for name, c := range map[string]*redis.Client{"session": session, "cart": cart, "ratelimit": rate} {
		if err := c.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("redis %s ping: %w", name, err)
		}
	}
	return &Store{Session: session, Cart: cart, RateLimit: rate}, nil
}

func newClient(cfg *config.Config, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       db,
	})
}

func (s *Store) Close() {
	s.Session.Close()
	s.Cart.Close()
	s.RateLimit.Close()
}
