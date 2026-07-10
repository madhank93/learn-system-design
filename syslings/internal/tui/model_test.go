package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
	"github.com/madhank93/learn-system-design/syslings/internal/runner"
)

func key(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func enter() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }

func testTopics(t *testing.T) []TopicEntry {
	t.Helper()
	return []TopicEntry{
		{
			Name: "load-balancing", Desc: "nginx things",
			StatePath: t.TempDir() + "/.state.json",
			St:        &exercise.State{Completed: map[string]bool{}, Attempted: map[string]bool{}},
			Exs: []exercise.Exercise{
				{ID: "01", Title: "scenario one", Type: exercise.TypeScenario, Check: "01/check.sh", Setup: "01/setup.sh", Hints: []string{"h1"}},
				{ID: "02", Title: "fix two", Desc: "Something is broken.", Type: exercise.TypeFix, File: "a.conf", Check: "02/check.sh", Hints: []string{"h1", "h2"}},
			},
		},
		{
			Name: "caching", Desc: "cache things",
			StatePath: t.TempDir() + "/.state.json",
			St:        &exercise.State{Completed: map[string]bool{}, Attempted: map[string]bool{}},
			Exs: []exercise.Exercise{
				{ID: "01", Title: "cache one", Type: exercise.TypeFix, File: "c.conf", Check: "01/check.sh"},
			},
		},
	}
}

// testModel: splash dismissed, sized, exec stubbed, stack up for load-balancing.
func testModel(t *testing.T) Model {
	m := New(testTopics(t), "")
	m.stubbed = true
	m.Splash = false
	m.ActiveTopic = "load-balancing"
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return next.(Model)
}

func TestRowsGroupByTopicHeader(t *testing.T) {
	m := testModel(t)
	if len(m.rows) != 5 { // 2 headers + 3 exercises
		t.Fatalf("rows = %d, want 5", len(m.rows))
	}
	if m.rows[0].header != "load-balancing" || m.rows[3].header != "caching" {
		t.Fatalf("headers wrong: %+v", m.rows)
	}
	if len(m.sel) != 3 {
		t.Fatalf("selectable = %d, want 3", len(m.sel))
	}
}

func TestSplashOpensAndAnyKeyDismisses(t *testing.T) {
	m := New(testTopics(t), "")
	if !m.Splash {
		t.Fatal("no-preselect launch must show the splash")
	}
	m.stubbed = true
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	next, _ = next.(Model).Update(key("x"))
	if next.(Model).Splash {
		t.Fatal("any key must dismiss the splash")
	}
}

func TestPreselectSkipsSplash(t *testing.T) {
	m := New(testTopics(t), "load-balancing")
	if m.Splash {
		t.Fatal("preselected topic must skip the splash")
	}
}

func TestPlaySameTopicGoesStraightToEnter(t *testing.T) {
	m := testModel(t)
	m.Cursor = 1 // fix two (load-balancing, stack already up)
	next, cmd := m.Update(enter())
	m2 := next.(Model)
	if m2.Playing != (ref{0, 1}) {
		t.Fatalf("playing = %+v", m2.Playing)
	}
	if cmd == nil {
		t.Fatal("want EnterMsg cmd")
	}
	if _, ok := cmd().(EnterMsg); !ok {
		t.Fatal("want EnterMsg")
	}
}

func TestPlayOtherTopicDownsThenBoots(t *testing.T) {
	m := testModel(t) // active = load-balancing
	m.Cursor = 2      // caching exercise
	next, cmd := m.Update(enter())
	m2 := next.(Model)
	if m2.pendingBoot != "caching" || cmd == nil {
		t.Fatalf("want down-then-boot chain, pending=%q", m2.pendingBoot)
	}
	// down completes → boot fires, queued play survives
	next, cmd = m2.Update(DownDoneMsg{})
	m3 := next.(Model)
	if !m3.Booting || cmd == nil {
		t.Fatal("down completion must start the boot")
	}
	if m3.Playing != (ref{1, 0}) {
		t.Fatalf("queued play must survive the switch, got %+v", m3.Playing)
	}
	next, cmd = m3.Update(BootDoneMsg{Topic: "caching"})
	m4 := next.(Model)
	if m4.ActiveTopic != "caching" || cmd == nil {
		t.Fatal("boot completion must set active topic and resume the play")
	}
	if _, ok := cmd().(EnterMsg); !ok {
		t.Fatal("resumed play must fire EnterMsg")
	}
}

func TestPlayWithStackDownBoots(t *testing.T) {
	m := testModel(t)
	m.ActiveTopic = ""
	m.Cursor = 1
	next, cmd := m.Update(enter())
	m2 := next.(Model)
	if !m2.Booting || cmd == nil {
		t.Fatal("play with stack down must boot")
	}
	next, cmd = m2.Update(BootDoneMsg{Topic: "load-balancing"})
	m3 := next.(Model)
	if m3.ActiveTopic != "load-balancing" || cmd == nil {
		t.Fatal("boot completion must resume the queued play")
	}
	if _, ok := cmd().(EnterMsg); !ok {
		t.Fatal("queued play must fire EnterMsg")
	}
}

func TestScenarioSetupThenCheck(t *testing.T) {
	m := testModel(t)
	m.Playing = ref{0, 0}
	next, cmd := m.Update(EnterMsg{})
	m2 := next.(Model)
	if m2.Phase != PhaseSettingUp || cmd == nil {
		t.Fatal("scenario must run setup first")
	}
	next, cmd = m2.Update(SetupMsg{})
	if next.(Model).Phase != PhaseChecking || cmd == nil {
		t.Fatal("setup completion must schedule the check")
	}
}

