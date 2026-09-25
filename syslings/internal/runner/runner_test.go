package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func script(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "check.sh")
	if err := os.WriteFile(path, []byte("#!/usr/bin/env bash\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunPassingCheck(t *testing.T) {
	res := Run(context.Background(), script(t, `echo "  ✓ all good"; exit 0`), t.TempDir(), 5*time.Second)
	if !res.Passed {
		t.Fatalf("want pass, got fail: %s", res.Output)
	}
	if !strings.Contains(res.Output, "✓ all good") {
		t.Fatalf("output missing: %q", res.Output)
	}
}

func TestRunFailingCheckCapturesOutputAndHint(t *testing.T) {
	res := Run(context.Background(), script(t, `echo "  ✗ broke"; echo "  hint: fix it" >&2; exit 1`), t.TempDir(), 5*time.Second)
	if res.Passed {
		t.Fatal("want fail")
	}
	if !strings.Contains(res.Output, "✗ broke") || !strings.Contains(res.Output, "hint: fix it") {
		t.Fatalf("stdout+stderr not merged: %q", res.Output)
	}
}

func TestRunTimesOut(t *testing.T) {
	res := Run(context.Background(), script(t, `sleep 10`), t.TempDir(), 300*time.Millisecond)
	if res.Passed {
		t.Fatal("want fail on timeout")
	}
	if !res.TimedOut {
		t.Fatal("want TimedOut flag")
	}
}

func TestRunSetsWorkDir(t *testing.T) {
	dir := t.TempDir()
	res := Run(context.Background(), script(t, `pwd`), dir, 5*time.Second)
	if !strings.Contains(res.Output, filepath.Base(dir)) {
		t.Fatalf("check did not run in work dir: %q", res.Output)
	}
}
