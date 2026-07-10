// Package tui is the syslings terminal UI, following the kubelings/golings
// house style: the exercise list IS the home screen — ◌/◐/✓ progress markers,
// one-key actions on the selected row, a detail pane for output and hints,
// and file-watch re-checking for the exercise being played.
package tui

import (
	"context"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
	"github.com/madhank93/learn-system-design/syslings/internal/runner"
)

type Phase int

const (
	PhaseIdle      Phase = iota // nothing being played
	PhaseSettingUp              // scenario/fault setup running
	PhaseChecking
	PhaseFailed
	PhasePassed
)

// Messages.
type (
	CheckMsg struct{ Res runner.Result }
	SetupMsg struct{ Err error }
	WatchMsg struct{}
	// EnterMsg starts (or restarts) the playing exercise: fault setup first
	// if it declares one, then a check.
	EnterMsg struct{}
	// BootDoneMsg reports the topic stack coming up (seed + compose up).
	BootDoneMsg struct{ Err error }
	// DownDoneMsg reports a manual stack-down finishing.
	DownDoneMsg struct{ Err error }
	// SolutionMsg reports the shipped solution being applied.
	SolutionMsg struct{ Err error }
	// EditorDoneMsg fires when the external $EDITOR exits.
	EditorDoneMsg struct{ Err error }
)

// TopicEntry is one selectable topic in the topic picker.
type TopicEntry struct {
	Name      string
	Desc      string
	Exs       []exercise.Exercise
	St        *exercise.State
	StatePath string
}

const checkTimeout = 60 * time.Second

type Model struct {
	Topic     string
	Exs       []exercise.Exercise
	St        *exercise.State
	StatePath string

	// home-screen list
	Cursor  int
	Playing int // index being played/watched; -1 = none
	Phase   Phase
	Output  string
	Hints   map[string]int // exercise id → hints revealed

	// modes / overlays
	ShowHelp       bool
	ConfirmSolItem int // index awaiting solution confirm; -1 = none
	Quitting       bool

	// stack
	StackUp bool
	DownErr error

	// topic picker (open on launch unless a topic was preselected)
	Topics      []TopicEntry
	TopicOpen   bool
	TopicCursor int
	Booting     bool
	BootErr     error

	// exec seams — replaced by stubs in tests.
	check     func(ex exercise.Exercise) tea.Cmd
	setup     func(ex exercise.Exercise) tea.Cmd
	resetFile func(ex exercise.Exercise) error
	bootTopic func(name string) tea.Cmd
	stackDown func() tea.Cmd
	solution  func(ex exercise.Exercise) tea.Cmd
	editorCmd func(ex exercise.Exercise) *exec.Cmd
}

// Hooks are the real-world side effects wired in by main. Paths resolve
// against the ACTIVE topic (set when BootTopic succeeds).
type Hooks struct {
	CheckPath     func(ex exercise.Exercise) string // absolute path to check.sh
	WorkDir       func() string                     // active topic's work dir
	RunSetup      func(ex exercise.Exercise) error
	ResetFile     func(ex exercise.Exercise) error
	BootTopic     func(name string) error // seed + compose up + watcher
	StackDown     func() error
	ApplySolution func(ex exercise.Exercise) error
	EditorCmd     func(ex exercise.Exercise) *exec.Cmd // nil if no $EDITOR / no file
}

// NewPicker builds the entry point: topic picker first (unless a topic was
// preselected on the CLI), Docker boot on selection, then the exercise home.
func NewPicker(topics []TopicEntry, preselect string) Model {
	m := Model{
		Topics:         topics,
		TopicOpen:      true,
		Playing:        -1,
		ConfirmSolItem: -1,
		Hints:          map[string]int{},
		resetFile:      func(exercise.Exercise) error { return nil },
	}
	for i, t := range topics {
		if t.Name == preselect {
			m = m.activateTopic(i)
			break
		}
	}
	return m
}

