// syslings — rustlings-style hands-on system-design exercises where the
// "compiler" is a live Docker stack.
//
//	syslings [topic]          interactive TUI (default topic: load-balancing)
//	syslings verify [topic]   headless: prove every exercise solvable (CI)
//	syslings reset <id>       restore an exercise's work file from its template
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
	"github.com/madhank93/learn-system-design/syslings/internal/runner"
	"github.com/madhank93/learn-system-design/syslings/internal/stack"
	"github.com/madhank93/learn-system-design/syslings/internal/tui"
	"github.com/madhank93/learn-system-design/syslings/internal/watch"
)

// topics maps a topic name to its compose dir (relative to repo root) and a
// probe URL that answers once the stack is up.
var topics = map[string]struct {
	dir   string
	probe string
}{
	"load-balancing": {
		dir:   "high-level-design/networking-and-edge/2.1_load_balancer",
		probe: "http://localhost:8081",
	},
}

type layout struct {
	repoRoot     string
	topic        string
	exercisesDir string // syslings/exercises/<topic>
	workDir      string // syslings/work/<topic>
	topicDir     string // compose project dir
	overridePath string // exercises/<topic>/compose.override.yml
	statePath    string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "syslings:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	mode, topic := "tui", "load-balancing"
	switch {
	case len(args) == 0:
	case args[0] == "verify":
		mode = "verify"
		if len(args) > 1 {
			topic = args[1]
		}
	case args[0] == "reset":
		if len(args) < 2 {
			return fmt.Errorf("usage: syslings reset <exercise-id> [topic]")
		}
		mode = "reset:" + args[1]
		if len(args) > 2 {
			topic = args[2]
		}
	default:
		topic = args[0]
	}

	lay, err := resolveLayout(topic)
	if err != nil {
		return err
	}
	exs, err := exercise.Load(filepath.Join(lay.exercisesDir, "exercises.yaml"))
	if err != nil {
		return err
	}

	// scripts (check.sh / setup.sh / solution.sh) find the stack through env
	os.Setenv("SYSLINGS_TOPIC_DIR", lay.topicDir)
	os.Setenv("SYSLINGS_OVERRIDE", lay.overridePath)
	os.Setenv("SYSLINGS_WORK", lay.workDir)

	switch {
	case mode == "verify":
		return verify(lay, exs)
	case len(mode) > 6 && mode[:6] == "reset:":
		return resetExercise(lay, exs, mode[6:])
	default:
		return interactive(lay, exs)
	}
}

// resolveLayout finds the repo root (dir containing go.work) upward from cwd
// and derives every path syslings needs.
func resolveLayout(topic string) (layout, error) {
	t, ok := topics[topic]
	if !ok {
		return layout{}, fmt.Errorf("unknown topic %q (available: load-balancing)", topic)
	}
	dir, err := os.Getwd()
	if err != nil {
		return layout{}, err
	}
	root := dir
	for {
		if _, err := os.Stat(filepath.Join(root, "go.work")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return layout{}, fmt.Errorf("repo root not found (no go.work above %s)", dir)
		}
		root = parent
	}
	lay := layout{
		repoRoot:     root,
		topic:        topic,
		exercisesDir: filepath.Join(root, "syslings", "exercises", topic),
		workDir:      filepath.Join(root, "syslings", "work", topic),
		topicDir:     filepath.Join(root, t.dir),
		statePath:    filepath.Join(root, "syslings", "work", topic, ".state.json"),
	}
	lay.overridePath = filepath.Join(lay.exercisesDir, "compose.override.yml")
	return lay, nil
}

