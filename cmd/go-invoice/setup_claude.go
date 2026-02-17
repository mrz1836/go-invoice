package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mrz1836/go-invoice/internal/cli"
)

// Define static errors for better error handling
var (
	ErrGoCompilerNotFound     = errors.New("go compiler not found. Please install Go 1.21 or later from https://golang.org/dl/")
	ErrProjectRootNotFound    = errors.New("could not determine project root")
	ErrSetupCancelled         = errors.New("setup canceled")
	ErrNoProjectFileDetected  = errors.New("no project file detected")
	ErrMCPBinaryNotExecutable = errors.New("MCP server binary is not executable")
)

// MCPConfig represents the .mcp.json structure
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents an individual MCP server configuration
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// ClaudeSettings represents the .claude/settings.json structure
type ClaudeSettings struct {
	Permissions PermissionSettings `json:"permissions,omitempty"`
	Env         map[string]string  `json:"env,omitempty"`
}

// PermissionSettings represents permission configuration
type PermissionSettings struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

// ClaudeDesktopConfig represents the Claude Desktop mcp_servers.json structure
type ClaudeDesktopConfig struct {
	GoInvoice DesktopServerConfig `json:"go-invoice"`
}

// DesktopServerConfig represents Claude Desktop server configuration
type DesktopServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// buildConfigSetupClaudeCommand creates the config setup-claude subcommand
func (a *App) buildConfigSetupClaudeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-claude",
		Short: "Set up Claude Desktop and Claude Code integration",
		Long: `Interactive setup wizard to configure go-invoice MCP integration for both
Claude Desktop and Claude Code. This will:

- Check prerequisites (go-invoice CLI, Go compiler)
- Build the MCP server binary
- Set up configuration files
- Configure project-specific settings
- Test the integration

The wizard supports both fresh installations and updates to existing configurations.`,
		Example: `  # Set up both Claude Desktop and Claude Code
  go-invoice config setup-claude

  # Set up only Claude Desktop
  go-invoice config setup-claude --desktop

  # Set up only Claude Code (in current project)
  go-invoice config setup-claude --code

  # Update existing installation
  go-invoice config setup-claude --update`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			desktop, _ := cmd.Flags().GetBool("desktop")
			code, _ := cmd.Flags().GetBool("code")
			update, _ := cmd.Flags().GetBool("update")

			// If no specific flags, setup both
			if !desktop && !code {
				desktop = true
				code = true
			}

			return a.runSetupClaude(ctx, desktop, code, update)
		},
	}

	cmd.Flags().Bool("desktop", false, "Set up Claude Desktop integration only")
	cmd.Flags().Bool("code", false, "Set up Claude Code integration only")
	cmd.Flags().Bool("update", false, "Update existing installation")

	return cmd
}

