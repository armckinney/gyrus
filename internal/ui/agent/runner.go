package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// StreamEvent represents a single event chunk from the agent.
type StreamEvent struct {
	Type   string      `json:"type"` // "token", "status", "action", "error", "done"
	Data   string      `json:"data"`
	Action *ActionItem `json:"action,omitempty"`
}

// TurnOptions configures a single conversational turn with an agent.
type TurnOptions struct {
	SessionID string `json:"session_id"`
	Prompt    string `json:"prompt"`
	Model     string `json:"model,omitempty"`
	Effort    string `json:"effort,omitempty"`
}

// Runner defines the interface for an embedded agent harness.
type Runner interface {
	Name() string
	ClientType() string
	IsAvailable() bool
	BinaryPath() string
	StreamTurn(ctx context.Context, opts TurnOptions, out chan<- StreamEvent) error
}

// NewRunner returns the appropriate Runner implementation for the target client.
func NewRunner(clientType string, customCmd string, workspaceDir string) Runner {
	clientType = strings.TrimSpace(strings.ToLower(clientType))
	if customCmd != "" {
		return NewGenericRunner("Custom Command", clientType, customCmd, workspaceDir)
	}

	switch clientType {
	case "antigravity", "agy":
		return NewAntigravityRunner(workspaceDir)
	case "claude", "claude-code":
		return NewClaudeRunner(workspaceDir)
	case "codex":
		return NewGenericRunner("Codex CLI", "codex", "codex", workspaceDir)
	case "copilot":
		return NewGenericRunner("GitHub Copilot CLI", "copilot", "github-copilot-cli", workspaceDir)
	default:
		return NewGenericRunner(clientType, clientType, clientType, workspaceDir)
	}
}

// runStreamingCmd starts a subprocess, streaming its stdout token-by-token or chunk-by-chunk to out.
func runStreamingCmd(cmd *exec.Cmd, out chan<- StreamEvent) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent process: %w", err)
	}

	buf := make([]byte, 512)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			out <- StreamEvent{
				Type: "token",
				Data: string(buf[:n]),
			}
		}
		if err != nil {
			if err != io.EOF && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
				// Non-EOF read error
				out <- StreamEvent{
					Type: "error",
					Data: err.Error(),
				}
			}
			break
		}
	}

	waitErr := cmd.Wait()
	if waitErr != nil && cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
		errMsg := strings.TrimSpace(stderrBuf.String())
		if errMsg == "" {
			errMsg = waitErr.Error()
		}
		out <- StreamEvent{
			Type: "error",
			Data: errMsg,
		}
		return waitErr
	}

	out <- StreamEvent{
		Type: "done",
		Data: "[DONE]",
	}
	return nil
}
