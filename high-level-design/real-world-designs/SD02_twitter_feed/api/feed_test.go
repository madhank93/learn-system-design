package main

import (
	"testing"
)

func TestTweetKeyFormat(t *testing.T) {
	k := tweetKey("tweet-123")
	if k != "tweet:tweet-123" {
		t.Errorf("expected tweet:tweet-123, got %q", k)
	}
}

func TestTimelineKeyFormat(t *testing.T) {
	k := timelineKey("alice")
	if k != "timeline:alice" {
		t.Errorf("expected timeline:alice, got %q", k)
	}
}

func TestFollowersKeyFormat(t *testing.T) {
	k := followersKey("alice")
	if k != "followers:alice" {
		t.Errorf("expected followers:alice, got %q", k)
	}
}

func TestFollowingKeyFormat(t *testing.T) {
	k := followingKey("bob")
	if k != "following:bob" {
		t.Errorf("expected following:bob, got %q", k)
	}
}
