package exercise

import (
	"path/filepath"
	"testing"
)

func sampleExercises() []Exercise {
	return []Exercise{
		{ID: "01", Title: "one", Type: TypeScenario, Check: "01/check.sh"},
		{ID: "02", Title: "two", Type: TypeFix, File: "a.conf", Check: "02/check.sh"},
		{ID: "03", Title: "three", Type: TypeFix, File: "b.conf", Check: "03/check.sh"},
	}
}

func TestLoadStateMissingFileGivesEmptyState(t *testing.T) {
	st, err := LoadState(filepath.Join(t.TempDir(), ".state.json"))
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(st.Completed) != 0 {
		t.Fatalf("want empty completed, got %v", st.Completed)
	}
}

func TestStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".state.json")
	st := &State{Completed: map[string]bool{"01": true}}
	if err := st.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if !got.Completed["01"] || got.Completed["02"] {
		t.Fatalf("round trip wrong: %v", got.Completed)
	}
}

func TestNextIncomplete(t *testing.T) {
	exs := sampleExercises()
	st := &State{Completed: map[string]bool{"01": true}}
	if i := NextIncomplete(exs, st); i != 1 {
		t.Fatalf("want index 1, got %d", i)
	}
	st.Completed["02"] = true
	st.Completed["03"] = true
	if i := NextIncomplete(exs, st); i != -1 {
		t.Fatalf("want -1 when all done, got %d", i)
	}
}

func TestNextIncompleteEmptyState(t *testing.T) {
	if i := NextIncomplete(sampleExercises(), &State{Completed: map[string]bool{}}); i != 0 {
		t.Fatalf("want 0, got %d", i)
	}
}
