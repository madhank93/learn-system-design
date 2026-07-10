// Package tui is the rustlings-style interactive loop: show the current
// exercise, watch the work dir, re-check on save, advance on green.
package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
	"github.com/madhank93/learn-system-design/syslings/internal/runner"
)

type Phase int

const (
	PhaseSettingUp Phase = iota // scenario setup.sh running
	PhaseChecking
	PhaseFailed
	PhasePassed // brief state between pass and advancing
	PhaseAllDone
)

// Messages.
type (
	CheckMsg struct{ Res runner.Result }
	SetupMsg struct{ Err error }
	WatchMsg struct{}
	// EnterMsg starts (or restarts) the current exercise: scenario setup
	// first if needed, then a check.
	EnterMsg struct{}
	// BootDoneMsg reports the topic stack coming up (seed + compose up).
	BootDoneMsg struct{ Err error }
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

	Cur        int
	Phase      Phase
	Output     string
	HintsShown int
	Quitting   bool
	SetupErr   error

	// exercise picker (open after boot; `l` reopens)
	ListOpen   bool
	ListCursor int

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
}

// Hooks are the real-world side effects wired in by main. Paths resolve
// against the ACTIVE topic (set when BootTopic succeeds).
type Hooks struct {
	CheckPath func(ex exercise.Exercise) string // absolute path to check.sh
	WorkDir   func() string                     // active topic's work dir
	RunSetup  func(ex exercise.Exercise) error
	ResetFile func(ex exercise.Exercise) error
	BootTopic func(name string) error // seed + compose up + watcher
}

func New(topic string, exs []exercise.Exercise, st *exercise.State, statePath string) Model {
	cur := exercise.NextIncomplete(exs, st)
	if cur < 0 {
		cur = 0
	}
	return Model{
		Topic:      topic,
		Exs:        exs,
		St:         st,
		StatePath:  statePath,
		Cur:        cur,
		Phase:      PhaseChecking,
		ListOpen:   true,
		ListCursor: cur,
		resetFile:  func(exercise.Exercise) error { return nil },
	}
}

// NewPicker builds the multi-topic entry point: topic picker first (unless a
// topic was preselected on the CLI), Docker boot on selection, then the
// exercise picker. Exercise/state data for every topic is preloaded.
func NewPicker(topics []TopicEntry, preselect string) Model {
	m := Model{
		Topics:    topics,
		TopicOpen: true,
		resetFile: func(exercise.Exercise) error { return nil },
	}
	for i, t := range topics {
		if t.Name == preselect {
			m = m.activateTopic(i)
			break
		}
	}
	return m
}

// activateTopic loads a topic's data into the model and moves to booting.
func (m Model) activateTopic(i int) Model {
	t := m.Topics[i]
	m.Topic = t.Name
	m.Exs = t.Exs
	m.St = t.St
	m.StatePath = t.StatePath
	cur := exercise.NextIncomplete(m.Exs, m.St)
	if cur < 0 {
		cur = 0
	}
	m.Cur = cur
	m.ListCursor = cur
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
	return m
}

func (m Model) Init() tea.Cmd {
	if m.Booting && m.bootTopic != nil {
		return m.bootTopic(m.Topic)
	}
	if m.TopicOpen || m.ListOpen {
		return nil // wait for the learner to pick
	}
	return func() tea.Msg { return EnterMsg{} }
}

func (m Model) current() exercise.Exercise {
	if m.Cur >= 0 && m.Cur < len(m.Exs) {
		return m.Exs[m.Cur]
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
		m.ListOpen = true
		return m, nil

	case EnterMsg:
		// runs the current exercise even if already completed — the picker
		// lets learners revisit; AllDone is decided when a check passes.
		ex := m.current()
		m.HintsShown = 0
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
			m.SetupErr = msg.Err
			m.Output = "setup failed: " + msg.Err.Error()
			return m, nil
		}
		m.Phase = PhaseChecking
		return m, m.check(m.current())

	case WatchMsg:
		if m.ListOpen || m.Phase == PhaseChecking || m.Phase == PhaseAllDone {
			return m, nil
		}
		m.Phase = PhaseChecking
		return m, m.check(m.current())

	case CheckMsg:
		m.Output = msg.Res.Output
		if !msg.Res.Passed {
			m.Phase = PhaseFailed
			return m, nil
		}
		m.St.Completed[m.current().ID] = true
		_ = m.St.Save(m.StatePath)
		next := exercise.NextIncomplete(m.Exs, m.St)
		if next < 0 {
			m.Phase = PhaseAllDone
			return m, nil
		}
		m.Cur = next
		return m, func() tea.Msg { return EnterMsg{} }

	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m Model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.TopicOpen {
		return m.topicKey(msg)
	}
	if m.ListOpen {
		return m.pickerKey(msg)
	}
	switch msg.String() {
	case "q", "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "l":
		m.ListOpen = true
		m.ListCursor = m.Cur
		return m, nil
	case "h":
		if n := len(m.current().Hints); m.HintsShown < n {
			m.HintsShown++
		}
		return m, nil
	case "r":
		if err := m.resetFile(m.current()); err == nil {
			m.Phase = PhaseChecking
			return m, m.check(m.current())
		}
		return m, nil
	case "s":
		if m.Cur < len(m.Exs)-1 {
			m.Cur++
			return m, func() tea.Msg { return EnterMsg{} }
		}
		return m, nil
	}
	return m, nil
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

func (m Model) pickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.ListCursor > 0 {
			m.ListCursor--
		}
	case "down", "j":
		if m.ListCursor < len(m.Exs)-1 {
			m.ListCursor++
		}
	case "enter", "l":
		m.ListOpen = false
		m.Cur = m.ListCursor
		return m, func() tea.Msg { return EnterMsg{} }
	}
	return m, nil
}