// seed copies each fix exercise's broken template into the work dir unless the
// learner already has a copy (never clobber their edits).
func seed(lay layout, exs []exercise.Exercise, force map[string]bool) error {
	if err := os.MkdirAll(lay.workDir, 0o755); err != nil {
		return err
	}
	for _, ex := range exs {
		if ex.Type != exercise.TypeFix {
			continue
		}
		dst := filepath.Join(lay.workDir, ex.File)
		if _, err := os.Stat(dst); err == nil && !force[ex.ID] {
			continue
		}
		src := filepath.Join(lay.exercisesDir, ex.ID, ex.File)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("seed %s: %w", ex.ID, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func newStack(lay layout) *stack.Stack {
	return &stack.Stack{TopicDir: lay.topicDir, OverridePath: lay.overridePath, WorkDir: lay.workDir}
}

func runScript(lay layout, rel string) error {
	cmd := exec.Command("bash", filepath.Join(lay.exercisesDir, rel))
	cmd.Dir = lay.workDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func interactive(lay layout, exs []exercise.Exercise) error {
	if err := seed(lay, exs, nil); err != nil {
		return err
	}
	st, err := exercise.LoadState(lay.statePath)
	if err != nil {
		return err
	}

	fmt.Println("🐳 starting the", lay.topic, "stack (first run builds images)…")
	s := newStack(lay)
	ctx := context.Background()
	if err := s.Up(ctx); err != nil {
		return err
	}
	if err := s.WaitUp(ctx, topics[lay.topic].probe, 60*time.Second); err != nil {
		return err
	}

	m := tui.New(lay.topic, exs, st, lay.statePath).Wire(tui.Hooks{
		CheckPath: func(ex exercise.Exercise) string { return filepath.Join(lay.exercisesDir, ex.Check) },
		WorkDir:   lay.workDir,
		RunSetup:  func(ex exercise.Exercise) error { return runScriptQuiet(lay, ex.Setup) },
		ResetFile: func(ex exercise.Exercise) error {
			if ex.File == "" {
				return nil
			}
			return seed(lay, []exercise.Exercise{ex}, map[string]bool{ex.ID: true})
		},
	})

	p := tea.NewProgram(m)

	w, err := watch.New(lay.workDir, 300*time.Millisecond)
	if err != nil {
		return err
	}
	defer w.Close()
	go func() {
		for range w.Events() {
			p.Send(tui.WatchMsg{})
		}
	}()

	if _, err := p.Run(); err != nil {
		return err
	}
	fmt.Println("🐳 tearing the stack down…")
	return s.Down(context.Background())
}

// runScriptQuiet is runScript without dumping into the TUI's terminal.
func runScriptQuiet(lay layout, rel string) error {
	cmd := exec.Command("bash", filepath.Join(lay.exercisesDir, rel))
	cmd.Dir = lay.workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

// verify proves every exercise is solvable: seed broken templates, stand the
// stack up, then per exercise apply the shipped solution and require the
// check to pass. Mirrors rustlings' CI.
func verify(lay layout, exs []exercise.Exercise) error {
	// fresh work dir so learner progress never affects CI
	if err := os.RemoveAll(lay.workDir); err != nil {
		return err
	}
	if err := seed(lay, exs, nil); err != nil {
		return err
	}

	s := newStack(lay)
	ctx := context.Background()
	fmt.Println("── syslings verify:", lay.topic, "──")
	if err := s.Up(ctx); err != nil {
		return err
	}
	defer s.Down(context.Background())
	if err := s.WaitUp(ctx, topics[lay.topic].probe, 60*time.Second); err != nil {
		return err
	}

	for _, ex := range exs {
		fmt.Printf("▶ %s: %s\n", ex.ID, ex.Title)
		if ex.Setup != "" {
			if err := runScript(lay, ex.Setup); err != nil {
				return fmt.Errorf("%s setup: %w", ex.ID, err)
			}
		}
		// apply the solution: file for fix exercises, script for scenarios
		if ex.Type == exercise.TypeFix {
			sol := filepath.Join(lay.exercisesDir, ex.ID, "solution", ex.File)
			if err := copyFile(sol, filepath.Join(lay.workDir, ex.File)); err != nil {
				return fmt.Errorf("%s solution: %w", ex.ID, err)
			}
		}
		if solScript := filepath.Join(lay.exercisesDir, ex.ID, "solution.sh"); fileExists(solScript) {
			if err := runScript(lay, filepath.Join(ex.ID, "solution.sh")); err != nil {
				return fmt.Errorf("%s solution.sh: %w", ex.ID, err)
			}
		}
		res := runner.Run(ctx, filepath.Join(lay.exercisesDir, ex.Check), lay.workDir, 90*time.Second)
		fmt.Print(res.Output)
		if !res.Passed {
			return fmt.Errorf("%s: solution did not pass its check", ex.ID)
		}
	}
	fmt.Println("── syslings verify:", lay.topic, "ALL SOLVABLE ──")
	return nil
}

func resetExercise(lay layout, exs []exercise.Exercise, id string) error {
	for _, ex := range exs {
		if ex.ID == id {
			if ex.File == "" {
				return fmt.Errorf("%s has no work file to reset", id)
			}
			if err := seed(lay, []exercise.Exercise{ex}, map[string]bool{id: true}); err != nil {
				return err
			}
			fmt.Println("reset", ex.File, "from template")
			return nil
		}
	}
	return fmt.Errorf("no exercise %q", id)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
