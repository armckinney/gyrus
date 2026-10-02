package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

// AntigravityRunner runs the Antigravity agent CLI ('agy').
type AntigravityRunner struct {
	workspaceDir   string
	binaryPath     string
	available      bool
	mu             sync.Mutex
	sessionTurns   map[string]int
	sessionConvIDs map[string]string
}

// NewAntigravityRunner constructs an AntigravityRunner.
func NewAntigravityRunner(workspaceDir string) *AntigravityRunner {
	binPath, found := findAgyPath()
	return &AntigravityRunner{
		workspaceDir:   workspaceDir,
		binaryPath:     binPath,
		available:      found,
		sessionTurns:   make(map[string]int),
		sessionConvIDs: make(map[string]string),
	}
}

func (r *AntigravityRunner) Name() string {
	return "Antigravity"
}

func (r *AntigravityRunner) ClientType() string {
	return "antigravity"
}

func (r *AntigravityRunner) IsAvailable() bool {
	return r.available
}

func (r *AntigravityRunner) BinaryPath() string {
	return r.binaryPath
}

func (r *AntigravityRunner) StreamTurn(ctx context.Context, opts TurnOptions, out chan<- StreamEvent) error {
	if !r.available {
		return fmt.Errorf("Antigravity CLI ('agy') not found in PATH or standard installation directories")
	}

	r.mu.Lock()
	turnCount := r.sessionTurns[opts.SessionID]
	r.sessionTurns[opts.SessionID] = turnCount + 1
	convID := r.sessionConvIDs[opts.SessionID]
	r.mu.Unlock()

	args := []string{}
	if convID != "" {
		args = append(args, "--conversation", convID)
	} else if turnCount > 0 {
		args = append(args, "-c")
	}

	if opts.Model != "" && opts.Model != "default" {
		args = append(args, "--model", opts.Model)

		// Respect model-specific effort requirements in agy CLI:
		switch opts.Model {
		case "gemini-3.8-flash", "gemini-3.7-flash", "gemini-3.6-flash":
			eff := strings.ToLower(opts.Effort)
			if eff != "low" && eff != "medium" && eff != "high" {
				eff = "high"
			}
			args = append(args, "--effort", eff)
		case "gemini-3.1-pro":
			eff := strings.ToLower(opts.Effort)
			if eff != "low" && eff != "high" {
				eff = "high"
			}
			args = append(args, "--effort", eff)
		case "gpt-oss-120b":
			args = append(args, "--effort", "medium")
		case "claude-sonnet-4-6", "claude-opus-4-6-thinking":
			// agy does not accept --effort for Claude models; omit --effort
		default:
			if opts.Effort != "" && opts.Effort != "default" {
				args = append(args, "--effort", opts.Effort)
			}
		}
	} else if opts.Effort != "" && opts.Effort != "default" {
		args = append(args, "--effort", opts.Effort)
	}

	args = append(args, "-p", opts.Prompt, "--output-format", "stream-json", "--dangerously-skip-permissions")

	cmd := exec.CommandContext(ctx, r.binaryPath, args...)
	if r.workspaceDir != "" {
		cmd.Dir = r.workspaceDir
	}

	return runStreamingJSONCmd(cmd, out, func(cID string) {
		if cID != "" {
			r.mu.Lock()
			r.sessionConvIDs[opts.SessionID] = cID
			r.mu.Unlock()
		}
	})
}

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

	args := []string{"-p", opts.Prompt, "--dangerously-skip-permissions"}
	if opts.Model != "" && opts.Model != "default" {
		args = append(args, "--model", opts.Model)
	}

	cmd := exec.CommandContext(ctx, r.binaryPath, args...)
	if r.workspaceDir != "" {
		cmd.Dir = r.workspaceDir
	}

	return runStreamingCmd(cmd, out)
}

// GenericRunner executes arbitrary or generic CLI binaries.
type GenericRunner struct {
	name         string
	clientType   string
	command      string
	workspaceDir string
	binaryPath   string
	available    bool
}

// NewGenericRunner creates a generic command runner.
func NewGenericRunner(name, clientType, command, workspaceDir string) *GenericRunner {
	parts := strings.Fields(command)
	bin := command
	if len(parts) > 0 {
		bin = parts[0]
	}
	binPath, err := exec.LookPath(bin)
	return &GenericRunner{
		name:         name,
		clientType:   clientType,
		command:      command,
		workspaceDir: workspaceDir,
		binaryPath:   binPath,
		available:    err == nil,
	}
}

func (r *GenericRunner) Name() string {
	return r.name
}

func (r *GenericRunner) ClientType() string {
	return r.clientType
}

func (r *GenericRunner) IsAvailable() bool {
	return r.available
}

func (r *GenericRunner) BinaryPath() string {
	return r.binaryPath
}

func (r *GenericRunner) StreamTurn(ctx context.Context, opts TurnOptions, out chan<- StreamEvent) error {
	if !r.available {
		return fmt.Errorf("CLI command '%s' not found in PATH", r.command)
	}

	parts := strings.Fields(r.command)
	args := parts[1:]
	if opts.Model != "" && opts.Model != "default" {
		args = append(args, "--model", opts.Model)
	}
	args = append(args, "-p", opts.Prompt)

	cmd := exec.CommandContext(ctx, parts[0], args...)
	if r.workspaceDir != "" {
		cmd.Dir = r.workspaceDir
	}

	return runStreamingCmd(cmd, out)
}

