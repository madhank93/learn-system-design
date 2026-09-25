package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Tweet struct {
	ID     string    `json:"id"`
	Author string    `json:"author"`
	Text   string    `json:"text"`
	At     time.Time `json:"at"`
}

type Feed struct {
	rdb *redis.Client
}

func NewFeed(rdb *redis.Client) *Feed {
	return &Feed{rdb: rdb}
}

func tweetKey(id string) string    { return "tweet:" + id }
func timelineKey(u string) string  { return "timeline:" + u }
func followersKey(u string) string { return "followers:" + u }
func followingKey(u string) string { return "following:" + u }

// Post stores a tweet and fans out to all followers' timelines.
func (f *Feed) Post(ctx context.Context, author, text string) (*Tweet, error) {
	id := fmt.Sprintf("%s-%d", author, time.Now().UnixNano())
	t := Tweet{ID: id, Author: author, Text: text, At: time.Now()}
	data, _ := json.Marshal(t)

	pipe := f.rdb.Pipeline()
	// Store tweet content
	pipe.Set(ctx, tweetKey(id), data, 7*24*time.Hour)
	// Add to author's own timeline
	pipe.ZAdd(ctx, timelineKey(author), redis.Z{Score: float64(t.At.UnixNano()), Member: id})
	pipe.Exec(ctx)

	// Fanout to followers
	followers, _ := f.rdb.SMembers(ctx, followersKey(author)).Result()
	if len(followers) > 0 {
		pipe2 := f.rdb.Pipeline()
		for _, follower := range followers {
			pipe2.ZAdd(ctx, timelineKey(follower), redis.Z{
				Score:  float64(t.At.UnixNano()),
				Member: id,
			})
			// Keep only last 200 tweets per timeline
			pipe2.ZRemRangeByRank(ctx, timelineKey(follower), 0, -201)
		}
		pipe2.Exec(ctx)
	}
	return &t, nil
}

// Timeline returns the latest N tweets for a user.
func (f *Feed) Timeline(ctx context.Context, user string, limit int) ([]*Tweet, error) {
	ids, err := f.rdb.ZRevRange(ctx, timelineKey(user), 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	tweets := make([]*Tweet, 0, len(ids))
	for _, id := range ids {
		data, err := f.rdb.Get(ctx, tweetKey(id)).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		var t Tweet
		if err := json.Unmarshal([]byte(data), &t); err == nil {
			tweets = append(tweets, &t)
		}
	}
	return tweets, nil
}

// Follow makes follower follow followee, and adds followee's recent tweets.
func (f *Feed) Follow(ctx context.Context, follower, followee string) error {
	pipe := f.rdb.Pipeline()
	pipe.SAdd(ctx, followersKey(followee), follower)
	pipe.SAdd(ctx, followingKey(follower), followee)
	pipe.Exec(ctx)

	// Backfill recent tweets from followee
	recent, _ := f.rdb.ZRevRange(ctx, timelineKey(followee), 0, 19).Result()
	if len(recent) > 0 {
		zs := make([]redis.Z, len(recent))
		for i, id := range recent {
			score, _ := f.rdb.ZScore(ctx, timelineKey(followee), id).Result()
			zs[i] = redis.Z{Score: score, Member: id}
		}
		f.rdb.ZAdd(ctx, timelineKey(follower), zs...)
	}
	return nil
}

// Followers returns the follower count for a user.
func (f *Feed) FollowerCount(ctx context.Context, user string) int64 {
	n, _ := f.rdb.SCard(ctx, followersKey(user)).Result()
	return n
}