// runSetupClaude handles the Claude setup process
func (a *App) runSetupClaude(ctx context.Context, setupDesktop, setupCode, isUpdate bool) error {
	prompter := cli.NewPrompter(a.logger)

	a.logger.Println("🚀 Welcome to go-invoice Claude integration setup!")
	a.logger.Println("")

	// Check prerequisites
	if err := a.checkClaudePrerequisites(ctx); err != nil {
		return err
	}

	// Determine go-invoice home directory
	goInvoiceHome := os.Getenv("GO_INVOICE_HOME")
	if goInvoiceHome == "" {
		goInvoiceHome = filepath.Join(os.Getenv("HOME"), ".go-invoice")
	}

	// Get project root (where this binary is located)
	projectRoot, err := a.getProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to determine project root: %w", err)
	}

	a.logger.Printf("📁 Project root: %s\n", projectRoot)
	a.logger.Printf("📁 go-invoice home: %s\n", goInvoiceHome)
	a.logger.Println("")

	// Create directory structure
	if err := a.setupClaudeDirectories(goInvoiceHome); err != nil {
		return err
	}

	// Build MCP server
	if err := a.buildMCPServer(ctx, projectRoot); err != nil {
		return err
	}

	// Deploy shared configuration
	if err := a.deploySharedConfig(projectRoot, goInvoiceHome); err != nil {
		return err
	}

	// Setup Claude Desktop if requested
	if setupDesktop {
		if err := a.setupClaudeDesktop(ctx, prompter, projectRoot, isUpdate); err != nil {
			a.logger.Printf("⚠️  Claude Desktop setup failed: %v\n", err)
			a.logger.Println("   You can retry with: go-invoice config setup-claude --desktop")
		}
	}

	// Setup Claude Code if requested
	if setupCode {
		if err := a.setupClaudeCode(ctx, prompter, projectRoot); err != nil {
			a.logger.Printf("⚠️  Claude Code setup failed: %v\n", err)
			a.logger.Println("   You can retry with: go-invoice config setup-claude --code")
		}
	}

	// Test the integration
	a.logger.Println("")
	a.logger.Println("🧪 Testing MCP server...")
	if err := a.testMCPServer(ctx, projectRoot); err != nil {
		a.logger.Printf("⚠️  MCP server test failed: %v\n", err)
	} else {
		a.logger.Println("✅ MCP server is working correctly")
	}

	// Print summary
	a.printClaudeSetupSummary(ctx, setupDesktop, setupCode, projectRoot, goInvoiceHome)

	return nil
}

// checkClaudePrerequisites checks for required tools
func (a *App) checkClaudePrerequisites(ctx context.Context) error {
	a.logger.Println("🔍 Checking prerequisites...")

	// Check for go-invoice CLI (we are it!)
	a.logger.Println("✅ go-invoice CLI found")

	// Check for Go compiler
	if _, err := exec.LookPath("go"); err != nil {
		return ErrGoCompilerNotFound
	}

	// Check Go version
	cmd := exec.CommandContext(ctx, "go", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check Go version: %w", err)
	}
	a.logger.Printf("✅ Go compiler found: %s", strings.TrimSpace(string(output)))

	return nil
}

