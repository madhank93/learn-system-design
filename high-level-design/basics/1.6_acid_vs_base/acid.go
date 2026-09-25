package main

import (
	"errors"
	"fmt"
	"sync"
)

// ── ACID: in-memory bank demonstrating isolation levels ───────

type Account struct {
	mu      sync.Mutex
	ID      string
	Balance int
}

func NewAccount(id string, balance int) *Account {
	return &Account{ID: id, Balance: balance}
}

// ── Atomicity: transfer either fully commits or fully rolls back ─

var ErrInsufficientFunds = errors.New("insufficient funds")

// Transfer atomically moves amount from src to dst.
// Both accounts lock in ID order to prevent deadlock.
func Transfer(src, dst *Account, amount int) error {
	// lock ordering by ID to avoid deadlock
	first, second := src, dst
	if src.ID > dst.ID {
		first, second = dst, src
	}
	first.mu.Lock()
	defer first.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()

	if src.Balance < amount {
		return ErrInsufficientFunds // atomically: nothing changes
	}
	src.Balance -= amount
	dst.Balance += amount
	return nil
}

// ── Isolation levels ─────────────────────────────────────────

// DirtyReadStore demonstrates read-uncommitted (dirty reads possible).
type DirtyReadStore struct {
	mu    sync.Mutex
	value int
}

func (s *DirtyReadStore) Write(v int) {
	s.mu.Lock()
	s.value = v // written but not "committed" yet
	s.mu.Unlock()
}

func (s *DirtyReadStore) DirtyRead() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value // reads uncommitted data — DIRTY READ
}

// ReadCommittedStore: readers only see committed values.
type ReadCommittedStore struct {
	mu          sync.Mutex
	committed   int
	uncommitted *int // nil when no transaction in progress
}

func (s *ReadCommittedStore) Begin(v int) {
	s.mu.Lock()
	s.uncommitted = &v
	s.mu.Unlock()
}

func (s *ReadCommittedStore) Commit() {
	s.mu.Lock()
	if s.uncommitted != nil {
		s.committed = *s.uncommitted
		s.uncommitted = nil
	}
	s.mu.Unlock()
}

func (s *ReadCommittedStore) Rollback() {
	s.mu.Lock()
	s.uncommitted = nil
	s.mu.Unlock()
}

func (s *ReadCommittedStore) Read() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.committed // only sees committed values
}

// ── BASE: eventual consistency via async replication ─────────

// EventualStore models a primary + replica with replication lag.
type EventualStore struct {
	primary  int
	replicas []int
	mu       sync.RWMutex
}

func NewEventualStore(replicas int) *EventualStore {
	return &EventualStore{replicas: make([]int, replicas)}
}

// Write goes to primary immediately; replicas lag behind.
func (s *EventualStore) Write(v int) {
	s.mu.Lock()
	s.primary = v
	// Replicas NOT updated here — they converge asynchronously.
	s.mu.Unlock()
}

// Replicate simulates the async replication catching up.
func (s *EventualStore) Replicate() {
	s.mu.Lock()
	for i := range s.replicas {
		s.replicas[i] = s.primary // eventually consistent
	}
	s.mu.Unlock()
}

// ReadPrimary always returns latest value.
func (s *EventualStore) ReadPrimary() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.primary
}

// ReadReplica may return stale value before Replicate() runs.
func (s *EventualStore) ReadReplica(i int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.replicas[i]
}

func main() {
	fmt.Println("=== ACID: Atomic Transfer ===")
	alice := NewAccount("alice", 1000)
	bob := NewAccount("bob", 500)

	fmt.Printf("Before: alice=%d bob=%d\n", alice.Balance, bob.Balance)
	if err := Transfer(alice, bob, 200); err != nil {
		fmt.Println("Transfer failed:", err)
	}
	fmt.Printf("After transfer 200: alice=%d bob=%d\n", alice.Balance, bob.Balance)

	if err := Transfer(alice, bob, 900); err != nil {
		fmt.Println("Transfer 900 rejected:", err, "(balances unchanged)")
	}
	fmt.Printf("After failed transfer: alice=%d bob=%d\n", alice.Balance, bob.Balance)

	fmt.Println("\n=== Isolation: Dirty Read vs Read-Committed ===")
	dirty := &DirtyReadStore{value: 100}
	dirty.Write(999) // "in-progress" write
	fmt.Printf("Dirty read (uncommitted): %d  ← risky!\n", dirty.DirtyRead())

	committed := &ReadCommittedStore{committed: 100}
	committed.Begin(999) // start writing 999 but don't commit yet
	fmt.Printf("Read-committed (before commit): %d  ← safe, sees old value\n", committed.Read())
	committed.Commit()
	fmt.Printf("Read-committed (after commit):  %d  ← now sees new value\n", committed.Read())

	fmt.Println("\n=== BASE: Eventual Consistency ===")
	store := NewEventualStore(3)
	store.Write(42)
	fmt.Printf("Primary: %d\n", store.ReadPrimary())
	fmt.Printf("Replicas before sync: %v  ← stale!\n", store.replicas)
	store.Replicate()
	fmt.Printf("Replicas after  sync: %v  ← converged\n", store.replicas)

	fmt.Println("\n=== ACID vs BASE comparison ===")
	fmt.Printf("%-25s %-30s %-30s\n", "Property", "ACID", "BASE")
	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	rows := [][3]string{
		{"Consistency", "Strong (always valid)", "Eventual (converges)"},
		{"Availability", "May sacrifice", "Always available"},
		{"Partition handling", "Refuse or delay", "Accept + reconcile"},
		{"Transactions", "Full ACID txns", "Best-effort / saga"},
		{"Use case", "Banking, inventory", "Social feed, analytics"},
	}
	for _, r := range rows {
		fmt.Printf("%-25s %-30s %-30s\n", r[0], r[1], r[2])
	}
}
