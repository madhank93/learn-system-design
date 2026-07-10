// Package stack drives the topic's Docker Compose stack with the exercise
// override mounted, so learners edit work-dir files instead of the pristine
// repo confs. Thin exec wrapper; correctness is proven by `syslings verify`
// running the real stack end-to-end.
package stack

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Stack struct {
	// TopicDir is the compose project dir (has docker-compose.yml).
	TopicDir string
	// OverridePath is the exercise compose override (mounts work-dir confs).
	OverridePath string
	// WorkDir is exported to compose as SYSLINGS_WORK for the override's mounts.
	WorkDir string
}

func (s *Stack) compose(ctx context.Context, args ...string) *exec.Cmd {
	base := []string{"compose", "-f", "docker-compose.yml", "-f", s.OverridePath}
	cmd := exec.CommandContext(ctx, "docker", append(base, args...)...)
	cmd.Dir = s.TopicDir
	cmd.Env = append(os.Environ(), "SYSLINGS_WORK="+s.WorkDir)
	return cmd
}

func (s *Stack) Up(ctx context.Context) error {
	cmd := s.compose(ctx, "up", "--build", "-d")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compose up: %w\n%s", err, out)
	}
	return nil
}

func (s *Stack) Down(ctx context.Context) error {
	cmd := s.compose(ctx, "down", "--remove-orphans")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compose down: %w\n%s", err, out)
	}
	return nil
}

// IsUp probes one of the stack's HTTP endpoints.
func (s *Stack) IsUp(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// WaitUp polls url until it answers or the deadline passes.
func (s *Stack) WaitUp(ctx context.Context, url string, deadline time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	for {
		if s.IsUp(url) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("stack not answering on %s after %s", url, deadline)
		case <-time.After(500 * time.Millisecond):
		}
	}
}
