// Package runner executes an exercise's check script and reports the verdict.
package runner

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

type Result struct {
	Passed   bool
	TimedOut bool
	Output   string // stdout + stderr, interleaved
}

// Run executes the check script with the work directory as cwd. The check's
// exit code is the verdict; stdout/stderr carry the ✓/✗ lines and hints.
func Run(ctx context.Context, checkPath, workDir string, timeout time.Duration) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", checkPath)
	cmd.Dir = workDir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	res := Result{Output: buf.String()}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		res.TimedOut = true
		res.Output += "\n  ✗ check timed out"
		return res
	}
	res.Passed = err == nil
	return res
}
