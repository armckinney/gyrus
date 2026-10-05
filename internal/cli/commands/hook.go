package commands

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

// HookInputPayload represents the common input payload sent via stdin by agent lifecycle hooks.
type HookInputPayload struct {
	ConversationID        string   `json:"conversationId"`
	WorkspacePaths        []string `json:"workspacePaths"`
	TranscriptPath        string   `json:"transcriptPath"`
	ArtifactDirectoryPath string   `json:"artifactDirectoryPath"`
	ModelName             string   `json:"modelName"`
	ExecutionNum          int      `json:"executionNum"`
	TerminationReason     string   `json:"terminationReason"`
	StepIdx               int      `json:"stepIdx"`
}

// PreInvocationOutput represents the response payload for PreInvocation hooks.
type PreInvocationOutput struct {
	InjectSteps []InjectStep `json:"injectSteps,omitempty"`
}

// InjectStep represents a step injected into the agent conversation loop.
type InjectStep struct {
	EphemeralMessage string `json:"ephemeralMessage,omitempty"`
}

// StopOutput represents the response payload for Stop hooks.
type StopOutput struct {
	Decision string `json:"decision,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// StopHookState represents cached state to prevent prompt loops.
type StopHookState struct {
	LastPromptedHash   string `json:"last_prompted_hash"`
	LastConversationID string `json:"last_conversation_id"`
}

// NewHookCmd constructs the 'hook' Cobra command group for agent lifecycle hooks.
func NewHookCmd(application *app.App) *cobra.Command {
	hookCmd := &cobra.Command{
		Use:   "hook",
		Short: "Agent lifecycle hook handlers (pre-invocation, stop)",
		Long:  "Commands invoked by agent runtime lifecycle hooks to inject context and perform architectural checks.",
	}

	hookCmd.AddCommand(newPreInvocationHookCmd(application))
	hookCmd.AddCommand(newStopHookCmd(application))

	return hookCmd
}

func newPreInvocationHookCmd(application *app.App) *cobra.Command {
	var explicitPrompt string

	cmd := &cobra.Command{
		Use:   "pre-invocation",
		Short: "Handle PreInvocation hook event and inject high-level context",
		RunE: func(cmd *cobra.Command, args []string) error {
			if os.Getenv("GYRUS_HOOKS_ENABLED") == "false" {
				fmt.Println("{}")
				return nil
			}

			payload := readHookPayload()
			_ = payload
			cfg := application.Config()
			if cfg != nil && (!cfg.HooksEnabled() || !cfg.AutoContext()) {
				fmt.Println("{}")
				return nil
			}

			engine, err := application.Engine()
			if err != nil {
				fmt.Println("{}")
				return nil
			}

			promptToSearch := explicitPrompt
			if promptToSearch == "" && len(args) > 0 {
				promptToSearch = strings.Join(args, " ")
			}
			if promptToSearch == "" {
				promptToSearch = "architecture specifications active requirements adrs"
			}

			filter := gyrus.SearchFilter{}
			contextLayer, err := engine.SuggestContextWithFilter(context.Background(), promptToSearch, filter, 5)
			if err != nil {
				fmt.Println("{}")
				return nil
			}

			if strings.TrimSpace(contextLayer) == "" || strings.TrimSpace(contextLayer) == "=== SUGGESTED CONTEXT LAYER ===" {
				fmt.Println("{}")
				return nil
			}

			const sdlcDirectives = `[Gyrus Architectural Context & SDLC Directives]
1. Source Code owns implementation details. Gyrus owns high-level architectural intent, requirements, contracts, and decisions. Never duplicate routine code edits or commit logs in Gyrus.
2. Gyrus SDLC Lifecycle Phases (from standards-001):
   - Plan: PRD (prd) -> Improvement Proposal (improvement-proposal) -> Decisions (adr) / Blueprints (specification) / Rules (standards)
   - Implement: Source code implements the specs. Update Technical Reference (technical-reference) & Guides (guide) when public interfaces, CLI/configs, or setup workflows change.
   - Post-Implement: Release Notes (release-note) & Product Pages (product) for version releases and portal overviews.
3. Living Documents (specification, standards, guide, technical-reference): Keep updated when system boundaries or public interfaces evolve.
4. Decision Logs (adr): Immutable once accepted. If an architectural decision changes, propose a new ADR superseding the previous one.

`
			fullMessage := sdlcDirectives + contextLayer
			out := PreInvocationOutput{
				InjectSteps: []InjectStep{
					{
						EphemeralMessage: fullMessage,
					},
				},
			}

			bytes, err := json.Marshal(out)
			if err != nil {
				fmt.Println("{}")
				return nil
			}
			fmt.Println(string(bytes))
			return nil
		},
	}

	cmd.Flags().StringVar(&explicitPrompt, "prompt", "", "Explicit prompt to suggest context for (optional)")
	return cmd
}

func newStopHookCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Handle Stop hook event and perform architectural reflection check",
		RunE: func(cmd *cobra.Command, args []string) error {
			if os.Getenv("GYRUS_HOOKS_ENABLED") == "false" {
				fmt.Println("{}")
				return nil
			}

			payload := readHookPayload()
			cfg := application.Config()
			if cfg != nil && (!cfg.HooksEnabled() || !cfg.ArchitecturalCheck()) {
				fmt.Println("{}")
				return nil
			}

			// Check git status
			changedFiles, err := getGitChangedFiles()
			if err != nil || len(changedFiles) == 0 {
				fmt.Println("{}")
				return nil
			}

			// Did any files in .gyrus/docs/ change?
			docChanged := false
			for _, file := range changedFiles {
				if strings.Contains(file, ".gyrus/docs/") || strings.Contains(file, "docs/") {
					docChanged = true
					break
				}
			}
			if docChanged {
				fmt.Println("{}")
				return nil
			}

			// Filter ignored paths
			ignored := cfg.IgnoredPaths()
			var meaningfulFiles []string
			for _, file := range changedFiles {
				if !isPathIgnored(file, ignored) {
					meaningfulFiles = append(meaningfulFiles, file)
				}
			}

			if len(meaningfulFiles) == 0 {
				fmt.Println("{}")
				return nil
			}

			// Compute hash of modified files to prevent infinite loops (single-prompt escape hatch)
			hashInput := strings.Join(meaningfulFiles, "|")
			hasher := sha256.New()
			hasher.Write([]byte(hashInput))
			currentHash := hex.EncodeToString(hasher.Sum(nil))

			statePath := filepath.Join(".gyrus", "cache", "stop_hook_state.json")
			var lastState StopHookState
			if data, err := os.ReadFile(statePath); err == nil {
				_ = json.Unmarshal(data, &lastState)
			}

			// If we already prompted for this exact set of changed files, allow stop cleanly
			if lastState.LastPromptedHash == currentHash {
				fmt.Println("{}")
				return nil
			}

			// Record prompt state
			newState := StopHookState{
				LastPromptedHash:   currentHash,
				LastConversationID: payload.ConversationID,
			}
			_ = os.MkdirAll(filepath.Dir(statePath), 0755)
			if stateBytes, err := json.Marshal(newState); err == nil {
				_ = os.WriteFile(statePath, stateBytes, 0644)
			}

			out := StopOutput{
				Decision: "continue",
				Reason:   "Gyrus SDLC Architectural Check: If this task introduced a new architectural decision, updated public API/config contracts, or changed system boundaries, consider updating the corresponding living spec (specification), technical reference (technical-reference), or recording an ADR (adr). If this was an internal implementation-only change, you may conclude without documentation updates.",
			}

			bytes, err := json.Marshal(out)
			if err != nil {
				fmt.Println("{}")
				return nil
			}
			fmt.Println(string(bytes))
			return nil
		},
	}
}

func readHookPayload() HookInputPayload {
	var payload HookInputPayload
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err == nil && len(data) > 0 {
			_ = json.Unmarshal(data, &payload)
		}
	}
	return payload
}

func getGitChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 3 {
			continue
		}
		// Git status format: XY <path> or XY <path1> -> <path2>
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			files = append(files, parts[len(parts)-1])
		}
	}
	return files, nil
}

func isPathIgnored(path string, ignoredPatterns []string) bool {
	cleanPath := filepath.ToSlash(path)

	// Standard hidden/meta directories
	if strings.HasPrefix(cleanPath, ".") || strings.HasPrefix(cleanPath, ".git/") || strings.HasPrefix(cleanPath, ".agents/") {
		return true
	}

	for _, pattern := range ignoredPatterns {
		pattern = filepath.ToSlash(pattern)
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if strings.HasPrefix(cleanPath, prefix+"/") || cleanPath == prefix {
				return true
			}
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(cleanPath)); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, cleanPath); matched {
			return true
		}
	}
	return false
}
