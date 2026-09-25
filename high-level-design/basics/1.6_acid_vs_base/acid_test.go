package main

import (
	"sync"
	"testing"
)

func TestTransferAtomic(t *testing.T) {
	alice := NewAccount("alice", 1000)
	bob := NewAccount("bob", 500)

	if err := Transfer(alice, bob, 200); err != nil {
		t.Fatalf("transfer 200: %v", err)
	}
	if alice.Balance != 800 {
		t.Errorf("alice: want 800, got %d", alice.Balance)
	}
	if bob.Balance != 700 {
		t.Errorf("bob: want 700, got %d", bob.Balance)
	}
}

func TestTransferInsufficientFunds(t *testing.T) {
	alice := NewAccount("alice", 100)
	bob := NewAccount("bob", 0)

	err := Transfer(alice, bob, 200)
	if err != ErrInsufficientFunds {
		t.Fatalf("want ErrInsufficientFunds, got %v", err)
	}
	// Balances unchanged — atomicity
	if alice.Balance != 100 {
		t.Errorf("alice balance changed on failed transfer: %d", alice.Balance)
	}
	if bob.Balance != 0 {
		t.Errorf("bob balance changed on failed transfer: %d", bob.Balance)
	}
}

func TestTransferConservesTotal(t *testing.T) {
	alice := NewAccount("alice", 1000)
	bob := NewAccount("bob", 1000)
	total := alice.Balance + bob.Balance

	for i := 0; i < 10; i++ {
		_ = Transfer(alice, bob, 100)
		_ = Transfer(bob, alice, 50)
	}
	if alice.Balance+bob.Balance != total {
		t.Errorf("money not conserved: %d+%d != %d", alice.Balance, bob.Balance, total)
	}
}

func TestTransferConcurrentSafety(t *testing.T) {
	alice := NewAccount("alice", 10000)
	bob := NewAccount("bob", 10000)
	total := alice.Balance + bob.Balance

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = Transfer(alice, bob, 10)
		}()
		go func() {
			defer wg.Done()
			_ = Transfer(bob, alice, 10)
		}()
	}
	wg.Wait()

	if alice.Balance+bob.Balance != total {
		t.Errorf("concurrent transfers: money not conserved: %d+%d != %d",
			alice.Balance, bob.Balance, total)
	}
}

func TestReadCommittedIsolation(t *testing.T) {
	s := &ReadCommittedStore{committed: 42}

	s.Begin(99)
	// Reader must not see uncommitted value
	if got := s.Read(); got != 42 {
		t.Errorf("read-committed: before commit want 42, got %d", got)
	}
	s.Commit()
	if got := s.Read(); got != 99 {
		t.Errorf("read-committed: after commit want 99, got %d", got)
	}
}

func TestReadCommittedRollback(t *testing.T) {
	s := &ReadCommittedStore{committed: 42}
	s.Begin(99)
	s.Rollback()
	if got := s.Read(); got != 42 {
		t.Errorf("after rollback want 42, got %d", got)
	}
}

func TestEventualConsistencyLag(t *testing.T) {
	store := NewEventualStore(3)
	store.Write(100)

	// Primary has new value
	if got := store.ReadPrimary(); got != 100 {
		t.Fatalf("primary: want 100, got %d", got)
	}
	// Replicas are stale before sync
	for i := range store.replicas {
		if store.ReadReplica(i) != 0 {
			t.Errorf("replica %d should be stale (0), got %d", i, store.ReadReplica(i))
		}
	}
}

func TestEventualConsistencyConverges(t *testing.T) {
	store := NewEventualStore(3)
	store.Write(42)
	store.Replicate()

	for i := range store.replicas {
		if got := store.ReadReplica(i); got != 42 {
			t.Errorf("replica %d after sync: want 42, got %d", i, got)
		}
	}
}

func TestDirtyReadSeesUncommitted(t *testing.T) {
	s := &DirtyReadStore{value: 0}
	s.Write(777) // "uncommitted" write
	// Dirty read sees it immediately
	if got := s.DirtyRead(); got != 777 {
		t.Errorf("dirty read: want 777 (uncommitted), got %d", got)
	}
}
