// Package watch emits debounced change events for a directory tree, so a
// flurry of editor writes triggers exactly one re-check.
package watch

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fs     *fsnotify.Watcher
	events chan struct{}
	done   chan struct{}
}

// New watches dir (non-recursive — work dirs are flat) and coalesces raw
// filesystem events into single notifications separated by at least debounce.
func New(dir string, debounce time.Duration) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := fsw.Add(dir); err != nil {
		fsw.Close()
		return nil, err
	}
	w := &Watcher{fs: fsw, events: make(chan struct{}, 1), done: make(chan struct{})}
	go w.loop(debounce)
	return w, nil
}

func (w *Watcher) loop(debounce time.Duration) {
	var timer *time.Timer
	var timerC <-chan time.Time
	for {
		select {
		case ev, ok := <-w.fs.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			if timer == nil {
				timer = time.NewTimer(debounce)
				timerC = timer.C
			} else {
				timer.Reset(debounce)
			}
		case <-timerC:
			timer = nil
			timerC = nil
			select {
			case w.events <- struct{}{}:
			default: // a notification is already pending; drop
			}
		case <-w.done:
			return
		}
	}
}

// Events delivers one signal per debounced burst of changes.
func (w *Watcher) Events() <-chan struct{} { return w.events }

func (w *Watcher) Close() error {
	close(w.done)
	return w.fs.Close()
}
