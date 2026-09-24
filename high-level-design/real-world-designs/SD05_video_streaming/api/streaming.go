package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

var videoSeq uint64

const chunkSize = 1024 * 256 // 256 KB simulated chunk

// Video metadata stored in Redis hash.
type Video struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Owner     string   `json:"owner"`
	SizeBytes int      `json:"size_bytes"`
	Chunks    int      `json:"chunks"`
	Resols    []string `json:"resolutions"` // ["360p","720p","1080p"]
	Views     int64    `json:"views"`
	CreatedAt int64    `json:"created_at"`
}

type Chunk struct {
	VideoID  string `json:"video_id"`
	Num      int    `json:"num"`
	Total    int    `json:"total"`
	Resoln   string `json:"resolution"`
	Data     string `json:"data"` // base64 sim — just a placeholder string
	Duration int    `json:"duration_ms"`
}

type Manifest struct {
	VideoID     string   `json:"video_id"`
	Title       string   `json:"title"`
	TotalChunks int      `json:"total_chunks"`
	Resolutions []string `json:"resolutions"`
	ChunkURLs   []string `json:"chunk_urls"`
}

// Store abstracts video persistence — Redis in prod, MemVideoStore in tests.
type VideoStore interface {
	SaveVideo(ctx context.Context, v Video) error
	GetVideo(ctx context.Context, id string) (Video, error)
	IncrViews(ctx context.Context, id string) (int64, error)
	TopVideos(ctx context.Context, n int) ([]Video, error)
	AddToHistory(ctx context.Context, userID, videoID string) error
	GetHistory(ctx context.Context, userID string, n int) ([]string, error)
}

// ── Redis store ───────────────────────────────────────────────

type RedisVideoStore struct{ rdb *redis.Client }

func NewRedisVideoStore(rdb *redis.Client) *RedisVideoStore {
	return &RedisVideoStore{rdb}
}

func (s *RedisVideoStore) SaveVideo(ctx context.Context, v Video) error {
	b, _ := json.Marshal(v)
	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, "video:"+v.ID, b, 0)
	pipe.ZAdd(ctx, "videos:views", redis.Z{Score: float64(v.Views), Member: v.ID})
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisVideoStore) GetVideo(ctx context.Context, id string) (Video, error) {
	raw, err := s.rdb.Get(ctx, "video:"+id).Result()
	if err != nil {
		return Video{}, err
	}
	var v Video
	return v, json.Unmarshal([]byte(raw), &v)
}

