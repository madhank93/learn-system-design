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
		{ID: "02", Title: "fix two", Desc: "Something is broken.", Type: exercise.TypeFix, File: "a.conf", Check: "02/check.sh", Hints: []string{"h1", "h2"}},
		{ID: "03", Title: "fix three", Type: exercise.TypeFix, File: "b.conf", Check: "03/check.sh"},
	}
	st := &exercise.State{Completed: map[string]bool{}, Attempted: map[string]bool{}}
	m := New("load-balancing", exs, st, t.TempDir()+"/.state.json")
	// stub out exec-side effects
	m.check = func(ex exercise.Exercise) tea.Cmd { return func() tea.Msg { return CheckMsg{} } }
	m.setup = func(ex exercise.Exercise) tea.Cmd { return func() tea.Msg { return SetupMsg{} } }
	m.solution = func(ex exercise.Exercise) tea.Cmd { return func() tea.Msg { return SolutionMsg{} } }
	m.resetFile = func(ex exercise.Exercise) error { return nil }
	return m
}

func TestPlayArmsExerciseAndChecks(t *testing.T) {
	m := testModel(t)
	m.Cursor = 1
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := next.(Model)
	if m2.Playing != 1 {
		t.Fatalf("playing = %d, want 1", m2.Playing)
	}
	if cmd == nil {
		t.Fatal("want EnterMsg scheduled")
	}
	// EnterMsg on a fix exercise without setup goes straight to checking
	next, cmd = m2.Update(EnterMsg{})
	if next.(Model).Phase != PhaseChecking || cmd == nil {
		t.Fatal("want checking phase + check cmd")
	}
}

func TestCheckPassMarksDoneAndHighlightsNext(t *testing.T) {
	m := testModel(t)
	m.Playing = 0
	m.Phase = PhaseChecking
	next, _ := m.Update(CheckMsg{Res: runner.Result{Passed: true, Output: "✓ nice"}})
	m2 := next.(Model)
	if !m2.St.Completed["01"] {
		t.Fatal("exercise 01 not marked complete")
	}
	if m2.Playing != -1 {
		t.Fatal("pass must stop watching (playing = -1)")
	}
	if m2.Cursor != 1 {
		t.Fatalf("cursor = %d, want next unfinished (1)", m2.Cursor)
	}
	if m2.Phase != PhasePassed {
		t.Fatalf("phase = %v, want passed", m2.Phase)
	}
}

func TestCheckFailMarksAttempted(t *testing.T) {
	m := testModel(t)
	m.Playing = 1
	next, _ := m.Update(CheckMsg{Res: runner.Result{Passed: false, Output: "✗ broke\nhint: x"}})
	m2 := next.(Model)
	if m2.Phase != PhaseFailed {
		t.Fatalf("phase = %v, want failed", m2.Phase)
	}
	if !m2.St.Attempted["02"] {
		t.Fatal("fail must mark attempted (◐)")
	}
	if !strings.Contains(m2.Output, "✗ broke") {
		t.Fatalf("output not captured: %q", m2.Output)
	}
}

func TestWatchRechecksOnlyWhilePlaying(t *testing.T) {
	m := testModel(t)
	m.Playing = -1
	if _, cmd := m.Update(WatchMsg{}); cmd != nil {
		t.Fatal("watch must be ignored when nothing is playing")
	}
	m.Playing = 1
	m.Phase = PhaseFailed
	next, cmd := m.Update(WatchMsg{})
	if next.(Model).Phase != PhaseChecking || cmd == nil {
		t.Fatal("watch while playing must re-check")
	}
}

func TestHintRevealsProgressivelyPerExercise(t *testing.T) {
	m := testModel(t)
	m.Cursor = 1
	next, _ := m.Update(key("h"))
	next, _ = next.(Model).Update(key("h"))
	next, _ = next.(Model).Update(key("h"))
	m2 := next.(Model)
	if m2.Hints["02"] != 2 {
		t.Fatalf("hints for 02 = %d, want capped at 2", m2.Hints["02"])
	}
	if m2.Hints["01"] != 0 {
		t.Fatal("hints must be tracked per exercise")
	}
}

