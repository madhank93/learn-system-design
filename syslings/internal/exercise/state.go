package exercise

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// State is the learner's progress, persisted as JSON in the work dir.
type State struct {
	Completed map[string]bool `json:"completed"`
}

// LoadState reads progress from path; a missing file is a fresh start.
func LoadState(path string) (*State, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &State{Completed: map[string]bool{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state %s: %w", path, err)
	}
	var st State
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil, fmt.Errorf("parse state %s: %w", path, err)
	}
	if st.Completed == nil {
		st.Completed = map[string]bool{}
	}
	return &st, nil
}

func (s *State) Save(path string) error {
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

// NextIncomplete returns the index of the first exercise not yet completed,
// or -1 when everything is done.
func NextIncomplete(exs []Exercise, st *State) int {
	for i, ex := range exs {
		if !st.Completed[ex.ID] {
			return i
		}
	}
	return -1
}
