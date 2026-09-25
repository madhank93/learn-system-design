package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var ErrNotFound = errors.New("short code not found")

type Shortener struct {
	rdb *redis.Client
}

func NewShortener(rdb *redis.Client) *Shortener {
	return &Shortener{rdb: rdb}
}

// encodeBase62 converts a uint64 to a base-62 string of fixed length 7.
func encodeBase62(n uint64) string {
	code := make([]byte, 7)
	for i := 6; i >= 0; i-- {
		code[i] = base62[n%62]
		n /= 62
	}
	return string(code)
}

// hash generates a 7-char base-62 code from the URL.
func hash(url string) string {
	h := md5.Sum([]byte(url))
	var n uint64
	for i := 0; i < 8; i++ {
		n = n<<8 | uint64(h[i])
	}
	return encodeBase62(n)
}

// Shorten stores the URL and returns a short code. On collision, appends a counter.
func (s *Shortener) Shorten(ctx context.Context, url string) (string, error) {
	code := hash(url)

	// Check if code already maps to same URL (idempotent)
	existing, err := s.rdb.Get(ctx, "url:"+code).Result()
	if err == nil && existing == url {
		return code, nil
	}

	// Collision: try up to 5 variants
	for i := 0; i < 5; i++ {
		attempt := hash(fmt.Sprintf("%s%d", url, i))
		exists, _ := s.rdb.Exists(ctx, "url:"+attempt).Result()
		if exists == 0 {
			code = attempt
			break
		}
	}

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, "url:"+code, url, 30*24*time.Hour) // 30-day TTL
	pipe.Incr(ctx, "clicks:"+code)
	pipe.Set(ctx, "clicks:"+code, 0, 30*24*time.Hour)
	pipe.Exec(ctx)

	return code, nil
}

// Resolve returns the original URL for a short code and increments click count.
func (s *Shortener) Resolve(ctx context.Context, code string) (string, error) {
	url, err := s.rdb.Get(ctx, "url:"+code).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	s.rdb.Incr(ctx, "clicks:"+code)
	return url, nil
}

// Stats returns click count for a code.
func (s *Shortener) Stats(ctx context.Context, code string) (int64, error) {
	n, err := s.rdb.Get(ctx, "clicks:"+code).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, ErrNotFound
	}
	return n, err
}
