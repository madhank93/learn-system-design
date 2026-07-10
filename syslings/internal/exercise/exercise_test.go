package exercise

import (
	"os"
	"path/filepath"
	"testing"
)

const validYAML = `
- id: 01_all_aboard
  title: "All aboard"
  type: scenario
  check: 01_all_aboard/check.sh
  hints:
    - "curl the four ports"
- id: 02_round_robin
  title: "Round robin plays favorites"
  desc: "One server hogs every request — restore the fair rotation."
  type: fix
  file: round_robin.conf
  check: 02_round_robin/check.sh
  hints:
    - "look at the weight"
    - "remove it"
`

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "exercises.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadParsesOrderedExercises(t *testing.T) {
	exs, err := Load(writeTemp(t, validYAML))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(exs) != 2 {
		t.Fatalf("got %d exercises, want 2", len(exs))
	}
	if exs[0].ID != "01_all_aboard" || exs[1].ID != "02_round_robin" {
		t.Fatalf("order wrong: %s, %s", exs[0].ID, exs[1].ID)
	}
	if exs[0].Type != TypeScenario || exs[1].Type != TypeFix {
		t.Fatalf("types wrong: %s, %s", exs[0].Type, exs[1].Type)
	}
	if exs[1].File != "round_robin.conf" {
		t.Fatalf("file wrong: %s", exs[1].File)
	}
	if len(exs[1].Hints) != 2 {
		t.Fatalf("hints wrong: %d", len(exs[1].Hints))
	}
	if exs[1].Desc == "" {
		t.Fatal("desc not parsed")
	}
}

func TestLoadRejectsDuplicateIDs(t *testing.T) {
	yaml := `
- id: same
  title: a
  type: scenario
  check: a/check.sh
- id: same
  title: b
  type: scenario
  check: b/check.sh
`
	if _, err := Load(writeTemp(t, yaml)); err == nil {
		t.Fatal("want error for duplicate ids")
	}
}

func TestLoadRejectsFixWithoutFile(t *testing.T) {
	yaml := `
- id: broken
  title: no file
  type: fix
  check: broken/check.sh
`
	if _, err := Load(writeTemp(t, yaml)); err == nil {
		t.Fatal("want error for fix exercise without file")
	}
}

func TestLoadRejectsUnknownType(t *testing.T) {
	yaml := `
- id: weird
  title: weird
  type: quiz
  check: weird/check.sh
`
	if _, err := Load(writeTemp(t, yaml)); err == nil {
		t.Fatal("want error for unknown type")
	}
}

func TestLoadRejectsMissingCheck(t *testing.T) {
	yaml := `
- id: nocheck
  title: nocheck
  type: scenario
`
	if _, err := Load(writeTemp(t, yaml)); err == nil {
		t.Fatal("want error for missing check")
	}
}
