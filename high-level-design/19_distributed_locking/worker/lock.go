package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLockNotAcquired = errors.New("lock not acquired")

// Lock represents a held distributed lock.
type Lock struct {
	rdb   *redis.Client
	key   string
	token string // random token to ensure only the owner can release
}

// TryAcquire attempts to acquire the lock. Returns ErrLockNotAcquired if held.
// ttl: how long the lock lives (auto-expire prevents deadlocks).
func TryAcquire(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) (*Lock, error) {
	token := randomToken()
	ok, err := rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLockNotAcquired
	}
	return &Lock{rdb: rdb, key: key, token: token}, nil
}

// Release releases the lock only if this holder owns it (compare-and-delete via Lua).
func (l *Lock) Release(ctx context.Context) error {
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)
	return script.Run(ctx, l.rdb, []string{l.key}, l.token).Err()
}

func randomToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
