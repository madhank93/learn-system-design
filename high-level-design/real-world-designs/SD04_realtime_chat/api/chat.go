package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

type Message struct {
	ID   string `json:"id"`
	Room string `json:"room"`
	User string `json:"user"`
	Text string `json:"text"`
	TS   int64  `json:"ts"`
}

// Store abstracts message history and presence — Redis in prod, MemStore in tests.
type Store interface {
	AddMessage(ctx context.Context, room string, msg Message) error
	History(ctx context.Context, room string, n int) ([]Message, error)
	SetPresence(ctx context.Context, room, user string) error
	ClearPresence(ctx context.Context, room, user string) error
	Presence(ctx context.Context, room string) ([]string, error)
}

// ── Redis store ───────────────────────────────────────────────

type RedisStore struct{ rdb *redis.Client }

func NewRedisStore(rdb *redis.Client) *RedisStore { return &RedisStore{rdb} }

func (s *RedisStore) AddMessage(ctx context.Context, room string, msg Message) error {
	b, _ := json.Marshal(msg)
	return s.rdb.LPush(ctx, "msgs:"+room, b).Err()
}

func (s *RedisStore) History(ctx context.Context, room string, n int) ([]Message, error) {
	raw, err := s.rdb.LRange(ctx, "msgs:"+room, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	msgs := make([]Message, 0, len(raw))
	for _, r := range raw {
		var m Message
		if json.Unmarshal([]byte(r), &m) == nil {
			msgs = append(msgs, m)
		}
	}
	return msgs, nil
}

func (s *RedisStore) SetPresence(ctx context.Context, room, user string) error {
	return s.rdb.SAdd(ctx, "presence:"+room, user).Err()
}

func (s *RedisStore) ClearPresence(ctx context.Context, room, user string) error {
	return s.rdb.SRem(ctx, "presence:"+room, user).Err()
}

func (s *RedisStore) Presence(ctx context.Context, room string) ([]string, error) {
	return s.rdb.SMembers(ctx, "presence:"+room).Result()
}

// ── In-memory store (tests) ───────────────────────────────────

type MemStore struct {
	mu       sync.RWMutex
	msgs     map[string][]Message
	presence map[string]map[string]struct{}
}

func NewMemStore() *MemStore {
	return &MemStore{
		msgs:     make(map[string][]Message),
		presence: make(map[string]map[string]struct{}),
	}
}

func (s *MemStore) AddMessage(_ context.Context, room string, msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgs[room] = append([]Message{msg}, s.msgs[room]...)
	return nil
}

func (s *MemStore) History(_ context.Context, room string, n int) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs := s.msgs[room]
	if n < len(msgs) {
		msgs = msgs[:n]
	}
	return msgs, nil
}

func (s *MemStore) SetPresence(_ context.Context, room, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.presence[room] == nil {
		s.presence[room] = make(map[string]struct{})
	}
	s.presence[room][user] = struct{}{}
	return nil
}

func (s *MemStore) ClearPresence(_ context.Context, room, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.presence[room], user)
	return nil
}

func (s *MemStore) Presence(_ context.Context, room string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]string, 0, len(s.presence[room]))
	for u := range s.presence[room] {
		users = append(users, u)
	}
	return users, nil
}

// ── Hub — manages rooms and WebSocket connections ─────────────

type client struct {
	conn *websocket.Conn
	send chan Message
}

type Hub struct {
	mu     sync.RWMutex
	rooms  map[string]map[string]*client // room → user → client
	store  Store
	closed bool
}

func NewHub(store Store) *Hub {
	return &Hub{
		rooms: make(map[string]map[string]*client),
		store: store,
	}
}

func (h *Hub) Join(room, user string, conn *websocket.Conn) *client {
	c := &client{conn: conn, send: make(chan Message, 32)}
	h.mu.Lock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[string]*client)
	}
	h.rooms[room][user] = c
	h.mu.Unlock()

	ctx := context.Background()
	_ = h.store.SetPresence(ctx, room, user)

	// pump outgoing messages to WebSocket
	go func() {
		for msg := range c.send {
			b, _ := json.Marshal(msg)
			_ = websocket.Message.Send(conn, string(b))
		}
	}()
	return c
}

func (h *Hub) Leave(room, user string) {
	h.mu.Lock()
	if clients, ok := h.rooms[room]; ok {
		if c, ok := clients[user]; ok {
			close(c.send)
			delete(clients, user)
		}
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}
	h.mu.Unlock()
	_ = h.store.ClearPresence(context.Background(), room, user)
}

func (h *Hub) Broadcast(msg Message) {
	h.mu.RLock()
	clients := h.rooms[msg.Room]
	h.mu.RUnlock()
	for _, c := range clients {
		select {
		case c.send <- msg:
		default: // drop slow receiver
		}
	}
}

func (h *Hub) ListRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rooms := make([]string, 0, len(h.rooms))
	for r := range h.rooms {
		rooms = append(rooms, r)
	}
	return rooms
}

func (h *Hub) RoomCount(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[room])
}

func newMsgID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
