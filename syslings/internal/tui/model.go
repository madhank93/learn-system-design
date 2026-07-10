// Package tui is the syslings terminal UI in the kubelings/golings house
// style: a two-pane full-screen app — exercises grouped by topic on the left
// with ◌/◐/✓ markers, a detail/output viewport on the right, a splash screen,
// and one-key actions. Every stack/check action is async; file saves re-check
// the exercise being played.
package tui

import (
	"context"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
	"github.com/madhank93/learn-system-design/syslings/internal/runner"
)

type Phase int

const (
	PhaseIdle Phase = iota
	PhaseSettingUp
	PhaseChecking
	PhaseFailed
	PhasePassed
)

type viewMode int

const (
	modeDetail viewMode = iota
	modeOutput
	modeHelp
)

// TopicEntry is one topic module: its exercises and progress.
type TopicEntry struct {
	Name      string
	Desc      string
	Exs       []exercise.Exercise
	St        *exercise.State
	StatePath string
}

// ref addresses one exercise: Topics[T].Exs[E].
type ref struct{ T, E int }

// row is one line of the left list: a topic header or an exercise.
type row struct {
	header string // non-empty → module header (not selectable)
	ref    ref
}

// Messages.
type (
	CheckMsg struct{ Res runner.Result }
	SetupMsg struct{ Err error }
	WatchMsg struct{}
	// EnterMsg starts the playing exercise: fault setup first if declared,
	// then a check.
	EnterMsg struct{}
	// BootDoneMsg reports the topic stack coming up (seed + compose up).
	BootDoneMsg struct {
		Topic string
		Err   error
	}
	// DownDoneMsg reports a stack-down finishing.
	DownDoneMsg struct{ Err error }
	// SolutionMsg reports the shipped solution being applied.
	SolutionMsg struct{ Err error }
	// EditorDoneMsg fires when the external $EDITOR exits.
	EditorDoneMsg struct{ Err error }
)

const checkTimeout = 60 * time.Second

// Hooks are the real-world side effects wired in by main; all topic-aware.
type Hooks struct {
	CheckPath     func(topic string, ex exercise.Exercise) string
	WorkDir       func(topic string) string
	RunSetup      func(topic string, ex exercise.Exercise) error
	ResetFile     func(topic string, ex exercise.Exercise) error
	ApplySolution func(topic string, ex exercise.Exercise) error
	EditorCmd     func(topic string, ex exercise.Exercise) *exec.Cmd // nil if unavailable
	BootTopic     func(topic string) error                           // seed + compose up + watcher
	StackDown     func() error                                       // downs the active stack
}

type Model struct {
	Topics []TopicEntry
	rows   []row
	sel    []int // indices into rows that are exercises
	Cursor int   // index into sel

	// play state
	Playing ref // {-1,-1} = none
	Phase   Phase
	Output  string
	Hints   map[string]int // "topic/id" → hints revealed

	// stack state
	ActiveTopic string // topic whose stack is up ("" = none)
	Booting     bool
	BootErr     error
	pendingBoot string // topic queued behind a stack-down

	// chrome
	Splash     bool
	Mode       viewMode
	ConfirmSol bool
	Quitting   bool
	vp         viewport.Model
	spin       spinner.Model
	w, h       int
	ready      bool

	hooks   Hooks
	stubbed bool // tests: skip real exec
}

// New builds the model from preloaded topics.
func New(topics []TopicEntry, preselect string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = startedStyle
	m := Model{
		Topics:  topics,
		Playing: ref{-1, -1},
		Hints:   map[string]int{},
		Splash:  true,
		spin:    sp,
	}
	m.buildRows()
	// cursor on the preselected topic's (or first) next-incomplete exercise
	for i, s := range m.sel {
		r := m.rows[s].ref
		t := topics[r.T]
		if preselect != "" && t.Name != preselect {
			continue
		}
		if !t.St.Completed[t.Exs[r.E].ID] {
			m.Cursor = i
			break
		}
	}
	if preselect != "" {
		m.Splash = false
	}
	return m
}

func (m *Model) buildRows() {
	m.rows = nil
	m.sel = nil
	for ti, t := range m.Topics {
		m.rows = append(m.rows, row{header: t.Name})
		for ei := range t.Exs {
			m.rows = append(m.rows, row{ref: ref{ti, ei}})
			m.sel = append(m.sel, len(m.rows)-1)
		}
	}
}

// Wire attaches real exec side effects.
func (m Model) Wire(h Hooks) Model {
	m.hooks = h
	return m
}

func (m Model) selRef() ref {
	if len(m.sel) == 0 {
		return ref{-1, -1}
	}
	return m.rows[m.sel[m.Cursor]].ref
}