func (s *RedisVideoStore) IncrViews(ctx context.Context, id string) (int64, error) {
	pipe := s.rdb.Pipeline()
	incr := pipe.HIncrBy(ctx, "views", id, 1)
	pipe.ZIncrBy(ctx, "videos:views", 1, id)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (s *RedisVideoStore) TopVideos(ctx context.Context, n int) ([]Video, error) {
	ids, err := s.rdb.ZRevRange(ctx, "videos:views", 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	videos := make([]Video, 0, len(ids))
	for _, id := range ids {
		v, err := s.GetVideo(ctx, id)
		if err == nil {
			videos = append(videos, v)
		}
	}
	return videos, nil
}

func (s *RedisVideoStore) AddToHistory(ctx context.Context, userID, videoID string) error {
	return s.rdb.LPush(ctx, "history:"+userID, videoID).Err()
}

func (s *RedisVideoStore) GetHistory(ctx context.Context, userID string, n int) ([]string, error) {
	return s.rdb.LRange(ctx, "history:"+userID, 0, int64(n-1)).Result()
}

// ── In-memory store (tests) ───────────────────────────────────

type MemVideoStore struct {
	mu      sync.RWMutex
	videos  map[string]Video
	views   map[string]int64
	history map[string][]string
}

func NewMemVideoStore() *MemVideoStore {
	return &MemVideoStore{
		videos:  make(map[string]Video),
		views:   make(map[string]int64),
		history: make(map[string][]string),
	}
}

func (s *MemVideoStore) SaveVideo(_ context.Context, v Video) error {
	s.mu.Lock()
	s.videos[v.ID] = v
	s.mu.Unlock()
	return nil
}

func (s *MemVideoStore) GetVideo(_ context.Context, id string) (Video, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.videos[id]
	if !ok {
		return Video{}, fmt.Errorf("video %s not found", id)
	}
	return v, nil
}

func (s *MemVideoStore) IncrViews(_ context.Context, id string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.views[id]++
	v := s.videos[id]
	v.Views = s.views[id]
	s.videos[id] = v
	return s.views[id], nil
}

func (s *MemVideoStore) TopVideos(_ context.Context, n int) ([]Video, error) {
	s.mu.RLock()
	all := make([]Video, 0, len(s.videos))
	for _, v := range s.videos {
		all = append(all, v)
	}
	s.mu.RUnlock()
	// simple selection sort by views desc
	for i := 0; i < len(all)-1; i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].Views > all[i].Views {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	if n < len(all) {
		all = all[:n]
	}
	return all, nil
}

func (s *MemVideoStore) AddToHistory(_ context.Context, userID, videoID string) error {
	s.mu.Lock()
	s.history[userID] = append([]string{videoID}, s.history[userID]...)
	s.mu.Unlock()
	return nil
}

func (s *MemVideoStore) GetHistory(_ context.Context, userID string, n int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h := s.history[userID]
	if n < len(h) {
		h = h[:n]
	}
	return h, nil
}

// ── Streaming logic ───────────────────────────────────────────

type Streamer struct{ store VideoStore }

func NewStreamer(store VideoStore) *Streamer { return &Streamer{store} }

func (s *Streamer) Upload(ctx context.Context, title, owner string, sizeBytes int) (Video, error) {
	chunks := sizeBytes / chunkSize
	if chunks == 0 {
		chunks = 1
	}
	v := Video{
		ID:        fmt.Sprintf("v%d-%d", time.Now().UnixNano(), atomic.AddUint64(&videoSeq, 1)),
		Title:     title,
		Owner:     owner,
		SizeBytes: sizeBytes,
		Chunks:    chunks,
		Resols:    []string{"360p", "720p", "1080p"},
		Views:     0,
		CreatedAt: time.Now().Unix(),
	}
	return v, s.store.SaveVideo(ctx, v)
}

func (s *Streamer) Manifest(ctx context.Context, videoID, resolution string) (Manifest, error) {
	v, err := s.store.GetVideo(ctx, videoID)
	if err != nil {
		return Manifest{}, err
	}
	urls := make([]string, v.Chunks)
	for i := range urls {
		urls[i] = fmt.Sprintf("/chunk/%s/%d?res=%s", v.ID, i, resolution)
	}
	return Manifest{
		VideoID:     v.ID,
		Title:       v.Title,
		TotalChunks: v.Chunks,
		Resolutions: v.Resols,
		ChunkURLs:   urls,
	}, nil
}

func (s *Streamer) GetChunk(ctx context.Context, videoID string, chunkNum int, resolution string) (Chunk, error) {
	v, err := s.store.GetVideo(ctx, videoID)
	if err != nil {
		return Chunk{}, err
	}
	if chunkNum < 0 || chunkNum >= v.Chunks {
		return Chunk{}, fmt.Errorf("chunk %d out of range [0,%d)", chunkNum, v.Chunks)
	}
	return Chunk{
		VideoID:  videoID,
		Num:      chunkNum,
		Total:    v.Chunks,
		Resoln:   resolution,
		Data:     fmt.Sprintf("[chunk-%d-data-%s]", chunkNum, resolution),
		Duration: 4000, // 4 seconds per chunk
	}, nil
}

func (s *Streamer) Watch(ctx context.Context, videoID, userID string) (int64, error) {
	views, err := s.store.IncrViews(ctx, videoID)
	if err != nil {
		return 0, err
	}
	_ = s.store.AddToHistory(ctx, userID, videoID)
	return views, nil
}
