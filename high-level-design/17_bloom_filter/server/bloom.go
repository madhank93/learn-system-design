package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"math"
)

// BloomFilter is a probabilistic set membership data structure.
// False positives are possible; false negatives are not.
type BloomFilter struct {
	bits    []bool
	size    uint
	hashFns int
}

// NewBloomFilter creates a bloom filter sized for n expected elements
// and target false-positive rate p.
func NewBloomFilter(n int, p float64) *BloomFilter {
	// Optimal bit array size: m = -n*ln(p) / (ln2)^2
	m := uint(math.Ceil(-float64(n) * math.Log(p) / (math.Log(2) * math.Log(2))))
	// Optimal hash count: k = (m/n)*ln(2)
	k := int(math.Ceil(float64(m) / float64(n) * math.Log(2)))
	return &BloomFilter{bits: make([]bool, m), size: m, hashFns: k}
}

// hash produces k positions on the bit array for a given item.
func (bf *BloomFilter) positions(item string) []uint {
	data := []byte(item)

	h1b := md5.Sum(data)
	h1 := binary.BigEndian.Uint64(h1b[:8])

	h2h := sha1.New()
	h2h.Write(data)
	h2b := h2h.Sum(nil)
	h2 := binary.BigEndian.Uint64(h2b[:8])

	h3b := sha256.Sum256(data)
	h3 := binary.BigEndian.Uint64(h3b[:8])

	positions := make([]uint, bf.hashFns)
	for i := 0; i < bf.hashFns; i++ {
		// Double hashing: h(i, x) = h1(x) + i*h2(x) + i²*h3(x)
		combined := h1 + uint64(i)*h2 + uint64(i*i)*h3
		positions[i] = uint(combined) % bf.size
	}
	return positions
}

// Add inserts an item into the bloom filter.
func (bf *BloomFilter) Add(item string) {
	for _, pos := range bf.positions(item) {
		bf.bits[pos] = true
	}
}

// Contains reports whether item might be in the set.
// Returns true = "probably yes" (may be a false positive)
// Returns false = "definitely no"
func (bf *BloomFilter) Contains(item string) bool {
	for _, pos := range bf.positions(item) {
		if !bf.bits[pos] {
			return false
		}
	}
	return true
}

// FalsePositiveRate estimates the current FP rate based on bits set.
func (bf *BloomFilter) FalsePositiveRate() float64 {
	set := 0
	for _, b := range bf.bits {
		if b {
			set++
		}
	}
	ratio := float64(set) / float64(bf.size)
	return math.Pow(ratio, float64(bf.hashFns))
}
