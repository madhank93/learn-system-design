// Package exercise loads and tracks syslings exercises: ordered, checkable
// units of "fix this broken config" or "remediate this live-stack fault".
package exercise

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Type string

const (
	// TypeFix ships a broken file the learner edits until the check passes.
	TypeFix Type = "fix"
	// TypeScenario injects a fault into the running stack via setup.sh;
	// the check passes once the learner has remediated it.
	TypeScenario Type = "scenario"
)

type Exercise struct {
	ID    string   `yaml:"id"`
	Title string   `yaml:"title"`
	Desc  string   `yaml:"desc"` // one-liner shown in the exercise picker
	Type  Type     `yaml:"type"`
	File  string   `yaml:"file"`  // fix only: work file the learner edits
	Check string   `yaml:"check"` // path to check.sh, relative to the exercises dir
	Setup string   `yaml:"setup"` // scenario only: fault-injection script
	Hints []string `yaml:"hints"`
}

// Load parses an exercises.yaml and validates the invariants the rest of
// syslings relies on: unique ids, known types, fix exercises name a file,
// every exercise has a check.
func Load(path string) ([]Exercise, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var exs []Exercise
	if err := yaml.Unmarshal(raw, &exs); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	seen := make(map[string]bool, len(exs))
	for i, ex := range exs {
		switch {
		case ex.ID == "":
			return nil, fmt.Errorf("exercise %d: missing id", i)
		case seen[ex.ID]:
			return nil, fmt.Errorf("duplicate exercise id %q", ex.ID)
		case ex.Type != TypeFix && ex.Type != TypeScenario:
			return nil, fmt.Errorf("exercise %s: unknown type %q", ex.ID, ex.Type)
		case ex.Type == TypeFix && ex.File == "":
			return nil, fmt.Errorf("exercise %s: fix type requires a file", ex.ID)
		case ex.Check == "":
			return nil, fmt.Errorf("exercise %s: missing check", ex.ID)
		}
		seen[ex.ID] = true
	}
	return exs, nil
}
