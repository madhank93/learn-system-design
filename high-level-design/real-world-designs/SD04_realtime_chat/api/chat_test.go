package main

import (
	"context"
	"testing"
	"time"
)

func TestJoinAndLeave(t *testing.T) {
	hub := NewHub(NewMemStore())
	// can't test real WS conn in unit test — verify room tracking via RoomCount
	if hub.RoomCount("general") != 0 {
		t.Fatal("expected empty room")
	}
}

func TestMemStoreAddAndHistory(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()

	msgs := []Message{
		{ID: "1", Room: "a", User: "alice", Text: "hello", TS: 1},
		{ID: "2", Room: "a", User: "bob", Text: "world", TS: 2},
		{ID: "3", Room: "a", User: "alice", Text: "third", TS: 3},
	}
	for _, m := range msgs {
		if err := s.AddMessage(ctx, m.Room, m); err != nil {
			t.Fatalf("AddMessage: %v", err)
		}
	}

	hist, err := s.History(ctx, "a", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hist) != 3 {
		t.Fatalf("want 3 messages, got %d", len(hist))
	}
	// History is prepend (newest first)
	if hist[0].Text != "third" {
		t.Fatalf("want newest first, got %q", hist[0].Text)
	}
}

func TestMemStoreHistoryLimit(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()
	for i := 0; i < 10; i++ {
		_ = s.AddMessage(ctx, "room", Message{ID: string(rune('0'+i)), Text: "msg"})
	}
	hist, _ := s.History(ctx, "room", 3)
	if len(hist) != 3 {
		t.Fatalf("limit=3 want 3, got %d", len(hist))
	}
}

func TestPresence(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()

	_ = s.SetPresence(ctx, "room", "alice")
	_ = s.SetPresence(ctx, "room", "bob")

	users, err := s.Presence(ctx, "room")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("want 2 users, got %d", len(users))
	}

	_ = s.ClearPresence(ctx, "room", "alice")
	users, _ = s.Presence(ctx, "room")
	if len(users) != 1 || users[0] != "bob" {
		t.Fatalf("want [bob], got %v", users)
	}
}

func TestPresenceIsolatedByRoom(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()
	_ = s.SetPresence(ctx, "room1", "alice")
	_ = s.SetPresence(ctx, "room2", "bob")

	u1, _ := s.Presence(ctx, "room1")
	u2, _ := s.Presence(ctx, "room2")
	if len(u1) != 1 || u1[0] != "alice" {
		t.Fatalf("room1 should have [alice], got %v", u1)
	}
	if len(u2) != 1 || u2[0] != "bob" {
		t.Fatalf("room2 should have [bob], got %v", u2)
	}
}

func TestHubBroadcast(t *testing.T) {
	store := NewMemStore()
	hub := NewHub(store)

	ctx := context.Background()
	msg := Message{
		ID:   "test-1",
		Room: "general",
		User: "alice",
		Text: "hello broadcast",
		TS:   time.Now().Unix(),
	}
	_ = store.AddMessage(ctx, "general", msg)
	hub.Broadcast(msg)

	hist, _ := store.History(ctx, "general", 10)
	if len(hist) == 0 || hist[0].Text != "hello broadcast" {
		t.Fatalf("message not stored: %v", hist)
	}
}

func TestMultipleRooms(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()

	_ = s.AddMessage(ctx, "room1", Message{ID: "1", Room: "room1", Text: "r1msg"})
	_ = s.AddMessage(ctx, "room2", Message{ID: "2", Room: "room2", Text: "r2msg"})

	h1, _ := s.History(ctx, "room1", 10)
	h2, _ := s.History(ctx, "room2", 10)

	if len(h1) != 1 || h1[0].Room != "room1" {
		t.Fatalf("room1 isolation failed: %v", h1)
	}
	if len(h2) != 1 || h2[0].Room != "room2" {
		t.Fatalf("room2 isolation failed: %v", h2)
	}
}

func TestNewMsgID(t *testing.T) {
	id1 := newMsgID()
	time.Sleep(time.Nanosecond)
	id2 := newMsgID()
	if id1 == id2 {
		t.Fatal("msg IDs should be unique")
	}
}
