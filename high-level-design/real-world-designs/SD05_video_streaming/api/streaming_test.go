package main

import (
	"context"
	"testing"
)

func newTestStreamer() *Streamer {
	return NewStreamer(NewMemVideoStore())
}

func TestUpload(t *testing.T) {
	ctx := context.Background()
	s := newTestStreamer()

	v, err := s.Upload(ctx, "Big Buck Bunny", "alice", 50*1024*1024)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if v.ID == "" {
		t.Fatal("empty video ID")
	}
	if v.Title != "Big Buck Bunny" {
		t.Fatalf("want title %q, got %q", "Big Buck Bunny", v.Title)
	}
	expectedChunks := 50 * 1024 * 1024 / chunkSize
	if v.Chunks != expectedChunks {
		t.Fatalf("want %d chunks, got %d", expectedChunks, v.Chunks)
	}
	if len(v.Resols) != 3 {
		t.Fatalf("want 3 resolutions, got %d", len(v.Resols))
	}
}

func TestManifest(t *testing.T) {
	ctx := context.Background()
	s := newTestStreamer()

	v, _ := s.Upload(ctx, "Test Video", "bob", 5*1024*1024)
	m, err := s.Manifest(ctx, v.ID, "720p")
	if err != nil {
		t.Fatalf("Manifest: %v", err)
	}
	if m.VideoID != v.ID {
		t.Fatalf("manifest video ID mismatch")
	}
	if len(m.ChunkURLs) != v.Chunks {
		t.Fatalf("want %d chunk URLs, got %d", v.Chunks, len(m.ChunkURLs))
	}
	for i, url := range m.ChunkURLs {
		if url == "" {
			t.Fatalf("empty chunk URL at index %d", i)
		}
	}
}

func TestGetChunk(t *testing.T) {
	ctx := context.Background()
	s := newTestStreamer()

	v, _ := s.Upload(ctx, "Chunk Test", "carol", 2*1024*1024)
	chunk, err := s.GetChunk(ctx, v.ID, 0, "360p")
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if chunk.VideoID != v.ID {
		t.Fatalf("chunk video ID mismatch")
	}
	if chunk.Num != 0 {
		t.Fatalf("want chunk 0, got %d", chunk.Num)
	}
	if chunk.Resoln != "360p" {
		t.Fatalf("want 360p, got %q", chunk.Resoln)
	}
}

func TestGetChunkOutOfRange(t *testing.T) {
	ctx := context.Background()
	s := newTestStreamer()

	v, _ := s.Upload(ctx, "OOB Test", "dave", 1024*1024)
	_, err := s.GetChunk(ctx, v.ID, 9999, "720p")
	if err == nil {
		t.Fatal("expected error for out-of-range chunk")
	}
}

func TestWatch(t *testing.T) {
	ctx := context.Background()
	s := newTestStreamer()

	v, _ := s.Upload(ctx, "Popular Video", "eve", 10*1024*1024)

	views1, err := s.Watch(ctx, v.ID, "user1")
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	if views1 != 1 {
		t.Fatalf("want 1 view, got %d", views1)
	}

	views2, _ := s.Watch(ctx, v.ID, "user2")
	if views2 != 2 {
		t.Fatalf("want 2 views, got %d", views2)
	}
}

func TestWatchHistory(t *testing.T) {
	ctx := context.Background()
	s := newTestStreamer()

	v1, _ := s.Upload(ctx, "Vid1", "owner", 1024*1024)
	v2, _ := s.Upload(ctx, "Vid2", "owner", 1024*1024)

	_, _ = s.Watch(ctx, v1.ID, "alice")
	_, _ = s.Watch(ctx, v2.ID, "alice")

	hist, err := s.store.GetHistory(ctx, "alice", 10)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(hist) != 2 {
		t.Fatalf("want 2 history entries, got %d", len(hist))
	}
}

func TestTopVideos(t *testing.T) {
	ctx := context.Background()
	s := newTestStreamer()

	v1, _ := s.Upload(ctx, "Low Views", "a", 1024*1024)
	v2, _ := s.Upload(ctx, "High Views", "b", 1024*1024)

	_, _ = s.Watch(ctx, v2.ID, "u1")
	_, _ = s.Watch(ctx, v2.ID, "u2")
	_, _ = s.Watch(ctx, v2.ID, "u3")
	_, _ = s.Watch(ctx, v1.ID, "u1")

	top, err := s.store.TopVideos(ctx, 2)
	if err != nil {
		t.Fatalf("TopVideos: %v", err)
	}
	if len(top) != 2 {
		t.Fatalf("want 2, got %d", len(top))
	}
	if top[0].ID != v2.ID {
		t.Fatalf("want most-viewed first, got %q", top[0].Title)
	}
}

func TestManifestUnknownVideo(t *testing.T) {
	ctx := context.Background()
	s := newTestStreamer()

	_, err := s.Manifest(ctx, "nonexistent", "720p")
	if err == nil {
		t.Fatal("expected error for unknown video")
	}
}
