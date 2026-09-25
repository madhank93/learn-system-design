package main

import (
	"fmt"
	"testing"
)

func TestBTreeInsertAndSearch(t *testing.T) {
	bt := NewBTree()
	for i := 1; i <= 20; i++ {
		bt.Insert(i, fmt.Sprintf("val-%d", i))
	}
	for i := 1; i <= 20; i++ {
		val, ok := bt.Search(i)
		if !ok {
			t.Errorf("key %d: not found", i)
		}
		if val != fmt.Sprintf("val-%d", i) {
			t.Errorf("key %d: expected val-%d, got %q", i, i, val)
		}
	}
}

func TestBTreeSearchMissing(t *testing.T) {
	bt := NewBTree()
	bt.Insert(5, "five")
	_, ok := bt.Search(999)
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestBTreeRangeSearch(t *testing.T) {
	bt := NewBTree()
	for i := 1; i <= 10; i++ {
		bt.Insert(i*10, fmt.Sprintf("v%d", i*10))
	}
	results := bt.RangeSearch(30, 60)
	if len(results) != 4 {
		t.Errorf("expected 4 results (30,40,50,60), got %d: %v", len(results), results)
	}
}

func TestBTreeLargeInsert(t *testing.T) {
	bt := NewBTree()
	n := 1000
	for i := n; i >= 1; i-- {
		bt.Insert(i, fmt.Sprintf("val-%d", i))
	}
	for i := 1; i <= n; i++ {
		_, ok := bt.Search(i)
		if !ok {
			t.Errorf("key %d not found after large insert", i)
		}
	}
}