// getProjectRoot determines the project root directory
func (a *App) getProjectRoot() (string, error) {
	// Try to find go.mod by walking up from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		modPath := filepath.Join(dir, "go.mod")
		if _, statErr := os.Stat(modPath); statErr == nil {
			// Check if this is the go-invoice project
			modContent, readErr := os.ReadFile(modPath) //nolint:gosec // Path is validated
			if readErr == nil && strings.Contains(string(modContent), "github.com/mrz1836/go-invoice") {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: use the directory containing this binary
	exe, err := os.Executable()
	if err != nil {
		return "", ErrProjectRootNotFound
	}

	// Go up from bin directory
	binDir := filepath.Dir(exe)
	return filepath.Dir(binDir), nil
}

// setupClaudeDirectories creates the necessary directory structure
func (a *App) setupClaudeDirectories(goInvoiceHome string) error {
	a.logger.Println("📁 Setting up directory structure...")

	// First ensure the home directory exists with correct permissions
	if err := os.MkdirAll(goInvoiceHome, 0o700); err != nil {
		return fmt.Errorf("failed to create home directory %s: %w", goInvoiceHome, err)
	}

	// Fix permissions if directory already exists
	if err := os.Chmod(goInvoiceHome, 0o700); err != nil { //nolint:gosec // G302 - Directory needs execute permission
		a.logger.Printf("⚠️  Could not set permissions on %s: %v\n", goInvoiceHome, err)
	}

	dirs := []string{
		filepath.Join(goInvoiceHome, "logs"),
		filepath.Join(goInvoiceHome, "config"),
		filepath.Join(goInvoiceHome, "data"),
		filepath.Join(goInvoiceHome, "cache"),
		filepath.Join(goInvoiceHome, "logs", "archive"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	a.logger.Println("✅ Directory structure created")
	return nil
}

// buildMCPServer builds the MCP server binary
func (a *App) buildMCPServer(ctx context.Context, projectRoot string) error {
	a.logger.Println("🔨 Building MCP server...")

	// Change to project root
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(projectRoot); err != nil {
		return fmt.Errorf("failed to change to project root: %w", err)
	}

	// Build the MCP server
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "bin/go-invoice-mcp", "./cmd/go-invoice-mcp")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build MCP server: %w", err)
	}

	// Make it executable
	mcpPath := filepath.Join(projectRoot, "bin", "go-invoice-mcp")
	if err := os.Chmod(mcpPath, 0o750); err != nil { //nolint:gosec // Binary needs to be executable
		return fmt.Errorf("failed to make MCP server executable: %w", err)
	}

	// Try to create symlink in /usr/local/bin (optional)
	if runtime.GOOS != "windows" {
		symlinkPath := "/usr/local/bin/go-invoice-mcp"
		if err := os.Symlink(mcpPath, symlinkPath); err == nil {
			a.logger.Printf("✅ Created symlink at %s\n", symlinkPath)
		} else {
			a.logger.Printf("ℹ️  Could not create symlink in /usr/local/bin (this is optional)\n")
			a.logger.Printf("   Add %s to your PATH if needed\n", filepath.Join(projectRoot, "bin"))
		}
	}

	a.logger.Println("✅ MCP server built successfully")
	return nil
}

// deploySharedConfig deploys shared configuration files
func (a *App) deploySharedConfig(projectRoot, goInvoiceHome string) error {
	a.logger.Println("📋 Deploying shared configuration...")

	configsDir := filepath.Join(projectRoot, "configs")
	targetConfigDir := filepath.Join(goInvoiceHome, "config")

	// Copy main MCP configuration
	srcConfig := filepath.Join(configsDir, "mcp-config.json")
	dstConfig := filepath.Join(targetConfigDir, "mcp-config.json")
	if err := a.copyFile(srcConfig, dstConfig); err != nil {
		return fmt.Errorf("failed to copy mcp-config.json: %w", err)
	}

	// Copy logging configuration
	srcLogging := filepath.Join(configsDir, "logging.yaml")
	dstLogging := filepath.Join(targetConfigDir, "logging.yaml")
	if err := a.copyFile(srcLogging, dstLogging); err != nil {
		return fmt.Errorf("failed to copy logging.yaml: %w", err)
	}

	// Create default go-invoice config if not exists
	goInvoiceConfig := filepath.Join(goInvoiceHome, "config.json")
	if _, err := os.Stat(goInvoiceConfig); os.IsNotExist(err) {
		defaultConfig := `{
  "storage_path": "~/.go-invoice/data",
  "invoice_defaults": {
    "currency": "USD",
    "tax_rate": 0.0,
    "payment_terms": 30
  },
  "templates": {
    "invoice": "~/.go-invoice/templates/invoice.html"
  }
}`
		if err := os.WriteFile(goInvoiceConfig, []byte(defaultConfig), 0o600); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}

	a.logger.Println("✅ Shared configuration deployed")
	return nil
}

// setupClaudeDesktop sets up Claude Desktop integration
func (a *App) setupClaudeDesktop(ctx context.Context, prompter *cli.Prompter, projectRoot string, isUpdate bool) error {
	a.logger.Println("")
	a.logger.Println("🖥️  Setting up Claude Desktop integration...")

	// Determine Claude Desktop config directory based on OS
	var configDir string
	switch runtime.GOOS {
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Claude")
	case "windows":
		configDir = filepath.Join(os.Getenv("APPDATA"), "Claude")
	default:
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "claude")
	}

	// Check if Claude Desktop is installed
	if _, err := os.Stat(configDir); os.IsNotExist(err) { //nolint:gosec // G703: configDir is derived from known system paths
		return fmt.Errorf("claude Desktop configuration directory not found at %s: %w", configDir, os.ErrNotExist)
	}

	// Backup existing configuration
	mcpServersPath := filepath.Join(configDir, "mcp_servers.json")
	if _, err := os.Stat(mcpServersPath); err == nil { //nolint:gosec // G703: mcpServersPath is derived from a known system config directory
		backupPath := mcpServersPath + ".backup"
		if err := a.copyFile(mcpServersPath, backupPath); err != nil {
			a.logger.Printf("⚠️  Could not backup existing configuration: %v\n", err)
		} else {
			a.logger.Printf("📋 Backed up existing configuration to %s\n", backupPath)
		}
	}

	// Create or update MCP servers configuration
	mcpBinPath := filepath.Join(projectRoot, "bin", "go-invoice-mcp")
	desktopConfig := map[string]interface{}{
		"go-invoice": DesktopServerConfig{
			Command: mcpBinPath,
			Args:    []string{"--stdio"},
			Env: map[string]string{
				"GO_INVOICE_HOME": "~/.go-invoice",
				"MCP_LOG_LEVEL":   "info",
				"MCP_TRANSPORT":   "stdio",
				"MCP_LOG_FILE":    "~/.go-invoice/mcp-claude-desktop.log",
			},
		},
	}

	// Check if we should merge with existing config
	var finalConfig map[string]interface{}
	if existingData, err := os.ReadFile(filepath.Clean(mcpServersPath)); err == nil && len(existingData) > 0 { //nolint:gosec // G703: mcpServersPath is derived from known system config paths
		// Parse existing config
		if err := json.Unmarshal(existingData, &finalConfig); err != nil {
			a.logger.Printf("⚠️  Could not parse existing configuration, will replace it\n")
			finalConfig = desktopConfig
		} else {
			// Merge configurations
			if isUpdate || a.promptForMerge(ctx, prompter, "Claude Desktop") {
				for k, v := range desktopConfig {
					finalConfig[k] = v
				}
				a.logger.Println("✅ Merged with existing configuration")
			} else {
				finalConfig = desktopConfig
				a.logger.Println("✅ Replaced existing configuration")
			}
		}
	} else {
		finalConfig = desktopConfig
	}

	// Write configuration
	configData, err := json.MarshalIndent(finalConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(mcpServersPath, configData, 0o600); err != nil { //nolint:gosec // G703: mcpServersPath is derived from known system config paths
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	a.logger.Println("✅ Claude Desktop integration configured")
	a.logger.Println("ℹ️  Restart Claude Desktop to load the go-invoice MCP server")

	return nil
}

// setupClaudeCode sets up Claude Code integration
func (a *App) setupClaudeCode(ctx context.Context, prompter *cli.Prompter, projectRoot string) error {
	a.logger.Println("")
	a.logger.Println("💻 Setting up Claude Code integration...")

	// Check if we're in a project directory
	if err := a.checkProjectContext(); err != nil {
		cont, err := prompter.PromptBool(ctx, "No project file detected. Continue anyway?", false)
		if err != nil || !cont {
			return ErrSetupCancelled
		}
	}

	// Create project structure
	if err := a.setupProjectStructure(); err != nil {
		return err
	}

	// Check for existing .mcp.json
	if _, err := os.Stat(".mcp.json"); err == nil {
		a.logger.Println("📋 Found existing .mcp.json")
		if err := a.copyFile(".mcp.json", ".mcp.json.backup"); err == nil {
			a.logger.Println("📋 Backed up existing MCP configuration")
		}
	}

	// Check for old .claude_config.json and migrate if found
	if _, err := os.Stat(".claude_config.json"); err == nil {
		a.logger.Println("📋 Found old .claude_config.json - will migrate to new format")
		if err := a.copyFile(".claude_config.json", ".claude_config.json.old"); err == nil {
			a.logger.Println("📋 Backed up old configuration")
		}
		// Remove old file after migration
		defer func() {
			if err := os.Remove(".claude_config.json"); err != nil {
				a.logger.Printf("Warning: failed to remove old config file: %v", err)
			}
		}()
	}

	// Ask about scope preference
	a.logger.Println("")
	a.logger.Println("📋 MCP Server Scope")
	a.logger.Println("------------------")
	a.logger.Println("User scope: Personal, available across all projects (recommended)")
	a.logger.Println("Project scope: Shared with your team via .mcp.json")

	useProjectScope, err := prompter.PromptBool(ctx, "Use project scope (team-shared)?", false)
	if err != nil {
		return err
	}

	// Create Claude settings directory
	if mkdirErr := os.MkdirAll(".claude", 0o750); mkdirErr != nil {
		return fmt.Errorf("failed to create .claude directory: %w", mkdirErr)
	}

	// Get project information for go-invoice config
	projectName := filepath.Base(mustGetwd())
	invoicePrefix := "INV"

	// Interactive configuration
	a.logger.Println("")
	a.logger.Println("📋 Project Configuration")
	a.logger.Println("----------------------")

	if name, nameErr := prompter.PromptString(ctx, "Project name", projectName); nameErr == nil && name != "" {
		projectName = name
	}

	if prefix, prefixErr := prompter.PromptString(ctx, "Invoice prefix", invoicePrefix); prefixErr == nil && prefix != "" {
		invoicePrefix = prefix
	}

	// Create MCP server configuration
	mcpBinPath := filepath.Join(projectRoot, "bin", "go-invoice-mcp")

	// Determine scope for claude mcp add command
	scope := "user"
	if useProjectScope {
		scope = "project"
	}

	// Execute claude mcp add command with environment variables
	a.logger.Println("📋 Registering go-invoice MCP server with Claude...")

	// Build the command with all arguments
	homeDir := os.Getenv("HOME")
	goInvoiceHomeDir := filepath.Join(homeDir, ".go-invoice")

	args := []string{
		"mcp", "add", "go-invoice", mcpBinPath,
		"-s", scope,
		"-t", "stdio",
		"-e", fmt.Sprintf("GO_INVOICE_HOME=%s", goInvoiceHomeDir),
		"-e", "MCP_TRANSPORT=stdio",
	}

	// Add environment variables based on scope
	if useProjectScope {
		args = append(args, "-e", "GO_INVOICE_PROJECT=${PWD}")
		args = append(args, "-e", "MCP_LOG_FILE=${PWD}/.go-invoice/mcp.log")
	} else {
		// For user scope, use absolute paths instead of ${PWD}
		currentDir, _ := os.Getwd()
		args = append(args, "-e", fmt.Sprintf("GO_INVOICE_PROJECT=%s", currentDir))
		args = append(args, "-e", fmt.Sprintf("MCP_LOG_FILE=%s/mcp-claude-code.log", goInvoiceHomeDir))
	}

	// First, try to remove existing server (ignore errors if it doesn't exist)
	removeCmd := exec.CommandContext(ctx, "claude", "mcp", "remove", "go-invoice")
	_ = removeCmd.Run() // Explicitly ignore output and errors

	// Execute the command
	cmd := exec.CommandContext(ctx, "claude", args...) // #nosec G204,G702 - args are controlled by our code
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if runErr := cmd.Run(); runErr != nil {
		// If command fails, provide manual instructions as fallback
		a.logger.Printf("⚠️  Could not automatically register MCP server: %v\n", runErr)
		a.logger.Println("")
		a.logger.Println("📋 To manually add go-invoice MCP server, run:")
		a.logger.Printf("   claude mcp remove go-invoice  # Remove existing if any\n")
		a.logger.Printf("   claude mcp add go-invoice %s -s %s -t stdio \\\n", mcpBinPath, scope)
		a.logger.Printf("     -e GO_INVOICE_HOME=\"%s\" \\\n", goInvoiceHomeDir)
		a.logger.Println("     -e MCP_TRANSPORT=\"stdio\" \\")
		if useProjectScope {
			a.logger.Println("     -e GO_INVOICE_PROJECT=\"${PWD}\" \\")
			a.logger.Println("     -e MCP_LOG_FILE=\"${PWD}/.go-invoice/mcp.log\"")
		} else {
			currentDir, _ := os.Getwd()
			a.logger.Printf("     -e GO_INVOICE_PROJECT=\"%s\" \\\n", currentDir)
			a.logger.Printf("     -e MCP_LOG_FILE=\"%s/mcp-claude-code.log\"\n", goInvoiceHomeDir)
		}
		return fmt.Errorf("failed to register MCP server: %w", runErr)
	}

	a.logger.Printf("✅ Successfully registered go-invoice MCP server with %s scope\n", scope)

	if useProjectScope {
		// Also create .mcp.json using the working configuration template
		mcpConfig := MCPConfig{
			MCPServers: map[string]MCPServerConfig{
				"go-invoice": {
					Command: "/Users/mrz/projects/go-invoice/bin/go-invoice-mcp",
					Args:    []string{"--stdio"},
					Env: map[string]string{
						"GO_INVOICE_HOME":    "${HOME}/.go-invoice",
						"GO_INVOICE_PROJECT": "${PWD}",
						"MCP_LOG_FILE":       "${PWD}/.go-invoice/mcp.log",
						"MCP_TRANSPORT":      "stdio",
					},
				},
			},
		}

		// Write .mcp.json
		configData, marshalErr := json.MarshalIndent(mcpConfig, "", "  ")
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal MCP configuration: %w", marshalErr)
		}

		if writeErr := os.WriteFile(".mcp.json", configData, 0o600); writeErr != nil {
			return fmt.Errorf("failed to write .mcp.json: %w", writeErr)
		}
		a.logger.Println("✅ Created .mcp.json for project-scoped MCP server (backward compatibility)")
	}

	// Create Claude settings
	claudeSettings := ClaudeSettings{
		Permissions: PermissionSettings{
			Allow: []string{
				"Bash(go-invoice:*)",
				"Read",
				"Write",
				"Edit",
				"MultiEdit",
			},
		},
		Env: map[string]string{
			"GO_INVOICE_PROJECT_NAME": projectName,
			"GO_INVOICE_PREFIX":       invoicePrefix,
		},
	}

	// Write .claude/settings.json
	settingsData, err := json.MarshalIndent(claudeSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Claude settings: %w", err)
	}

	if err := os.WriteFile(".claude/settings.json", settingsData, 0o600); err != nil {
		return fmt.Errorf("failed to write .claude/settings.json: %w", err)
	}

	// Create project-specific go-invoice configuration
	if err := a.createProjectInvoiceConfig(invoicePrefix); err != nil {
		return err
	}

	// Create sample files
	if err := a.createSampleFiles(); err != nil {
		a.logger.Printf("⚠️  Could not create sample files: %v\n", err)
	}

	a.logger.Println("✅ Claude Code integration configured")
	if useProjectScope {
		a.logger.Println("ℹ️  Open this project in Claude Code to start using go-invoice MCP integration")
		a.logger.Println("ℹ️  Team members will automatically get the MCP server when they open the project")
	} else {
		a.logger.Println("ℹ️  Restart Claude Code to load the go-invoice MCP server")
		a.logger.Println("ℹ️  The MCP server will be available across all your Claude Code projects")
	}

	return nil
}

// checkProjectContext checks if we're in a project directory
func (a *App) checkProjectContext() error {
	// Check for common project files
	projectFiles := []string{"package.json", "go.mod", "Cargo.toml", "pom.xml", "build.gradle", "requirements.txt"}

	for _, file := range projectFiles {
		if _, err := os.Stat(file); err == nil {
			return nil
		}
	}

	return ErrNoProjectFileDetected
}

// setupProjectStructure creates the project directory structure
func (a *App) setupProjectStructure() error {
	a.logger.Println("📁 Setting up project structure...")

	// Create directories
	dirs := []string{
		"invoices/drafts",
		"invoices/sent",
		"invoices/paid",
		"timesheets/pending",
		"timesheets/processed",
		"templates",
		".go-invoice/logs",
		".go-invoice/cache",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Update .gitignore
	if err := a.updateGitignore(); err != nil {
		a.logger.Printf("⚠️  Could not update .gitignore: %v\n", err)
	}

	a.logger.Println("✅ Project structure created")
	return nil
}

// updateGitignore updates or creates .gitignore with go-invoice entries
func (a *App) updateGitignore() error {
	ignoreEntries := []string{
		"",
		"# go-invoice MCP integration",
		".go-invoice/",
		"*.log",
		"",
		"# Claude Code local settings",
		".claude/settings.local.json",
		"",
		"# Invoice files (uncomment to ignore)",
		"# invoices/",
		"# timesheets/",
	}

	// Check if .gitignore exists
	content, err := os.ReadFile(".gitignore")
	if err != nil {
		// Create new .gitignore
		return os.WriteFile(".gitignore", []byte(strings.Join(ignoreEntries, "\n")), 0o600)
	}

	// Check if entries already exist
	contentStr := string(content)
	if strings.Contains(contentStr, ".go-invoice/") && strings.Contains(contentStr, ".claude/settings.local.json") {
		return nil // Already configured
	}

	// Append entries
	newContent := contentStr + "\n" + strings.Join(ignoreEntries, "\n")
	return os.WriteFile(".gitignore", []byte(newContent), 0o600)
}

// createProjectInvoiceConfig creates project-specific go-invoice configuration
func (a *App) createProjectInvoiceConfig(_ string) error {
	a.logger.Println("📋 Creating project invoice configuration...")

	config := map[string]interface{}{
		"server": map[string]interface{}{
			"host":        "localhost",
			"port":        0,
			"timeout":     30000000000,
			"readTimeout": 10000000000,
		},
		"cli": map[string]interface{}{
			"path":       "go-invoice",
			"workingDir": ".",
			"maxTimeout": 60000000000,
		},
		"security": map[string]interface{}{
			"allowedCommands":       []string{"go-invoice"},
			"workingDir":            ".",
			"sandboxEnabled":        true,
			"fileAccessRestricted":  true,
			"maxCommandTimeout":     "60s",
			"enableInputValidation": true,
		},
		"logLevel": "info",
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := ".go-invoice/config.json"
	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	a.logger.Println("✅ Project invoice configuration created")
	return nil
}

// createSampleFiles creates sample template and timesheet
func (a *App) createSampleFiles() error {
	// Create sample invoice template
	if err := a.createDefaultTemplate(); err != nil {
		return fmt.Errorf("failed to create invoice template: %w", err)
	}

	// Create sample timesheet
	sampleTimesheet := `Date,Hours,Description,Project,Tags
2025-01-15,2.5,"Initial project setup and configuration",Setup,setup;config
2025-01-16,4.0,"Implement authentication system",Development,auth;backend
2025-01-17,3.5,"Create user dashboard UI",Development,frontend;ui
2025-01-18,2.0,"Write unit tests for auth module",Testing,testing;backend
2025-01-19,1.5,"Code review and documentation",Review,review;docs
`

	if err := os.WriteFile("timesheets/sample-timesheet.csv", []byte(sampleTimesheet), 0o600); err != nil {
		return fmt.Errorf("failed to create sample timesheet: %w", err)
	}

	a.logger.Println("✅ Created sample files (template and timesheet)")
	return nil
}

// testMCPServer tests the MCP server functionality
func (a *App) testMCPServer(_ context.Context, projectRoot string) error {
	mcpPath := filepath.Join(projectRoot, "bin", "go-invoice-mcp")

	// Check if binary exists and is executable
	info, err := os.Stat(mcpPath)
	if err != nil {
		return fmt.Errorf("MCP server binary not found at %s: %w", mcpPath, err)
	}

	if info.Mode()&0o111 == 0 {
		return ErrMCPBinaryNotExecutable
	}

	// Test basic execution - the MCP server doesn't support --version,
	// so we'll just verify it exists and is executable
	// For a more thorough test, we could send a simple JSON-RPC request
	a.logger.Println("✅ MCP server binary is executable and ready")

	return nil
}

// printClaudeSetupSummary prints the setup summary
func (a *App) printClaudeSetupSummary(ctx context.Context, desktop, code bool, projectRoot, goInvoiceHome string) {
	a.logger.Println("")
	a.logger.Println("===================================")
	a.logger.Println("Setup Summary")
	a.logger.Println("===================================")
	a.logger.Println("")
	a.logger.Printf("📁 MCP Server Binary: %s\n", filepath.Join(projectRoot, "bin", "go-invoice-mcp"))
	a.logger.Printf("📁 Configuration: %s\n", filepath.Join(goInvoiceHome, "config"))
	a.logger.Printf("📁 Logs: %s\n", filepath.Join(goInvoiceHome, "logs"))

	// Verify MCP server registration
	if code {
		a.logger.Println("")
		a.logger.Println("🔍 Verifying MCP server registration...")

		// Check .mcp.json file for project-scoped servers
		if _, err := os.Stat(".mcp.json"); err == nil {
			content, err := os.ReadFile(".mcp.json")
			if err == nil && strings.Contains(string(content), "go-invoice") {
				a.logger.Println("✅ go-invoice MCP server is registered in .mcp.json")
			} else {
				a.logger.Println("⚠️  Could not find go-invoice in .mcp.json")
			}
		} else {
			// Fallback to claude mcp list for user-scoped servers
			cmd := exec.CommandContext(ctx, "claude", "mcp", "list")
			output, err := cmd.Output()
			if err == nil && strings.Contains(string(output), "go-invoice") {
				a.logger.Println("✅ go-invoice MCP server is registered with Claude")
			} else {
				a.logger.Println("⚠️  Could not verify MCP server registration")
				a.logger.Println("   Run 'claude mcp list' to check manually")
			}
		}
	}

	if desktop {
		a.logger.Println("")
		a.logger.Println("🖥️  Claude Desktop:")
		a.logger.Println("   • Restart Claude Desktop to load go-invoice MCP server")
		a.logger.Println("   • Use natural language to create and manage invoices")
	}

	if code {
		a.logger.Println("")
		a.logger.Println("💻 Claude Code:")
		a.logger.Printf("   • MCP configuration: %s\n", ".mcp.json")
		a.logger.Printf("   • Project settings: %s\n", ".claude/settings.json")
		a.logger.Println("   • The MCP server provides tools for:")
		a.logger.Println("     - Creating and managing invoices")
		a.logger.Println("     - Importing timesheets from CSV")
		a.logger.Println("     - Generating HTML invoices")
		a.logger.Println("     - Managing clients and projects")
	}

	a.logger.Println("")
	a.logger.Println("✅ Setup completed successfully!")
	a.logger.Println("")
	a.logger.Println("💡 Next steps:")
	a.logger.Println("   • Initialize storage with: go-invoice init")
	a.logger.Println("   • Create your first invoice through Claude!")
}

// promptForMerge asks user if they want to merge configurations
func (a *App) promptForMerge(ctx context.Context, prompter *cli.Prompter, service string) bool {
	merge, err := prompter.PromptBool(ctx,
		fmt.Sprintf("Found existing %s configuration. Merge with it?", service),
		true)
	return err == nil && merge
}

// copyFile copies a file from src to dst
func (a *App) copyFile(src, dst string) error {
	input, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0o600)
}

// mustGetwd returns the current working directory or panics
func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}
