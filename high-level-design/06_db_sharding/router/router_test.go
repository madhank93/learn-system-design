package main

import (
	"fmt"
	"testing"
)

var testShards = []string{"shard1:8080", "shard2:8080", "shard3:8080"}

func setupGlobals() {
	shards = testShards
}

// TestHashShardDeterministic verifies that calling hashShard with the same key
// 100 times always returns the identical shard.
func TestHashShardDeterministic(t *testing.T) {
	setupGlobals()
	key := "user-42"
	first := hashShard(key)
	for i := 0; i < 100; i++ {
		if got := hashShard(key); got != first {
			t.Fatalf("call %d: hashShard(%q) = %q, want %q", i, key, got, first)
		}
	}
}

// TestHashShardDistribution verifies that 100 sequential numeric keys spread
// across all three shards (no shard receives zero keys).
func TestHashShardDistribution(t *testing.T) {
	setupGlobals()
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		shard := hashShard(fmt.Sprintf("%d", i))
		counts[shard]++
	}
	for _, s := range shards {
		if counts[s] == 0 {
			t.Errorf("shard %q received no keys out of 100", s)
		}
	}
}

// TestRangeShardBoundaries validates the documented bucket boundaries:
//
//	"0"   → shards[0]   (0 / 333 = 0)
//	"333" → shards[1]   (333 / 333 = 1)
//	"666" → shards[2]   (666 / 333 = 2)
//	"999" → shards[2]   (999 / 333 = 3, clamped to 2)
func TestRangeShardBoundaries(t *testing.T) {
	setupGlobals()
	cases := []struct {
		key  string
		want string
	}{
		{"0", shards[0]},
		{"333", shards[1]},
		{"666", shards[2]},
		{"999", shards[2]},
	}
	for _, tc := range cases {
		if got := rangeShard(tc.key); got != tc.want {
			t.Errorf("rangeShard(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

// TestRangeShardNonNumeric verifies that a non-numeric key falls back to shards[0].
func TestRangeShardNonNumeric(t *testing.T) {
	setupGlobals()
	if got := rangeShard("foo"); got != shards[0] {
		t.Errorf("rangeShard(%q) = %q, want %q (shards[0])", "foo", got, shards[0])
	}
}

// TestRouteHashStrategy confirms that with strategy="hash", route() delegates
// to hashShard and returns the same result as calling hashShard directly.
func TestRouteHashStrategy(t *testing.T) {
	setupGlobals()
	strategy = "hash"
	keys := []string{"alpha", "beta", "gamma", "42", "hello-world"}
	for _, key := range keys {
		want := hashShard(key)
		if got := route(key); got != want {
			t.Errorf("strategy=hash: route(%q) = %q, want %q", key, got, want)
		}
	}
}

// TestRouteRangeStrategy confirms that with strategy="range", route() delegates
// to rangeShard and returns the same result as calling rangeShard directly.
func TestRouteRangeStrategy(t *testing.T) {
	setupGlobals()
	strategy = "range"
	keys := []string{"0", "200", "333", "500", "666", "999", "foo"}
	for _, key := range keys {
		want := rangeShard(key)
		if got := route(key); got != want {
			t.Errorf("strategy=range: route(%q) = %q, want %q", key, got, want)
		}
	}
}
