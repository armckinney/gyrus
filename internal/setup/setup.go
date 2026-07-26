package setup

import (
	"path/filepath"
)

// SetupOptions defines options for workspace initialization and tool tailoring.
type SetupOptions struct {
	WorkspaceDir string
	Profile      Profile
	OwnerGroup   string
	MCPTarget    MCPTarget
	SkillTarget  SkillTarget
	BinaryCmd    string
	SkipMCP      bool
	SkipSkill    bool
	SkipConfig   bool
}

// SetupResult contains execution status details from workspace setup.
type SetupResult struct {
	ConfigFile   string
	StorageDir   string
	InstalledMCP []string
	SkillFiles   []string
}

// RunSetup executes the workspace initialization workflow.
func RunSetup(opts SetupOptions) (*SetupResult, error) {
	if opts.WorkspaceDir == "" {
		opts.WorkspaceDir = "."
	}
	if opts.Profile == "" {
		opts.Profile = ProfileLocal
	}
	if opts.OwnerGroup == "" {
		opts.OwnerGroup = "armckinney"
	}
	if opts.MCPTarget == "" {
		opts.MCPTarget = MCPTargetAll
	}
	if opts.SkillTarget == "" {
		opts.SkillTarget = SkillTargetAll
	}
	if opts.BinaryCmd == "" {
		opts.BinaryCmd = "gyrus"
	}

	result := &SetupResult{}

	// 1. Resolve storage root directory path
	storageDir := filepath.Join(opts.WorkspaceDir, "docs", ".gyrus", "docs")
	result.StorageDir = storageDir

	// 2. Generate .gyrus.yaml config file (if not skipped)
	if !opts.SkipConfig {
		cfgPath, err := WriteConfigFile(opts.WorkspaceDir, opts.Profile, opts.OwnerGroup)
		if err != nil {
			return nil, err
		}
		result.ConfigFile = cfgPath
	}

	// 3. Equip Agent Skill files (if not skipped)
	if !opts.SkipSkill {
		skillFiles, err := InstallAgentSkill(opts.WorkspaceDir, opts.SkillTarget)
		if err != nil {
			return nil, err
		}
		result.SkillFiles = skillFiles
	}

	// 4. Register MCP Server configs (if not skipped)
	if !opts.SkipMCP {
		mcpFiles, err := RegisterMCPServer(opts.WorkspaceDir, opts.MCPTarget, opts.BinaryCmd)
		if err != nil {
			return nil, err
		}
		result.InstalledMCP = mcpFiles
	}

	return result, nil
}