func TestCheckPassMarksSolvedAndStopsWatching(t *testing.T) {
	m := testModel(t)
	m.Playing = ref{0, 0}
	m.Phase = PhaseChecking
	next, _ := m.Update(CheckMsg{Res: runner.Result{Passed: true, Output: "  ✓ nice"}})
	m2 := next.(Model)
	if !m2.Topics[0].St.Completed["01"] {
		t.Fatal("pass must mark completed")
	}
	if m2.Playing.T != -1 {
		t.Fatal("pass must stop watching")
	}
	if m2.Phase != PhasePassed {
		t.Fatalf("phase = %v", m2.Phase)
	}
	// cursor highlights next incomplete (index 1: fix two)
	if m2.selRef() != (ref{0, 1}) {
		t.Fatalf("cursor ref = %+v, want {0 1}", m2.selRef())
	}
}

func TestCheckFailMarksAttempted(t *testing.T) {
	m := testModel(t)
	m.Playing = ref{0, 1}
	next, _ := m.Update(CheckMsg{Res: runner.Result{Passed: false, Output: "  ✗ broke"}})
	m2 := next.(Model)
	if m2.Phase != PhaseFailed || !m2.Topics[0].St.Attempted["02"] {
		t.Fatal("fail must mark attempted (◐)")
	}
}

func TestWatchRechecksOnlyWhilePlaying(t *testing.T) {
	m := testModel(t)
	if _, cmd := m.Update(WatchMsg{}); cmd != nil {
		t.Fatal("watch ignored when nothing playing")
	}
	m.Playing = ref{0, 1}
	m.Phase = PhaseFailed
	next, cmd := m.Update(WatchMsg{})
	if next.(Model).Phase != PhaseChecking || cmd == nil {
		t.Fatal("watch while playing must re-check")
	}
}

func TestHintsPerExercise(t *testing.T) {
	m := testModel(t)
	m.Cursor = 1
	next, _ := m.Update(key("h"))
	next, _ = next.(Model).Update(key("h"))
	next, _ = next.(Model).Update(key("h"))
	m2 := next.(Model)
	if m2.Hints["load-balancing/02"] != 2 {
		t.Fatalf("hints = %d, want capped 2", m2.Hints["load-balancing/02"])
	}
	if m2.Hints["load-balancing/01"] != 0 {
		t.Fatal("hints tracked per exercise")
	}
}

func TestSolutionConfirmFlow(t *testing.T) {
	m := testModel(t)
	m.Cursor = 1
	next, _ := m.Update(key("s"))
	m2 := next.(Model)
	if !m2.ConfirmSol {
		t.Fatal("s must ask for confirmation")
	}
	next, _ = m2.Update(key("n"))
	if next.(Model).ConfirmSol {
		t.Fatal("n must decline")
	}
	m2.ConfirmSol = true
	next, cmd := m2.Update(key("y"))
	m3 := next.(Model)
	if cmd == nil || m3.Playing != (ref{0, 1}) {
		t.Fatal("y must apply the solution to the selected exercise")
	}
}

func TestNextJumpsToNextUnfinished(t *testing.T) {
	m := testModel(t)
	m.Topics[0].St.Completed["01"] = true
	m.Cursor = 2
	next, _ := m.Update(key("n"))
	if got := next.(Model).selRef(); got != (ref{0, 1}) {
		t.Fatalf("n → %+v, want {0 1}", got)
	}
}

func TestHelpModeAndBack(t *testing.T) {
	m := testModel(t)
	next, _ := m.Update(key("?"))
	m2 := next.(Model)
	if m2.Mode != modeHelp {
		t.Fatal("? must open help")
	}
	if !strings.Contains(m2.View(), "markers") {
		t.Fatal("help must document markers")
	}
	next, _ = m2.Update(key("x"))
	if next.(Model).Mode != modeDetail {
		t.Fatal("any key must close help")
	}
}

func TestViewRendersHouseStyle(t *testing.T) {
	m := testModel(t)
	m.Topics[0].St.Completed["01"] = true
	m.Topics[0].St.Attempted["02"] = true
	m.Cursor = 1
	m.Playing = ref{0, 1}
	m.Phase = PhaseFailed
	m.Output = "  ✗ quick requests spread evenly\n  hint: check the directive"
	m.Mode = modeOutput
	v := m.View()
	for _, want := range []string{"syslings", "stack:", "● up", "load-balancing 1/2", "caching 0/1", "✓", "◌", "fix two", "✗ quick requests", "watching load-balancing/02"} {
		if !strings.Contains(v, want) {
			t.Fatalf("view missing %q:\n%s", want, v)
		}
	}
}

func TestDetailPaneShowsTypeAndFile(t *testing.T) {
	m := testModel(t)
	m.Cursor = 1
	m.Mode = modeDetail
	v := m.View()
	for _, want := range []string{"Something is broken.", "work/load-balancing/a.conf", "✎ fix", "Stack"} {
		if !strings.Contains(v, want) {
			t.Fatalf("detail missing %q:\n%s", want, v)
		}
	}
}

func TestStackDownKey(t *testing.T) {
	m := testModel(t)
	m.Playing = ref{0, 1}
	next, cmd := m.Update(key("d"))
	if cmd == nil {
		t.Fatal("d with stack up must schedule down")
	}
	next, _ = next.(Model).Update(DownDoneMsg{})
	m2 := next.(Model)
	if m2.ActiveTopic != "" || m2.Playing.T != -1 {
		t.Fatal("down must clear active topic and stop playing")
	}
}

func TestQuit(t *testing.T) {
	m := testModel(t)
	next, cmd := m.Update(key("q"))
	if !next.(Model).Quitting || cmd == nil {
		t.Fatal("q must quit")
	}
}
