package main

import (
	"testing"
)

func TestEncodeBase62Length(t *testing.T) {
	for _, n := range []uint64{0, 1, 62, 1000, 1 << 32, 1<<62 - 1} {
		code := encodeBase62(n)
		if len(code) != 7 {
			t.Errorf("encodeBase62(%d): expected 7 chars, got %d (%q)", n, len(code), code)
		}
	}
}

func TestEncodeBase62Charset(t *testing.T) {
	code := encodeBase62(123456789)
	for _, c := range code {
		found := false
		for _, b := range base62 {
			if c == b {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("character %q not in base62 charset", c)
		}
	}
}

func TestHashDeterministic(t *testing.T) {
	url := "https://example.com/some/very/long/path"
	h1 := hash(url)
	h2 := hash(url)
	if h1 != h2 {
		t.Errorf("hash not deterministic: %q != %q", h1, h2)
	}
}

func TestHashDifferentURLs(t *testing.T) {
	h1 := hash("https://google.com")
	h2 := hash("https://github.com")
	if h1 == h2 {
		t.Errorf("different URLs should produce different hashes")
	}
}

func TestHashLength(t *testing.T) {
	h := hash("https://example.com")
	if len(h) != 7 {
		t.Errorf("expected 7-char code, got %d: %q", len(h), h)
	}
}
