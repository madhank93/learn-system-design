package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
	"github.com/madhank93/learn-system-design/syslings/internal/runner"
)

func testModel(t *testing.T) Model {
	t.Helper()
	exs := []exercise.Exercise{
		{ID: "01", Title: "scenario one", Type: exercise.TypeScenario, Check: "01/check.sh", Setup: "01/setup.sh", Hints: []string{"h1"}},
		{ID: "02", Title: "fix two", Type: exercise.TypeFix, File: "a.conf", Check: "02/check.sh", Hints: []string{"h1", "h2"}},
	}
	st := &exercise.State{Completed: map[string]bool{}}
	m := New("load-balancing", exs, st, t.TempDir()+"/.state.json")
	m.ListOpen = false // focused tests drive the exercise flow directly
	// stub out exec-side effects
	m.check = func(ex exercise.Exercise) tea.Cmd { return func() tea.Msg { return CheckMsg{} } }
	m.setup = func(ex exercise.Exercise) tea.Cmd { return func() tea.Msg { return SetupMsg{} } }
	m.resetFile = func(ex exercise.Exercise) error { return nil }
	return m
}

func TestCheckPassAdvancesAndPersists(t *testing.T) {
	m := testModel(t)
	m.Cur = 0
	m.Phase = PhaseFailed

	next, _ := m.Update(CheckMsg{Res: runner.Result{Passed: true, Output: "✓"}})
	m2 := next.(Model)
	if !m2.St.Completed["01"] {
		t.Fatal("exercise 01 not marked complete")
	}
	if m2.Cur != 1 {
		t.Fatalf("cur = %d, want 1", m2.Cur)
	}
}

func TestCheckFailShowsOutput(t *testing.T) {
	m := testModel(t)
	next, _ := m.Update(CheckMsg{Res: runner.Result{Passed: false, Output: "✗ broke\nhint: x"}})
	m2 := next.(Model)
	if m2.Phase != PhaseFailed {
		t.Fatalf("phase = %v, want failed", m2.Phase)
	}
	if !strings.Contains(m2.Output, "✗ broke") {
		t.Fatalf("output not captured: %q", m2.Output)
	}
}

func TestAllDoneAfterLastPass(t *testing.T) {
	m := testModel(t)
	m.St.Completed["01"] = true
	m.Cur = 1
	next, _ := m.Update(CheckMsg{Res: runner.Result{Passed: true}})
	m2 := next.(Model)
	if m2.Phase != PhaseAllDone {
		t.Fatalf("phase = %v, want all-done", m2.Phase)
	}
}

func TestWatchEventTriggersRecheck(t *testing.T) {
	m := testModel(t)
	m.Phase = PhaseFailed
	next, cmd := m.Update(WatchMsg{})
	m2 := next.(Model)
	if m2.Phase != PhaseChecking {
		t.Fatalf("phase = %v, want checking", m2.Phase)
	}
	if cmd == nil {
		t.Fatal("want check command scheduled")
	}
}

func TestHintKeyRevealsProgressively(t *testing.T) {
	m := testModel(t)
	m.Cur = 1
	m.Phase = PhaseFailed
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m2 := next.(Model)
	if m2.HintsShown != 1 {
		t.Fatalf("hints shown = %d, want 1", m2.HintsShown)
	}
	next, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	next, _ = next.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if got := next.(Model).HintsShown; got != 2 {
		t.Fatalf("hints shown = %d, want capped at 2", got)
	}
}

func TestSkipAdvancesWithoutCompleting(t *testing.T) {
	m := testModel(t)
	m.Cur = 0
	m.Phase = PhaseFailed
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m2 := next.(Model)
	if m2.Cur != 1 {
		t.Fatalf("cur = %d, want 1", m2.Cur)
	}
	if m2.St.Completed["01"] {
		t.Fatal("skip must not mark complete")
	}
}

func TestQuitKey(t *testing.T) {
	m := testModel(t)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if !next.(Model).Quitting {
		t.Fatal("want quitting")
	}
	if cmd == nil {
		t.Fatal("want tea.Quit scheduled")
	}
}

func TestScenarioEntryRunsSetupThenCheck(t *testing.T) {
	m := testModel(t)
	// entering exercise 0 (scenario) should schedule setup
	next, cmd := m.Update(EnterMsg{})
	m2 := next.(Model)
	if m2.Phase != PhaseSettingUp {
		t.Fatalf("phase = %v, want setting-up", m2.Phase)
	}
	if cmd == nil {
		t.Fatal("want setup command")
	}
	// setup completion schedules the check
	next, cmd = m2.Update(SetupMsg{})
	if next.(Model).Phase != PhaseChecking {
		t.Fatalf("phase after setup = %v, want checking", next.(Model).Phase)
	}
	if cmd == nil {
		t.Fatal("want check command after setup")
	}
}

func TestViewRendersCoreElements(t *testing.T) {
	m := testModel(t)
	m.Cur = 1
	m.Phase = PhaseFailed
	m.Output = "  ✗ quick requests spread evenly"
	v := m.View()
	for _, want := range []string{"syslings", "fix two", "a.conf", "✗ quick requests", "[h]int"} {
		if !strings.Contains(v, want) {
			t.Fatalf("view missing %q:\n%s", want, v)
		}
	}
}
