package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherFiresOnWrite(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.conf")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	w, err := New(dir, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	if err := os.WriteFile(file, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-w.Events():
	case <-time.After(3 * time.Second):
		t.Fatal("no event after write")
	}
}

func TestWatcherDebouncesBurst(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.conf")

	w, err := New(dir, 150*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	// burst of writes inside one debounce window
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(file, []byte{byte(i)}, 0o644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// first (debounced) event
	select {
	case <-w.Events():
	case <-time.After(3 * time.Second):
		t.Fatal("no event after burst")
	}
	// no second event should arrive shortly after
	select {
	case <-w.Events():
		t.Fatal("burst produced more than one debounced event")
	case <-time.After(400 * time.Millisecond):
	}
}
