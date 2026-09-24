package main

import (
	"fmt"
	"testing"
)

// TestRingAddNode verifies that adding one node populates exactly vnodes
// positions in the sorted slice, and that get() returns that node for any key.
func TestRingAddNode(t *testing.T) {
	r := newRing()
	r.add("node1:8080")

	r.mu.RLock()
	got := len(r.sorted)
	r.mu.RUnlock()

	if got != vnodes {
		t.Fatalf("after adding 1 node: len(sorted) = %d, want %d", got, vnodes)
	}
	if node := r.get("any-key"); node != "node1:8080" {
		t.Errorf("get(any-key) = %q, want %q", node, "node1:8080")
	}
}

// TestRingGetDeterministic verifies that the same key resolves to the same node
// across 100 consecutive calls.
func TestRingGetDeterministic(t *testing.T) {
	r := newRing()
	r.add("node1:8080")
	r.add("node2:8080")

	key := "stable-key"
	first := r.get(key)
	for i := 0; i < 100; i++ {
		if got := r.get(key); got != first {
			t.Fatalf("call %d: get(%q) = %q, want %q", i, key, got, first)
		}
	}
}

// TestRingAddMultipleNodes verifies that after adding 3 nodes, routing 300
// distinct keys distributes work across all three nodes (no node gets 0 keys).
func TestRingAddMultipleNodes(t *testing.T) {
	r := newRing()
	nodes := []string{"node1:8080", "node2:8080", "node3:8080"}
	for _, n := range nodes {
		r.add(n)
	}

	counts := make(map[string]int)
	for i := 0; i < 300; i++ {
		counts[r.get(fmt.Sprintf("key-%d", i))]++
	}
	for _, n := range nodes {
		if counts[n] == 0 {
			t.Errorf("node %q received no keys out of 300", n)
		}
	}
}

// TestRingRemoveNode verifies that after removing node2, none of 100 probed
// keys resolve to that node.
func TestRingRemoveNode(t *testing.T) {
	r := newRing()
	r.add("node1:8080")
	r.add("node2:8080")
	r.add("node3:8080")
	r.remove("node2:8080")

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		if got := r.get(key); got == "node2:8080" {
			t.Errorf("get(%q) = %q after node2 was removed", key, got)
		}
	}
}

// TestRingVnodeCount verifies that each add() call contributes exactly vnodes
// (150) new positions to the sorted slice.
func TestRingVnodeCount(t *testing.T) {
	r := newRing()
	nodes := []string{"node1:8080", "node2:8080", "node3:8080"}
	for i, n := range nodes {
		r.add(n)
		r.mu.RLock()
		got := len(r.sorted)
		r.mu.RUnlock()
		want := (i + 1) * vnodes
		if got != want {
			t.Errorf("after adding %d node(s): len(sorted) = %d, want %d", i+1, got, want)
		}
	}
}

// TestRingMigrationFraction verifies that adding a fourth node displaces fewer
// than 50% of the previously routed keys.  With uniform distribution across 4
// nodes the expected disruption is ~25%.
func TestRingMigrationFraction(t *testing.T) {
	r := newRing()
	r.add("node1:8080")
	r.add("node2:8080")
	r.add("node3:8080")

	const total = 100
	before := make(map[string]string, total)
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("key-%d", i)
		before[key] = r.get(key)
	}

	r.add("node4:8080")

	changed := 0
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("key-%d", i)
		if r.get(key) != before[key] {
			changed++
		}
	}

	fraction := float64(changed) / float64(total)
	if fraction >= 0.50 {
		t.Errorf("migration fraction = %.2f (%.0f%%), want < 50%%", fraction, fraction*100)
	}
}

// TestRingEmptyReturnsEmpty verifies that get() on a ring with no nodes returns
// the empty string.
func TestRingEmptyReturnsEmpty(t *testing.T) {
	r := newRing()
	if got := r.get("any-key"); got != "" {
		t.Errorf("get on empty ring = %q, want %q", got, "")
	}
}
