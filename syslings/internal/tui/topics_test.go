package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
)

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func topicEntries(t *testing.T) []TopicEntry {
	t.Helper()
	m := testModel(t)
	return []TopicEntry{
		{Name: "load-balancing", Desc: "Nginx algorithms, health checks, failover", Exs: m.Exs, St: m.St, StatePath: m.StatePath},
		{Name: "caching", Desc: "Hit ratios, TTLs, stampedes", Exs: m.Exs, St: &exercise.State{Completed: map[string]bool{}, Attempted: map[string]bool{}}, StatePath: m.StatePath},
	}
}

func stubBoot(m Model) Model {
	m.bootTopic = func(name string) tea.Cmd { return func() tea.Msg { return BootDoneMsg{} } }
	return m
}

func TestNoPreselectOpensTopicPicker(t *testing.T) {
	m := stubBoot(NewPicker(topicEntries(t), ""))
	if !m.TopicOpen {
		t.Fatal("want topic picker open when no topic preselected")
	}
	if m.Init() != nil {
		t.Fatal("no boot until a topic is chosen")
	}
}

func TestPreselectSkipsTopicPickerAndBoots(t *testing.T) {
	m := stubBoot(NewPicker(topicEntries(t), "load-balancing"))
	if m.TopicOpen {
		t.Fatal("preselected topic must skip the topic picker")
	}
	if !m.Booting {
		t.Fatal("want booting state")
	}
	if m.Init() == nil {
		t.Fatal("want boot command on init")
	}
}

func TestTopicSelectionBootsThenShowsHome(t *testing.T) {
	m := stubBoot(NewPicker(topicEntries(t), ""))
	next, _ := m.Update(key("j"))
	m2 := next.(Model)
	if m2.TopicCursor != 1 {
		t.Fatalf("topic cursor = %d, want 1", m2.TopicCursor)
	}
	next, cmd := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := next.(Model)
	if m3.TopicOpen || !m3.Booting {
		t.Fatal("enter must close topic picker and start booting")
	}
	if m3.Topic != "caching" {
		t.Fatalf("topic = %q, want caching", m3.Topic)
	}
	if cmd == nil {
		t.Fatal("want boot command")
	}
	next, _ = m3.Update(BootDoneMsg{})
	m4 := next.(Model)
	if m4.Booting || !m4.StackUp {
		t.Fatal("boot completion must land on the exercise home with the stack up")
	}
	if !strings.Contains(m4.View(), "stack ● up") {
		t.Fatal("home view must show stack status")
	}
}

func TestBootFailureSurfaces(t *testing.T) {
	m := stubBoot(NewPicker(topicEntries(t), "load-balancing"))
	next, _ := m.Update(BootDoneMsg{Err: errFake})
	m2 := next.(Model)
	if m2.BootErr == nil {
		t.Fatal("boot error must be kept for the view")
	}
	if !strings.Contains(m2.View(), "fake docker explosion") {
		t.Fatalf("view must show boot error:\n%s", m2.View())
	}
}

func TestTopicPickerViewShowsDescriptions(t *testing.T) {
	m := stubBoot(NewPicker(topicEntries(t), ""))
	v := m.View()
	for _, want := range []string{"load-balancing", "Nginx algorithms", "caching", "Hit ratios"} {
		if !strings.Contains(v, want) {
			t.Fatalf("topic picker missing %q:\n%s", want, v)
		}
	}
}

func TestStackDownKey(t *testing.T) {
	m := testModel(t)
	m.stackDown = func() tea.Cmd { return func() tea.Msg { return DownDoneMsg{} } }
	m.Playing = 1
	next, cmd := m.Update(key("d"))
	if cmd == nil {
		t.Fatal("d with stack up must schedule stack down")
	}
	next, _ = next.(Model).Update(DownDoneMsg{})
	m2 := next.(Model)
	if m2.StackUp || m2.Playing != -1 {
		t.Fatal("down completion must mark stack down and stop playing")
	}
}

type fakeErr string

func (e fakeErr) Error() string { return string(e) }

var errFake = fakeErr("fake docker explosion")
