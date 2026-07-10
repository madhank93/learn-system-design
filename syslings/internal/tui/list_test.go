package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestTUIOpensWithPicker(t *testing.T) {
	m := testModel(t) // testModel closes the picker for the focused key tests
	fresh := New("load-balancing", m.Exs, m.St, m.StatePath)
	if !fresh.ListOpen {
		t.Fatal("TUI must open with the exercise picker visible")
	}
	if fresh.ListCursor != 0 {
		t.Fatalf("picker cursor = %d, want 0 (first incomplete)", fresh.ListCursor)
	}
}

func TestPickerCursorStartsAtNextIncomplete(t *testing.T) {
	m := testModel(t)
	m.St.Completed["01"] = true
	fresh := New("load-balancing", m.Exs, m.St, m.StatePath)
	if fresh.ListCursor != 1 {
		t.Fatalf("picker cursor = %d, want 1", fresh.ListCursor)
	}
}

func TestPickerNavigationAndJump(t *testing.T) {
	m := testModel(t)
	m.ListOpen = true
	m.ListCursor = 0

	next, _ := m.Update(key("j"))
	m2 := next.(Model)
	if m2.ListCursor != 1 {
		t.Fatalf("cursor after j = %d, want 1", m2.ListCursor)
	}
	// bounded at the end
	next, _ = m2.Update(key("j"))
	if got := next.(Model).ListCursor; got != 1 {
		t.Fatalf("cursor must clamp at last exercise, got %d", got)
	}

	next, cmd := next.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := next.(Model)
	if m3.ListOpen {
		t.Fatal("enter must close the picker")
	}
	if m3.Cur != 1 {
		t.Fatalf("enter must jump to selected exercise, cur = %d", m3.Cur)
	}
	if cmd == nil {
		t.Fatal("enter must schedule EnterMsg to start the exercise")
	}
	if _, ok := cmd().(EnterMsg); !ok {
		t.Fatal("scheduled cmd must produce EnterMsg")
	}
}

func TestListKeyReopensPicker(t *testing.T) {
	m := testModel(t)
	m.ListOpen = false
	next, _ := m.Update(key("l"))
	if !next.(Model).ListOpen {
		t.Fatal("l must reopen the picker")
	}
}

func TestWatchIgnoredWhilePickerOpen(t *testing.T) {
	m := testModel(t)
	m.ListOpen = true
	m.Phase = PhaseFailed
	next, cmd := m.Update(WatchMsg{})
	if cmd != nil || next.(Model).Phase != PhaseFailed {
		t.Fatal("watch events must be ignored while the picker is open")
	}
}

func TestPickerViewShowsStatusAndTitles(t *testing.T) {
	m := testModel(t)
	m.ListOpen = true
	m.St.Completed["01"] = true
	m.ListCursor = 1
	v := m.View()
	for _, want := range []string{"✅", "scenario one", "fix two", "▶"} {
		if !strings.Contains(v, want) {
			t.Fatalf("picker view missing %q:\n%s", want, v)
		}
	}
}