func TestNextKeyJumpsToNextUnfinished(t *testing.T) {
	m := testModel(t)
	m.St.Completed["01"] = true
	m.Cursor = 2
	next, _ := m.Update(key("n"))
	if got := next.(Model).Cursor; got != 1 {
		t.Fatalf("cursor = %d, want 1", got)
	}
}

func TestScenarioSetupRunsBeforeCheck(t *testing.T) {
	m := testModel(t)
	m.Playing = 0
	next, cmd := m.Update(EnterMsg{})
	m2 := next.(Model)
	if m2.Phase != PhaseSettingUp || cmd == nil {
		t.Fatalf("phase = %v, want setting-up with setup cmd", m2.Phase)
	}
	next, cmd = m2.Update(SetupMsg{})
	if next.(Model).Phase != PhaseChecking || cmd == nil {
		t.Fatal("setup completion must schedule the check")
	}
}

func TestSolutionAsksThenApplies(t *testing.T) {
	m := testModel(t)
	m.Cursor = 1
	next, _ := m.Update(key("s"))
	m2 := next.(Model)
	if m2.ConfirmSolItem != 1 {
		t.Fatalf("confirm item = %d, want 1", m2.ConfirmSolItem)
	}
	// "n" declines
	next, _ = m2.Update(key("n"))
	if next.(Model).ConfirmSolItem != -1 {
		t.Fatal("decline must clear the confirm")
	}
	// "y" applies
	m2.ConfirmSolItem = 1
	next, cmd := m2.Update(key("y"))
	m3 := next.(Model)
	if cmd == nil || m3.Playing != 1 {
		t.Fatal("confirm must apply solution and arm the exercise")
	}
	if _, ok := cmd().(SolutionMsg); !ok {
		t.Fatal("want SolutionMsg from solution cmd")
	}
}

func TestPlayWithStackDownBootsFirst(t *testing.T) {
	m := stubBoot(testModel(t))
	m.StackUp = false
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := next.(Model)
	if !m2.Booting || cmd == nil {
		t.Fatal("play with stack down must boot first")
	}
	next, cmd = m2.Update(BootDoneMsg{})
	m3 := next.(Model)
	if !m3.StackUp || cmd == nil {
		t.Fatal("boot completion must resume the queued play")
	}
	if _, ok := cmd().(EnterMsg); !ok {
		t.Fatal("queued play must fire EnterMsg after boot")
	}
}

func TestQuitKey(t *testing.T) {
	m := testModel(t)
	next, cmd := m.Update(key("q"))
	if !next.(Model).Quitting || cmd == nil {
		t.Fatal("want quitting + tea.Quit")
	}
}

func TestHelpTogglesAndAnyKeyCloses(t *testing.T) {
	m := testModel(t)
	next, _ := m.Update(key("?"))
	m2 := next.(Model)
	if !m2.ShowHelp {
		t.Fatal("? must open help")
	}
	if !strings.Contains(m2.View(), "markers") {
		t.Fatal("help view must document the markers")
	}
	next, _ = m2.Update(key("x"))
	if next.(Model).ShowHelp {
		t.Fatal("any key must close help")
	}
}

func TestViewRendersListMarkersAndDetail(t *testing.T) {
	m := testModel(t)
	m.St.Completed["01"] = true
	m.St.Attempted["03"] = true
	m.Cursor = 1
	m.Playing = 1
	m.Phase = PhaseFailed
	m.Output = "  ✗ quick requests spread evenly"
	v := m.View()
	for _, want := range []string{"✓", "◐", "◌", "fix two", "Something is broken.", "a.conf", "✗ quick requests", "watching 02", "stack ● up"} {
		if !strings.Contains(v, want) {
			t.Fatalf("view missing %q:\n%s", want, v)
		}
	}
}
