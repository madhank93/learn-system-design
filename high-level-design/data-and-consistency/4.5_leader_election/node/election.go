package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	leaderKey = "leader"
	leaseTTL  = 5 * time.Second
	heartbeat = 2 * time.Second
	campaign  = 500 * time.Millisecond
)

var ErrNotLeader = errors.New("not the leader")

type Node struct {
	id  string
	rdb *redis.Client
}

func NewNode(id, redisAddr string) *Node {
	return &Node{
		id:  id,
		rdb: redis.NewClient(&redis.Options{Addr: redisAddr}),
	}
}

// TryAcquire attempts to become leader. Returns true on success.
func (n *Node) TryAcquire(ctx context.Context) bool {
	ok, err := n.rdb.SetNX(ctx, leaderKey, n.id, leaseTTL).Result()
	return err == nil && ok
}

// Renew extends the lease if this node is still leader.
func (n *Node) Renew(ctx context.Context) error {
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)
	res, err := script.Run(ctx, n.rdb, []string{leaderKey}, n.id, int(leaseTTL.Milliseconds())).Int()
	if err != nil || res == 0 {
		return ErrNotLeader
	}
	return nil
}

// CurrentLeader returns which node currently holds the lease.
func (n *Node) CurrentLeader(ctx context.Context) string {
	val, err := n.rdb.Get(ctx, leaderKey).Result()
	if errors.Is(err, redis.Nil) {
		return "none"
	}
	if err != nil {
		return "error"
	}
	return val
}

// Run is the main election loop. Campaigns until leader, then renews until lost.
func (n *Node) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if n.TryAcquire(ctx) {
			log.Printf("[%s] became LEADER", n.id)
			n.leadLoop(ctx)
			log.Printf("[%s] lost leadership", n.id)
		} else {
			leader := n.CurrentLeader(ctx)
			log.Printf("[%s] FOLLOWER — current leader: %s", n.id, leader)
			time.Sleep(campaign)
		}
	}
}

func (n *Node) leadLoop(ctx context.Context) {
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := n.Renew(ctx); err != nil {
				return
			}
			log.Printf("[%s] renewed lease", n.id)
		}
	}
}

// Status returns a human-readable status string.
func (n *Node) Status(ctx context.Context) string {
	leader := n.CurrentLeader(ctx)
	if leader == n.id {
		return fmt.Sprintf("node:%s role:LEADER", n.id)
	}
	return fmt.Sprintf("node:%s role:FOLLOWER leader:%s", n.id, leader)
}
