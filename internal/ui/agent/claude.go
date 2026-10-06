package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// ClaudeRunner runs the Claude Code CLI ('claude').
type ClaudeRunner struct {
	workspaceDir string
	binaryPath   string
	available    bool
	mu           sync.Mutex
	sessionTurns map[string]int
}

// NewClaudeRunner constructs a ClaudeRunner.
func NewClaudeRunner(workspaceDir string) *ClaudeRunner {
	binPath, found := findClaudePath()
	return &ClaudeRunner{
		workspaceDir: workspaceDir,
		binaryPath:   binPath,
		available:    found,
		sessionTurns: make(map[string]int),
	}
}

func (r *ClaudeRunner) Name() string {
	return "Claude Code"
}

func (r *ClaudeRunner) ClientType() string {
	return "claude"
}

func (r *ClaudeRunner) IsAvailable() bool {
	return r.available
}

func (r *ClaudeRunner) BinaryPath() string {
	return r.binaryPath
}

func (r *ClaudeRunner) StreamTurn(ctx context.Context, opts TurnOptions, out chan<- StreamEvent) error {
	if !r.available {
		return fmt.Errorf("Claude CLI ('claude') not found in PATH or standard installation directories")
	}

	r.mu.Lock()
	turnCount := r.sessionTurns[opts.SessionID]
	r.sessionTurns[opts.SessionID] = turnCount + 1
	r.mu.Unlock()

	args := []string{"-p", opts.Prompt}
	if opts.Model != "" && opts.Model != "default" {
		args = append(args, "--model", opts.Model)
	}

	cmd := exec.CommandContext(ctx, r.binaryPath, args...)
	if r.workspaceDir != "" {
		cmd.Dir = r.workspaceDir
	}

	return runStreamingCmd(cmd, out)
}

// findClaudePath searches PATH and standard installation paths for 'claude'.
func findClaudePath() (string, bool) {
	if p, err := exec.LookPath("claude"); err == nil {
		return p, true
	}
	userHome, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(userHome, ".local", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(userHome, ".npm-global", "bin", "claude"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, true
		}
	}
	return "", false
}
