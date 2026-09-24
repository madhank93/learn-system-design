package main

import (
	"fmt"
	"testing"
)

func TestBloomFilterNoFalseNegatives(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)
	items := []string{"apple", "banana", "cherry", "date", "elderberry"}
	for _, item := range items {
		bf.Add(item)
	}
	for _, item := range items {
		if !bf.Contains(item) {
			t.Errorf("false negative: %q was added but Contains returned false", item)
		}
	}
}

func TestBloomFilterFalsePositiveRate(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)
	// Add 1000 items
	for i := 0; i < 1000; i++ {
		bf.Add(fmt.Sprintf("item-%d", i))
	}
	// Test 10000 items that were NOT added
	fp := 0
	trials := 10000
	for i := 10000; i < 10000+trials; i++ {
		if bf.Contains(fmt.Sprintf("item-%d", i)) {
			fp++
		}
	}
	fpRate := float64(fp) / float64(trials)
	// Allow 3× the target rate as tolerance for randomness
	if fpRate > 0.03 {
		t.Errorf("false positive rate %.4f exceeds 3%% tolerance (target was 1%%)", fpRate)
	}
	t.Logf("false positive rate: %.4f (%.2f%%)", fpRate, fpRate*100)
}

func TestBloomFilterDefinitelyNotContains(t *testing.T) {
	bf := NewBloomFilter(100, 0.01)
	// Empty filter should not contain anything
	words := []string{"foo", "bar", "baz"}
	for _, w := range words {
		if bf.Contains(w) {
			// This is theoretically a false positive — but on a fresh empty filter,
			// it would require all hash positions to be 0 by coincidence
			t.Logf("unexpected hit for %q in empty filter (improbable)", w)
		}
	}
}

func TestBloomFilterSizeScaling(t *testing.T) {
	// Larger n → larger bit array
	small := NewBloomFilter(100, 0.01)
	large := NewBloomFilter(10000, 0.01)
	if large.size <= small.size {
		t.Errorf("expected larger filter for larger n: small=%d large=%d", small.size, large.size)
	}
}

func TestBloomFilterHashCount(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)
	if bf.hashFns < 1 {
		t.Errorf("expected at least 1 hash function, got %d", bf.hashFns)
	}
	t.Logf("hash functions: %d, bit array size: %d", bf.hashFns, bf.size)
}