func (m Model) exAt(r ref) (TopicEntry, exercise.Exercise) {
	if r.T < 0 || r.T >= len(m.Topics) || r.E < 0 || r.E >= len(m.Topics[r.T].Exs) {
		return TopicEntry{}, exercise.Exercise{}
	}
	return m.Topics[r.T], m.Topics[r.T].Exs[r.E]
}

func (m Model) hintKey(r ref) string {
	t, ex := m.exAt(r)
	return t.Name + "/" + ex.ID
}

func (m Model) Init() tea.Cmd { return nil }

// ── async commands ──────────────────────────────────────────────────────────

func (m Model) checkCmd(r ref) tea.Cmd {
	if m.stubbed {
		return func() tea.Msg { return CheckMsg{} }
	}
	t, ex := m.exAt(r)
	return tea.Batch(m.spin.Tick, func() tea.Msg {
		return CheckMsg{Res: runner.Run(context.Background(), m.hooks.CheckPath(t.Name, ex), m.hooks.WorkDir(t.Name), checkTimeout)}
	})
}

func (m Model) setupCmd(r ref) tea.Cmd {
	if m.stubbed {
		return func() tea.Msg { return SetupMsg{} }
	}
	t, ex := m.exAt(r)
	return tea.Batch(m.spin.Tick, func() tea.Msg { return SetupMsg{Err: m.hooks.RunSetup(t.Name, ex)} })
}

func (m Model) bootCmd(topic string) tea.Cmd {
	if m.stubbed {
		return func() tea.Msg { return BootDoneMsg{Topic: topic} }
	}
	return tea.Batch(m.spin.Tick, func() tea.Msg { return BootDoneMsg{Topic: topic, Err: m.hooks.BootTopic(topic)} })
}

func (m Model) downCmd() tea.Cmd {
	if m.stubbed {
		return func() tea.Msg { return DownDoneMsg{} }
	}
	return tea.Batch(m.spin.Tick, func() tea.Msg { return DownDoneMsg{Err: m.hooks.StackDown()} })
}

func (m Model) solutionCmd(r ref) tea.Cmd {
	if m.stubbed {
		return func() tea.Msg { return SolutionMsg{} }
	}
	t, ex := m.exAt(r)
	return tea.Batch(m.spin.Tick, func() tea.Msg { return SolutionMsg{Err: m.hooks.ApplySolution(t.Name, ex)} })
}

// ── update ──────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.layout()
		m.ready = true
		return m, nil

	case spinner.TickMsg:
		if !m.busy() {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case BootDoneMsg:
		m.Booting = false
		if msg.Err != nil {
			m.BootErr = msg.Err
			m.Playing = ref{-1, -1}
			m.Phase = PhaseIdle
			return m, nil
		}
		m.BootErr = nil
		m.ActiveTopic = msg.Topic
		if m.Playing.T >= 0 {
			return m, func() tea.Msg { return EnterMsg{} }
		}
		return m, nil

	case DownDoneMsg:
		m.ActiveTopic = ""
		m.Phase = PhaseIdle
		if m.pendingBoot != "" { // topic switch: old stack down, boot the new
			// one; the queued play (m.Playing) survives the switch
			topic := m.pendingBoot
			m.pendingBoot = ""
			m.Booting = true
			return m, m.bootCmd(topic)
		}
		m.Playing = ref{-1, -1}
		return m, nil

	case EnterMsg:
		_, ex := m.exAt(m.Playing)
		m.Output = ""
		m.Mode = modeOutput
		if ex.Setup != "" {
			m.Phase = PhaseSettingUp
			return m, m.setupCmd(m.Playing)
		}
		m.Phase = PhaseChecking
		return m, m.checkCmd(m.Playing)

	case SetupMsg:
		if msg.Err != nil {
			m.Phase = PhaseFailed
			m.Output = "setup failed: " + msg.Err.Error()
			return m, nil
		}
		m.Phase = PhaseChecking
		return m, m.checkCmd(m.Playing)

	case SolutionMsg:
		if msg.Err != nil {
			m.Phase = PhaseFailed
			m.Output = "could not apply solution: " + msg.Err.Error()
			return m, nil
		}
		m.Phase = PhaseChecking
		return m, m.checkCmd(m.Playing)

	case EditorDoneMsg:
		if m.Playing.T >= 0 && m.Phase != PhaseChecking {
			m.Phase = PhaseChecking
			m.Mode = modeOutput
			return m, m.checkCmd(m.Playing)
		}
		return m, nil

	case WatchMsg:
		if m.Playing.T < 0 || m.Phase == PhaseChecking || m.Booting {
			return m, nil
		}
		m.Phase = PhaseChecking
		m.Mode = modeOutput
		return m, m.checkCmd(m.Playing)

	case CheckMsg:
		m.Output = msg.Res.Output
		m.Mode = modeOutput
		t, ex := m.exAt(m.Playing)
		if !msg.Res.Passed {
			m.Phase = PhaseFailed
			t.St.Attempted[ex.ID] = true
			_ = t.St.Save(t.StatePath)
			return m, nil
		}
		m.Phase = PhasePassed
		t.St.Completed[ex.ID] = true
		_ = t.St.Save(t.StatePath)
		m.Playing = ref{-1, -1}
		m.cursorToNextIncomplete()
		return m, nil

	case tea.KeyMsg:
		return m.onKey(msg)
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) busy() bool {
	return m.Booting || m.Phase == PhaseChecking || m.Phase == PhaseSettingUp
}

