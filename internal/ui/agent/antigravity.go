package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

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