type agyStreamEvent struct {
	Event          string `json:"event"`
	ConversationID string `json:"conversation_id"`
	StepUpdate     *struct {
		StepIndex       int     `json:"step_index"`
		StepType        string  `json:"step_type"`
		State           string  `json:"state"`
		TextDelta       string  `json:"text_delta"`
		ToolName        string  `json:"tool_name"`
		DurationSeconds float64 `json:"duration_seconds"`
		ToolInfo        *struct {
			Name       string                 `json:"name"`
			Parameters map[string]interface{} `json:"parameters"`
			Output     interface{}            `json:"output"`
		} `json:"tool_info"`
	} `json:"step_update"`
	Result *struct {
		Status   string `json:"status"`
		Response string `json:"response"`
	} `json:"result"`
}

// runStreamingJSONCmd starts a subprocess with NDJSON stream-json output format, parsing text deltas and tool events.
func runStreamingJSONCmd(cmd *exec.Cmd, out chan<- StreamEvent, onInit func(string)) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent process: %w", err)
	}

	reader := bufio.NewReaderSize(stdout, 128*1024)
	hasStreamedToken := false

	for {
		line, readErr := reader.ReadBytes('\n')
		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			var ev agyStreamEvent
			if err := json.Unmarshal(line, &ev); err == nil {
				if ev.Event == "init" && ev.ConversationID != "" && onInit != nil {
					onInit(ev.ConversationID)
				}

				if ev.StepUpdate != nil {
					if ev.StepUpdate.TextDelta != "" {
						hasStreamedToken = true
						out <- StreamEvent{
							Type: "token",
							Data: ev.StepUpdate.TextDelta,
						}
					} else if ev.StepUpdate.StepType == "tool" || ev.StepUpdate.ToolName != "" {
						toolName := ev.StepUpdate.ToolName
						if toolName == "" && ev.StepUpdate.ToolInfo != nil {
							toolName = ev.StepUpdate.ToolInfo.Name
						}

						summary := ""
						cmdStr := ""
						if ev.StepUpdate.ToolInfo != nil && ev.StepUpdate.ToolInfo.Parameters != nil {
							params := ev.StepUpdate.ToolInfo.Parameters
							if c, ok := params["CommandLine"].(string); ok && c != "" {
								summary = c
								cmdStr = c
							} else if path, ok := params["AbsolutePath"].(string); ok && path != "" {
								summary = path
							} else if tn, ok := params["ToolName"].(string); ok && tn != "" {
								summary = fmt.Sprintf("MCP: %s", tn)
							} else if q, ok := params["query"].(string); ok && q != "" {
								summary = q
							} else if p, ok := params["Prompt"].(string); ok && p != "" {
								summary = p
							} else {
								for _, v := range params {
									if s, ok := v.(string); ok && s != "" {
										summary = s
										break
									}
								}
							}
						}

						duration := ""
						if ev.StepUpdate.DurationSeconds > 0 {
							duration = fmt.Sprintf("%.2fs", ev.StepUpdate.DurationSeconds)
						}

						outputStr := ""
						if ev.StepUpdate.ToolInfo != nil && ev.StepUpdate.ToolInfo.Output != nil {
							switch v := ev.StepUpdate.ToolInfo.Output.(type) {
							case string:
								outputStr = strings.TrimSpace(v)
							default:
								b, _ := json.Marshal(v)
								outputStr = string(b)
							}
							if len(outputStr) > 500 {
								outputStr = outputStr[:500] + "..."
							}
						}

						action := &ActionItem{
							ID:        fmt.Sprintf("action-%d", ev.StepUpdate.StepIndex),
							StepIndex: ev.StepUpdate.StepIndex,
							ToolName:  toolName,
							Summary:   summary,
							Command:   cmdStr,
							Output:    outputStr,
							Duration:  duration,
							State:     ev.StepUpdate.State,
						}

						out <- StreamEvent{
							Type:   "action",
							Data:   fmt.Sprintf("%s: %s", toolName, summary),
							Action: action,
						}

						if ev.StepUpdate.State == "ACTIVE" {
							out <- StreamEvent{
								Type: "status",
								Data: fmt.Sprintf("Running: %s...", toolName),
							}
						} else if ev.StepUpdate.State == "DONE" {
							out <- StreamEvent{
								Type: "status",
								Data: "Thinking...",
							}
						}
					}
				} else if ev.Event == "result" && ev.Result != nil {
					if !hasStreamedToken && ev.Result.Response != "" {
						out <- StreamEvent{
							Type: "token",
							Data: ev.Result.Response,
						}
					}
				}
			} else {
				// Fallback: line is plain text
				out <- StreamEvent{
					Type: "token",
					Data: string(line) + "\n",
				}
			}
		}

		if readErr != nil {
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

// findAgyPath searches PATH and common installation directories for 'agy'.
func findAgyPath() (string, bool) {
	if p, err := exec.LookPath("agy"); err == nil {
		return p, true
	}
	userHome, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(userHome, ".local", "bin", "agy"),
		filepath.Join(userHome, ".gemini", "bin", "agy"),
		"/usr/local/bin/agy",
		"/root/.local/bin/agy",
		"/root/.gemini/bin/agy",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, true
		}
	}
	return "", false
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