func (m *Model) cursorToNextIncomplete() {
	for i, s := range m.sel {
		r := m.rows[s].ref
		t := m.Topics[r.T]
		if !t.St.Completed[t.Exs[r.E].ID] {
			m.Cursor = i
			return
		}
	}
}

func (m Model) onKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Splash {
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		default:
			m.Splash = false
			return m, nil
		}
	}
	if m.ConfirmSol {
		m.ConfirmSol = false
		if msg.String() == "y" || msg.String() == "Y" {
			m.Playing = m.selRef()
			if m.needsBoot(m.Playing) {
				return m.bootFor(m.Playing)
			}
			m.Phase = PhaseChecking
			m.Mode = modeOutput
			return m, m.solutionCmd(m.Playing)
		}
		return m, nil
	}
	if m.Mode == modeHelp {
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		default:
			m.Mode = modeDetail
			return m, nil
		}
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "?":
		m.Mode = modeHelp
	case "a":
		m.Splash = true
	case "esc":
		m.Mode = modeDetail
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			m.Mode = modeDetail
		}
	case "down", "j":
		if m.Cursor < len(m.sel)-1 {
			m.Cursor++
			m.Mode = modeDetail
		}
	case "n":
		m.cursorToNextIncomplete()
		m.Mode = modeDetail
	case "enter", " ":
		return m.play(m.selRef())
	case "h":
		k := m.hintKey(m.selRef())
		_, ex := m.exAt(m.selRef())
		if m.Hints[k] < len(ex.Hints) {
			m.Hints[k]++
		}
		m.Mode = modeDetail
	case "r":
		t, ex := m.exAt(m.selRef())
		if m.hooks.ResetFile != nil && !m.stubbed {
			if err := m.hooks.ResetFile(t.Name, ex); err != nil {
				return m, nil
			}
		}
		if m.Playing == m.selRef() && m.Phase != PhaseChecking {
			m.Phase = PhaseChecking
			m.Mode = modeOutput
			return m, m.checkCmd(m.Playing)
		}
	case "s":
		m.ConfirmSol = true
	case "e":
		if m.hooks.EditorCmd != nil {
			t, ex := m.exAt(m.selRef())
			if c := m.hooks.EditorCmd(t.Name, ex); c != nil {
				return m, tea.ExecProcess(c, func(err error) tea.Msg { return EditorDoneMsg{Err: err} })
			}
		}
	case "u":
		r := m.selRef()
		if t, _ := m.exAt(r); t.Name != "" && m.ActiveTopic == "" && !m.Booting {
			m.Booting = true
			return m, m.bootCmd(t.Name)
		}
	case "d":
		if m.ActiveTopic != "" {
			return m, m.downCmd()
		}
	case "pgup", "pgdown", "ctrl+u", "ctrl+d":
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) needsBoot(r ref) bool {
	t, _ := m.exAt(r)
	return m.ActiveTopic != t.Name
}

// bootFor boots r's topic, downing a different active stack first.
func (m Model) bootFor(r ref) (tea.Model, tea.Cmd) {
	t, _ := m.exAt(r)
	if m.ActiveTopic != "" && m.ActiveTopic != t.Name {
		m.pendingBoot = t.Name
		return m, m.downCmd()
	}
	m.Booting = true
	return m, m.bootCmd(t.Name)
}

// play arms an exercise: boot its topic's stack if needed, then EnterMsg.
func (m Model) play(r ref) (tea.Model, tea.Cmd) {
	if r.T < 0 {
		return m, nil
	}
	m.Playing = r
	m.Output = ""
	m.Mode = modeOutput
	if m.needsBoot(r) {
		if m.Booting {
			return m, nil // play queued behind the in-flight boot
		}
		return m.bootFor(r)
	}
	return m, func() tea.Msg { return EnterMsg{} }
}

func (m *Model) layout() {
	left := leftWidth(m.w)
	bodyH := m.h - 4 // header + footer(2) + pane border
	if bodyH < 3 {
		bodyH = 3
	}
	rightW := m.w - left - 3
	if rightW < 10 {
		rightW = 10
	}
	m.vp = viewport.New(rightW, bodyH)
}