// New builds a single-topic model with the stack assumed up (used in tests).
func New(topic string, exs []exercise.Exercise, st *exercise.State, statePath string) Model {
	m := Model{
		Topic:          topic,
		Exs:            exs,
		St:             st,
		StatePath:      statePath,
		Playing:        -1,
		ConfirmSolItem: -1,
		StackUp:        true,
		Hints:          map[string]int{},
		resetFile:      func(exercise.Exercise) error { return nil },
	}
	m.Cursor = clampIndex(exercise.NextIncomplete(exs, st), len(exs))
	return m
}

func clampIndex(i, n int) int {
	if i < 0 || i >= n {
		return 0
	}
	return i
}

// activateTopic loads a topic's data into the model and moves to booting.
func (m Model) activateTopic(i int) Model {
	t := m.Topics[i]
	m.Topic = t.Name
	m.Exs = t.Exs
	m.St = t.St
	m.StatePath = t.StatePath
	m.Cursor = clampIndex(exercise.NextIncomplete(m.Exs, m.St), len(m.Exs))
	m.Playing = -1
	m.TopicOpen = false
	m.Booting = true
	m.BootErr = nil
	return m
}

// Wire attaches real exec side effects.
func (m Model) Wire(h Hooks) Model {
	m.check = func(ex exercise.Exercise) tea.Cmd {
		return func() tea.Msg {
			return CheckMsg{Res: runner.Run(context.Background(), h.CheckPath(ex), h.WorkDir(), checkTimeout)}
		}
	}
	m.setup = func(ex exercise.Exercise) tea.Cmd {
		return func() tea.Msg { return SetupMsg{Err: h.RunSetup(ex)} }
	}
	m.resetFile = h.ResetFile
	if h.BootTopic != nil {
		m.bootTopic = func(name string) tea.Cmd {
			return func() tea.Msg { return BootDoneMsg{Err: h.BootTopic(name)} }
		}
	}
	if h.StackDown != nil {
		m.stackDown = func() tea.Cmd {
			return func() tea.Msg { return DownDoneMsg{Err: h.StackDown()} }
		}
	}
	if h.ApplySolution != nil {
		m.solution = func(ex exercise.Exercise) tea.Cmd {
			return func() tea.Msg { return SolutionMsg{Err: h.ApplySolution(ex)} }
		}
	}
	m.editorCmd = h.EditorCmd
	return m
}

func (m Model) Init() tea.Cmd {
	if m.Booting && m.bootTopic != nil {
		return m.bootTopic(m.Topic)
	}
	return nil
}

func (m Model) selected() exercise.Exercise {
	if m.Cursor >= 0 && m.Cursor < len(m.Exs) {
		return m.Exs[m.Cursor]
	}
	return exercise.Exercise{}
}

