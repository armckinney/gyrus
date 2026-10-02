package agent

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

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