func (m Model) playing() exercise.Exercise {
	if m.Playing >= 0 && m.Playing < len(m.Exs) {
		return m.Exs[m.Playing]
	}
	return exercise.Exercise{}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case BootDoneMsg:
		m.Booting = false
		if msg.Err != nil {
			m.BootErr = msg.Err
			return m, nil
		}
		m.StackUp = true
		if m.Playing >= 0 { // play was queued behind the boot
			return m, func() tea.Msg { return EnterMsg{} }
		}
		return m, nil

	case DownDoneMsg:
		m.StackUp = false
		m.Playing = -1
		m.Phase = PhaseIdle
		m.DownErr = msg.Err
		return m, nil

	case EnterMsg:
		ex := m.playing()
		m.Output = ""
		if ex.Setup != "" { // fault injection runs for any exercise that declares one
			m.Phase = PhaseSettingUp
			return m, m.setup(ex)
		}
		m.Phase = PhaseChecking
		return m, m.check(ex)

	case SetupMsg:
		if msg.Err != nil {
			m.Phase = PhaseFailed
			m.Output = "setup failed: " + msg.Err.Error()
			return m, nil
		}
		m.Phase = PhaseChecking
		return m, m.check(m.playing())

	case SolutionMsg:
		if msg.Err != nil {
			m.Phase = PhaseFailed
			m.Output = "could not apply solution: " + msg.Err.Error()
			return m, nil
		}
		m.Phase = PhaseChecking
		return m, m.check(m.playing())

	case EditorDoneMsg:
		// the watcher usually catches the save, but re-check explicitly so
		// editors that write via rename never leave a stale verdict
		if m.Playing >= 0 && m.Phase != PhaseChecking {
			m.Phase = PhaseChecking
			return m, m.check(m.playing())
		}
		return m, nil

	case WatchMsg:
		if m.Playing < 0 || m.Phase == PhaseChecking || m.TopicOpen || m.Booting {
			return m, nil
		}
		m.Phase = PhaseChecking
		return m, m.check(m.playing())

	case CheckMsg:
		m.Output = msg.Res.Output
		id := m.playing().ID
		if !msg.Res.Passed {
			m.Phase = PhaseFailed
			m.St.Attempted[id] = true
			_ = m.St.Save(m.StatePath)
			return m, nil
		}
		m.Phase = PhasePassed
		m.St.Completed[id] = true
		_ = m.St.Save(m.StatePath)
		// stop watching; highlight the next unfinished exercise (golings
		// style) but let the learner press ⏎ to start it
		m.Playing = -1
		if next := exercise.NextIncomplete(m.Exs, m.St); next >= 0 {
			m.Cursor = next
		}
		return m, nil

	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m Model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.TopicOpen {
		return m.topicKey(msg)
	}
	if m.ConfirmSolItem >= 0 {
		return m.confirmKey(msg)
	}
	if m.ShowHelp {
		m.ShowHelp = false // any key closes help
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "?":
		m.ShowHelp = true
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Exs)-1 {
			m.Cursor++
		}
	case "n":
		if next := exercise.NextIncomplete(m.Exs, m.St); next >= 0 {
			m.Cursor = next
		}
	case "enter", " ":
		return m.play(m.Cursor)
	case "h":
		id := m.selected().ID
		if m.Hints[id] < len(m.selected().Hints) {
			m.Hints[id]++
		}
	case "r":
		ex := m.selected()
		if err := m.resetFile(ex); err == nil && m.Playing == m.Cursor && m.Phase != PhaseChecking {
			m.Phase = PhaseChecking
			return m, m.check(ex)
		}
	case "s":
		if m.solution != nil {
			m.ConfirmSolItem = m.Cursor
		}
	case "e":
		if m.editorCmd != nil {
			if c := m.editorCmd(m.selected()); c != nil {
				return m, tea.ExecProcess(c, func(err error) tea.Msg { return EditorDoneMsg{Err: err} })
			}
		}
	case "u":
		if !m.StackUp && !m.Booting && m.bootTopic != nil {
			m.Booting = true
			return m, m.bootTopic(m.Topic)
		}
	case "d":
		if m.StackUp && m.stackDown != nil {
			return m, m.stackDown()
		}
	}
	return m, nil
}

// play arms an exercise: boots the stack first if needed, then EnterMsg.
func (m Model) play(i int) (tea.Model, tea.Cmd) {
	m.Playing = i
	m.Cursor = i
	m.Output = ""
	if !m.StackUp {
		if m.Booting || m.bootTopic == nil {
			return m, nil // boot already in flight; play is queued
		}
		m.Booting = true
		return m, m.bootTopic(m.Topic)
	}
	return m, func() tea.Msg { return EnterMsg{} }
}

func (m Model) confirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	i := m.ConfirmSolItem
	m.ConfirmSolItem = -1
	if msg.String() != "y" {
		return m, nil
	}
	m.Playing = i
	m.Cursor = i
	if !m.StackUp {
		return m.play(i) // solution needs the stack; check runs after boot
	}
	m.Phase = PhaseChecking
	return m, m.solution(m.Exs[i])
}

func (m Model) topicKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.TopicCursor > 0 {
			m.TopicCursor--
		}
	case "down", "j":
		if m.TopicCursor < len(m.Topics)-1 {
			m.TopicCursor++
		}
	case "enter":
		m = m.activateTopic(m.TopicCursor)
		if m.bootTopic != nil {
			return m, m.bootTopic(m.Topic)
		}
	}
	return m, nil
}
