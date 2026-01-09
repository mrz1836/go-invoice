package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BridgeError represents tool bridge errors.
type BridgeError struct {
	Op  string
	Msg string
}

func (e *BridgeError) Error() string {
	return fmt.Sprintf("bridge %s: %s", e.Op, e.Msg)
}

// Bridge errors
var (
	ErrToolNotFound               = &BridgeError{Op: "lookup", Msg: "tool not found"}
	ErrInvalidToolInput           = &BridgeError{Op: "validate", Msg: "invalid tool input"}
	ErrMissingRequired            = &BridgeError{Op: "validate", Msg: "missing required parameter"}
	ErrCommandBuildFailed         = &BridgeError{Op: "build", Msg: "failed to build command"}
	ErrMissingUpdateFields        = &BridgeError{Op: "validate", Msg: "at least one field to update must be provided"}
	ErrMissingItemIdentifier      = &BridgeError{Op: "validate", Msg: "either item_id or item_index must be provided"}
	ErrMissingClientIdentifier    = &BridgeError{Op: "validate", Msg: "one of client_id, client_name, or client_email must be provided"}
	ErrMissingClientIDOrName      = &BridgeError{Op: "validate", Msg: "client_id or client_name must be provided"}
	ErrMissingClientNameForCreate = &BridgeError{Op: "validate", Msg: "client_name required when create_new is true"}
	ErrCommandFailed              = &BridgeError{Op: "execute", Msg: "command execution failed"}
	ErrCollectionFailed           = &BridgeError{Op: "collect", Msg: "file collection failed"}
	ErrBatchNotSupported          = &BridgeError{Op: "validate", Msg: "batch invoice generation is not supported by the CLI - please generate invoices individually"}
)

// ToolCommand represents the mapping from an MCP tool to CLI command.
type ToolCommand struct {
	// Tool is the MCP tool name
	Tool string

	// Command is the base CLI command
	Command string

	// SubCommands are the CLI subcommands
	SubCommands []string

	// BuildArgs builds the command arguments from tool input
	BuildArgs func(input map[string]interface{}) ([]string, error)

	// RequiresFiles indicates if this command needs file handling
	RequiresFiles bool

	// OutputPatterns are glob patterns for expected output files
	OutputPatterns []string

	// ExpectJSON indicates if the command outputs JSON
	ExpectJSON bool

	// Timeout is the specific timeout for this command
	Timeout time.Duration
}

// CLIBridge bridges MCP tool requests to CLI command execution.
type CLIBridge struct {
	logger       Logger
	executor     CommandExecutor
	fileHandler  FileHandler
	toolCommands map[string]*ToolCommand
	cliPath      string
}

// NewCLIBridge creates a new CLI bridge.
func NewCLIBridge(logger Logger, executor CommandExecutor, fileHandler FileHandler, cliPath string) *CLIBridge {
	if logger == nil {
		panic("logger is required")
	}
	if executor == nil {
		panic("executor is required")
	}
	if fileHandler == nil {
		panic("fileHandler is required")
	}

	// Default CLI path
	if cliPath == "" {
		cliPath = "go-invoice"
	}

	bridge := &CLIBridge{
		logger:       logger,
		executor:     executor,
		fileHandler:  fileHandler,
		toolCommands: make(map[string]*ToolCommand),
		cliPath:      cliPath,
	}

	// Register all tool commands
	bridge.registerToolCommands()

	return bridge
}

// ExecuteToolCommand executes a CLI command for an MCP tool.
func (b *CLIBridge) ExecuteToolCommand(ctx context.Context, toolName string, input map[string]interface{}) (*ExecutionResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Find tool command mapping
	toolCmd, exists := b.toolCommands[toolName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolName)
	}

	// Build command arguments
	args, err := toolCmd.BuildArgs(input)
	if err != nil {
		b.logger.Error("argument building failed",
			"tool", toolName,
			"error", err,
			"input", input,
		)
		return nil, fmt.Errorf("%w: %w", ErrCommandBuildFailed, err)
	}

	// For go-invoice imports, we need to handle dynamic subcommands
	var finalSubCommands []string
	if toolName == "import_csv" {
		// Determine subcommand based on import mode (skip for debug echo command)
		if importMode, ok := input["import_mode"].(string); ok && importMode == "append_invoice" {
			finalSubCommands = []string{"import", "append"}
		} else {
			finalSubCommands = []string{"import", "create"}
		}
	} else {
		finalSubCommands = toolCmd.SubCommands
	}

	// For go-invoice, config must come before subcommands as it's a global flag
	var fullArgs []string

	// Check if args start with --config (from getConfigArgs)
	if len(args) >= 2 && args[0] == "--config" {
		// Place config args before subcommands
		fullArgs = append(fullArgs, args[0], args[1])
		fullArgs = append(fullArgs, finalSubCommands...)
		if len(args) > 2 {
			fullArgs = append(fullArgs, args[2:]...)
		}
	} else {
		// Default behavior
		fullArgs = append(finalSubCommands, args...)
	}

	// Prepare execution request
	req := &ExecutionRequest{
		Command:    b.cliPath,
		Args:       fullArgs,
		ExpectJSON: toolCmd.ExpectJSON,
		Timeout:    toolCmd.Timeout,
	}

	// Handle file operations if needed
	if toolCmd.RequiresFiles {
		if prepareErr := b.prepareFilesForCommand(ctx, req, input); prepareErr != nil {
			b.logger.Error("file preparation failed",
				"tool", toolName,
				"error", prepareErr,
				"input", input,
			)
			return nil, fmt.Errorf("file preparation failed: %w", prepareErr)
		}
	}

	// Log command execution
	b.logger.Info("executing tool command",
		"tool", toolName,
		"command", req.Command,
		"args", req.Args,
	)

	// Execute the command
	resp, err := b.executor.Execute(ctx, req)
	if err != nil {
		b.logger.Error("command execution failed",
			"tool", toolName,
			"command", req.Command,
			"args", req.Args,
			"error", err,
		)
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	// Collect output files if patterns are specified
	if len(toolCmd.OutputPatterns) > 0 && resp.ExitCode == 0 {
		outputFiles, err := b.fileHandler.CollectOutputFiles(ctx, req.WorkingDir, toolCmd.OutputPatterns)
		if err != nil {
			b.logger.Warn("failed to collect output files",
				"tool", toolName,
				"error", err,
			)
		} else {
			resp.OutputFiles = outputFiles
		}
	}

	return resp, nil
}

// prepareFilesForCommand prepares files for command execution.
func (b *CLIBridge) prepareFilesForCommand(ctx context.Context, req *ExecutionRequest, input map[string]interface{}) error {
	// Check for file_path parameter (common in import operations)
	if filePath, ok := input["file_path"].(string); ok {
		// Validate the file
		if err := b.fileHandler.ValidateFile(ctx, filePath); err != nil {
			return fmt.Errorf("file validation failed: %w", err)
		}

		// Add to input files
		req.InputFiles = append(req.InputFiles, FileReference{
			Path:        filePath,
			ContentType: "text/csv", // Assume CSV for now, could be smarter
		})
	}

	return nil
}

// registerToolCommands registers all tool-to-command mappings.
func (b *CLIBridge) registerToolCommands() {
	// Invoice management tools
	b.toolCommands["invoice_create"] = &ToolCommand{
		Tool:        "invoice_create",
		Command:     b.cliPath,
		SubCommands: []string{"invoice", "create"},
		BuildArgs:   b.buildInvoiceCreateArgs,
		Timeout:     10 * time.Second,
	}

	b.toolCommands["invoice_list"] = &ToolCommand{
		Tool:        "invoice_list",
		Command:     b.cliPath,
		SubCommands: []string{"invoice", "list"},
		BuildArgs:   b.buildInvoiceListArgs,
		ExpectJSON:  true,
		Timeout:     10 * time.Second,
	}

	b.toolCommands["invoice_show"] = &ToolCommand{
		Tool:        "invoice_show",
		Command:     b.cliPath,
		SubCommands: []string{"invoice", "show"},
		BuildArgs:   b.buildInvoiceShowArgs,
		ExpectJSON:  true,
		Timeout:     5 * time.Second,
	}

	b.toolCommands["invoice_update"] = &ToolCommand{
		Tool:        "invoice_update",
		Command:     b.cliPath,
		SubCommands: []string{"invoice", "update"},
		BuildArgs:   b.buildInvoiceUpdateArgs,
		Timeout:     10 * time.Second,
	}

	b.toolCommands["invoice_delete"] = &ToolCommand{
		Tool:        "invoice_delete",
		Command:     b.cliPath,
		SubCommands: []string{"invoice", "delete"},
		BuildArgs:   b.buildInvoiceDeleteArgs,
		Timeout:     5 * time.Second,
	}

	b.toolCommands["invoice_add_item"] = &ToolCommand{
		Tool:        "invoice_add_item",
		Command:     b.cliPath,
		SubCommands: []string{"invoice", "add-item"},
		BuildArgs:   b.buildInvoiceAddItemArgs,
		Timeout:     10 * time.Second,
	}

	b.toolCommands["invoice_remove_item"] = &ToolCommand{
		Tool:        "invoice_remove_item",
		Command:     b.cliPath,
		SubCommands: []string{"invoice", "remove-item"},
		BuildArgs:   b.buildInvoiceRemoveItemArgs,
		Timeout:     10 * time.Second,
	}

	// Client management tools
	b.toolCommands["client_create"] = &ToolCommand{
		Tool:        "client_create",
		Command:     b.cliPath,
		SubCommands: []string{"client", "create"},
		BuildArgs:   b.buildClientCreateArgs,
		Timeout:     10 * time.Second,
	}

	b.toolCommands["client_list"] = &ToolCommand{
		Tool:        "client_list",
		Command:     b.cliPath,
		SubCommands: []string{"client", "list"},
		BuildArgs:   b.buildClientListArgs,
		ExpectJSON:  true,
		Timeout:     10 * time.Second,
	}

	b.toolCommands["client_show"] = &ToolCommand{
		Tool:        "client_show",
		Command:     b.cliPath,
		SubCommands: []string{"client", "show"},
		BuildArgs:   b.buildClientShowArgs,
		ExpectJSON:  true,
		Timeout:     5 * time.Second,
	}

	b.toolCommands["client_update"] = &ToolCommand{
		Tool:        "client_update",
		Command:     b.cliPath,
		SubCommands: []string{"client", "update"},
		BuildArgs:   b.buildClientUpdateArgs,
		Timeout:     10 * time.Second,
	}

	b.toolCommands["client_delete"] = &ToolCommand{
		Tool:        "client_delete",
		Command:     b.cliPath,
		SubCommands: []string{"client", "delete"},
		BuildArgs:   b.buildClientDeleteArgs,
		Timeout:     5 * time.Second,
	}

	// Import tools
	b.toolCommands["import_csv"] = &ToolCommand{
		Tool:          "import_csv",
		Command:       b.cliPath,
		SubCommands:   []string{"import"}, // Will add subcommand dynamically based on mode
		BuildArgs:     b.buildImportCSVArgs,
		RequiresFiles: false, // Keep disabled for now
		Timeout:       30 * time.Second,
	}

	b.toolCommands["import_validate"] = &ToolCommand{
		Tool:          "import_validate",
		Command:       b.cliPath,
		SubCommands:   []string{"import", "validate"},
		BuildArgs:     b.buildImportValidateArgs,
		RequiresFiles: false, // Temporarily disable file validation to test
		Timeout:       10 * time.Second,
	}

	b.toolCommands["import_preview"] = &ToolCommand{
		Tool:          "import_preview",
		Command:       b.cliPath,
		SubCommands:   []string{"import", "validate"},
		BuildArgs:     b.buildImportPreviewArgs,
		RequiresFiles: false, // Temporarily disable file validation to test
		ExpectJSON:    false, // validate command doesn't output JSON by default
		Timeout:       10 * time.Second,
	}

	// Generation tools
	b.toolCommands["generate_html"] = &ToolCommand{
		Tool:           "generate_html",
		Command:        b.cliPath,
		SubCommands:    []string{"generate", "invoice"},
		BuildArgs:      b.buildGenerateHTMLArgs,
		OutputPatterns: []string{"invoice-*.html", "*.html"},
		Timeout:        20 * time.Second,
	}

	b.toolCommands["generate_summary"] = &ToolCommand{
		Tool:           "generate_summary",
		Command:        b.cliPath,
		SubCommands:    []string{"summary"},
		BuildArgs:      b.buildGenerateSummaryArgs,
		ExpectJSON:     true,
		OutputPatterns: []string{"summary-*.html", "report-*.html"},
		Timeout:        30 * time.Second,
	}

	b.toolCommands["export_data"] = &ToolCommand{
		Tool:           "export_data",
		Command:        b.cliPath,
		SubCommands:    []string{"export"},
		BuildArgs:      b.buildExportDataArgs,
		OutputPatterns: []string{"export-*.json", "export-*.csv", "export-*.xlsx"},
		Timeout:        30 * time.Second,
	}

	// Configuration tools
	b.toolCommands["config_show"] = &ToolCommand{
		Tool:        "config_show",
		Command:     b.cliPath,
		SubCommands: []string{"config", "show"},
		BuildArgs:   b.buildConfigShowArgs,
		ExpectJSON:  true,
		Timeout:     5 * time.Second,
	}

	b.toolCommands["config_validate"] = &ToolCommand{
		Tool:        "config_validate",
		Command:     b.cliPath,
		SubCommands: []string{"config", "validate"},
		BuildArgs:   b.buildConfigValidateArgs,
		Timeout:     5 * time.Second,
	}

	b.toolCommands["config_init"] = &ToolCommand{
		Tool:        "config_init",
		Command:     b.cliPath,
		SubCommands: []string{"config", "init"},
		BuildArgs:   b.buildConfigInitArgs,
		Timeout:     10 * time.Second,
	}
}

// Helper functions to build command arguments for each tool
// These map the MCP tool input parameters to CLI arguments

// getConfigArgs returns the config path arguments for MCP commands
func (b *CLIBridge) getConfigArgs() []string {
	// Always use the absolute config path for MCP
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".go-invoice", ".env.config")
	return []string{"--config", configPath}
}

func (b *CLIBridge) buildInvoiceCreateArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Handle different client identifier options
	if clientID, ok := input["client_id"].(string); ok && clientID != "" {
		args = append(args, "--client-id", clientID)
	} else if clientName, ok := input["client_name"].(string); ok && clientName != "" {
		args = append(args, "--client", clientName)
	} else if clientEmail, ok := input["client_email"].(string); ok && clientEmail != "" {
		args = append(args, "--client-email", clientEmail)
	} else {
		return nil, fmt.Errorf("%w: one of client_id, client_name, or client_email is required", ErrMissingRequired)
	}

	// Optional parameters
	if description, ok := input["description"].(string); ok && description != "" {
		args = append(args, "--description", description)
	}
	if invoiceDate, ok := input["invoice_date"].(string); ok && invoiceDate != "" {
		args = append(args, "--date", invoiceDate)
	}
	if dueDate, ok := input["due_date"].(string); ok && dueDate != "" {
		args = append(args, "--due", dueDate)
	}

	// Handle work_items if provided
	if workItems, ok := input["work_items"].([]interface{}); ok && len(workItems) > 0 {
		// Add work items directly during creation
		for _, item := range workItems {
			if workItem, ok := item.(map[string]interface{}); ok {
				if description, ok := workItem["description"].(string); ok && description != "" {
					args = append(args, "--add-item-description", description)
				}
				if hours, ok := getFloatValue(workItem["hours"]); ok {
					args = append(args, "--add-item-hours", fmt.Sprintf("%.2f", hours))
				}
				if rate, ok := getFloatValue(workItem["rate"]); ok {
					args = append(args, "--add-item-rate", fmt.Sprintf("%.2f", rate))
				}
				if date, ok := workItem["date"].(string); ok && date != "" {
					args = append(args, "--add-item-date", date)
				}
			}
		}
	}

	// Handle create_client_if_missing
	if createClient, ok := input["create_client_if_missing"].(bool); ok && createClient {
		args = append(args, "--create-client")

		// Add new client details if provided
		if email, ok := input["new_client_email"].(string); ok && email != "" {
			args = append(args, "--new-client-email", email)
		}
		if phone, ok := input["new_client_phone"].(string); ok && phone != "" {
			args = append(args, "--new-client-phone", phone)
		}
		if address, ok := input["new_client_address"].(string); ok && address != "" {
			args = append(args, "--new-client-address", address)
		}
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceListArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	args = append(args, "--output", "json") // Always output JSON for MCP

	// Optional filters
	if status, ok := input["status"].(string); ok && status != "" {
		args = append(args, "--status", status)
	}
	if clientName, ok := input["client_name"].(string); ok && clientName != "" {
		args = append(args, "--client", clientName)
	}
	if fromDate, ok := input["from_date"].(string); ok && fromDate != "" {
		args = append(args, "--from", fromDate)
	}
	if toDate, ok := input["to_date"].(string); ok && toDate != "" {
		args = append(args, "--to", toDate)
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceShowArgs(input map[string]interface{}) ([]string, error) {
	args := make([]string, 0, 4)

	// Either invoice_id or invoice_number is required
	invoiceID, hasID := input["invoice_id"].(string)
	invoiceNumber, hasNumber := input["invoice_number"].(string)

	if (!hasID || invoiceID == "") && (!hasNumber || invoiceNumber == "") {
		return nil, fmt.Errorf("%w: either invoice_id or invoice_number is required", ErrMissingRequired)
	}

	// Prefer invoice_id if both are provided
	identifier := invoiceID
	if identifier == "" {
		identifier = invoiceNumber
	}

	args = append(args, identifier, "--output", "json")

	// Add config path
	configArgs := b.getConfigArgs()
	args = append(args, configArgs...)

	return args, nil
}

func (b *CLIBridge) buildInvoiceUpdateArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: invoice_id or invoice_number
	invoiceID, hasID := input["invoice_id"].(string)
	invoiceNumber, hasNumber := input["invoice_number"].(string)

	if (!hasID || invoiceID == "") && (!hasNumber || invoiceNumber == "") {
		return nil, fmt.Errorf("%w: either invoice_id or invoice_number is required", ErrMissingRequired)
	}

	// Prefer invoice_id if both are provided
	identifier := invoiceID
	if identifier == "" {
		identifier = invoiceNumber
	}
	args = append(args, identifier)

	// At least one update field required
	hasUpdate := false
	if status, ok := input["status"].(string); ok && status != "" {
		args = append(args, "--status", status)
		hasUpdate = true
	}
	if dueDate, ok := input["due_date"].(string); ok && dueDate != "" {
		args = append(args, "--due", dueDate)
		hasUpdate = true
	}
	if description, ok := input["description"].(string); ok && description != "" {
		args = append(args, "--description", description)
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, ErrMissingUpdateFields
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceDeleteArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: invoice_id or invoice_number
	invoiceID, hasID := input["invoice_id"].(string)
	invoiceNumber, hasNumber := input["invoice_number"].(string)

	if (!hasID || invoiceID == "") && (!hasNumber || invoiceNumber == "") {
		return nil, fmt.Errorf("%w: either invoice_id or invoice_number is required", ErrMissingRequired)
	}

	// Prefer invoice_id if both are provided
	identifier := invoiceID
	if identifier == "" {
		identifier = invoiceNumber
	}

	// Add the identifier as positional argument first
	args = append(args, identifier)

	// Optional: hard_delete (order matters - CLI expects it before --force)
	if hardDelete, ok := input["hard_delete"].(bool); ok && hardDelete {
		args = append(args, "--hard")
	}

	// Optional: force (should come after other flags)
	if force, ok := input["force"].(bool); ok && force {
		args = append(args, "--force")
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceAddItemArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: invoice_id or invoice_number
	invoiceID, hasID := input["invoice_id"].(string)
	invoiceNumber, hasNumber := input["invoice_number"].(string)

	if (!hasID || invoiceID == "") && (!hasNumber || invoiceNumber == "") {
		return nil, fmt.Errorf("%w: either invoice_id or invoice_number is required", ErrMissingRequired)
	}

	// Prefer invoice_id if both are provided
	identifier := invoiceID
	if identifier == "" {
		identifier = invoiceNumber
	}
	args = append(args, identifier)

	// Check if we have work_items array (MCP schema format)
	if workItems, ok := input["work_items"].([]interface{}); ok && len(workItems) > 0 {
		// Handle array of work items
		for _, item := range workItems {
			if workItem, ok := item.(map[string]interface{}); ok {
				// Add each work item
				if description, ok := workItem["description"].(string); ok && description != "" {
					args = append(args, "--description", description)
				}

				if hours, ok := getFloatValue(workItem["hours"]); ok {
					args = append(args, "--hours", fmt.Sprintf("%.2f", hours))
				}

				if rate, ok := getFloatValue(workItem["rate"]); ok {
					args = append(args, "--rate", fmt.Sprintf("%.2f", rate))
				}

				if date, ok := workItem["date"].(string); ok && date != "" {
					args = append(args, "--date", date)
				}

				// This assumes the CLI can handle multiple work items with repeated flags
				// If not, we may need to adjust the CLI or use a different approach
			}
		}
	} else {
		// Fallback to individual parameters for backwards compatibility
		description, hasDesc := input["description"].(string)
		hours, hasHours := getFloatValue(input["hours"])
		rate, hasRate := getFloatValue(input["rate"])

		if !hasDesc || description == "" {
			return nil, fmt.Errorf("%w: description or work_items", ErrMissingRequired)
		}
		if !hasHours {
			return nil, fmt.Errorf("%w: hours or work_items", ErrMissingRequired)
		}
		if !hasRate {
			return nil, fmt.Errorf("%w: rate or work_items", ErrMissingRequired)
		}

		args = append(args, "--description", description)
		args = append(args, "--hours", fmt.Sprintf("%.2f", hours))
		args = append(args, "--rate", fmt.Sprintf("%.2f", rate))

		// Optional: date
		if date, ok := input["date"].(string); ok && date != "" {
			args = append(args, "--date", date)
		}
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceRemoveItemArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: invoice_id or invoice_number
	invoiceID, hasID := input["invoice_id"].(string)
	invoiceNumber, hasNumber := input["invoice_number"].(string)

	if (!hasID || invoiceID == "") && (!hasNumber || invoiceNumber == "") {
		return nil, fmt.Errorf("%w: either invoice_id or invoice_number is required", ErrMissingRequired)
	}

	// Prefer invoice_id if both are provided
	identifier := invoiceID
	if identifier == "" {
		identifier = invoiceNumber
	}
	args = append(args, identifier)

	// Determine removal criteria - work_item_id, work_item_description, work_item_date, or legacy item_id/item_index
	hasRemovalCriteria := false

	if workItemID, ok := input["work_item_id"].(string); ok && workItemID != "" {
		args = append(args, "--item-id", workItemID)
		hasRemovalCriteria = true
	} else if itemID, ok := input["item_id"].(string); ok && itemID != "" {
		// Legacy support
		args = append(args, "--item-id", itemID)
		hasRemovalCriteria = true
	}

	if workItemDesc, ok := input["work_item_description"].(string); ok && workItemDesc != "" {
		args = append(args, "--description", workItemDesc)
		hasRemovalCriteria = true
	}

	if workItemDate, ok := input["work_item_date"].(string); ok && workItemDate != "" {
		args = append(args, "--date", workItemDate)
		hasRemovalCriteria = true
	}

	// Legacy index support
	if !hasRemovalCriteria {
		if itemIndex, ok := getIntValue(input["item_index"]); ok {
			args = append(args, "--index", fmt.Sprintf("%d", itemIndex))
			hasRemovalCriteria = true
		}
	}

	if !hasRemovalCriteria {
		return nil, fmt.Errorf("%w: one of work_item_id, work_item_description, work_item_date, or item_index is required", ErrMissingRequired)
	}

	// Optional: remove_all_matching
	if removeAll, ok := input["remove_all_matching"].(bool); ok && removeAll {
		args = append(args, "--all")
	}

	// Optional: confirm
	if confirm, ok := input["confirm"].(bool); ok && confirm {
		args = append(args, "--yes")
	}

	return args, nil
}

func (b *CLIBridge) buildClientCreateArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: name
	name, ok := input["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("%w: name", ErrMissingRequired)
	}
	args = append(args, "--name", name)

	// Optional parameters
	if email, ok := input["email"].(string); ok && email != "" {
		args = append(args, "--email", email)
	}
	if phone, ok := input["phone"].(string); ok && phone != "" {
		args = append(args, "--phone", phone)
	}
	if address, ok := input["address"].(string); ok && address != "" {
		args = append(args, "--address", address)
	}
	if taxID, ok := input["tax_id"].(string); ok && taxID != "" {
		args = append(args, "--tax-id", taxID)
	}

	return args, nil
}

func (b *CLIBridge) buildClientListArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	args = append(args, "--output", "json") // Always output JSON for MCP

	// Optional filters
	if active, ok := input["active"].(bool); ok {
		if active {
			args = append(args, "--active")
		} else {
			args = append(args, "--inactive")
		}
	}
	// Check for name_search (MCP parameter) or search
	if search, ok := input["name_search"].(string); ok && search != "" {
		args = append(args, "--search", search)
	} else if search, ok := input["search"].(string); ok && search != "" {
		args = append(args, "--search", search)
	}

	return args, nil
}

func (b *CLIBridge) buildClientShowArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: client identifier (name, email, or id)
	if clientID, ok := input["client_id"].(string); ok && clientID != "" {
		args = append(args, clientID, "--output", "json")
	} else if clientName, ok := input["client_name"].(string); ok && clientName != "" {
		args = append(args, clientName, "--output", "json")
	} else if clientEmail, ok := input["client_email"].(string); ok && clientEmail != "" {
		args = append(args, clientEmail, "--output", "json")
	} else {
		return nil, ErrMissingClientIdentifier
	}

	return args, nil
}

func (b *CLIBridge) buildClientUpdateArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: client identifier
	if clientID, ok := input["client_id"].(string); ok && clientID != "" {
		args = append(args, clientID)
	} else if clientName, ok := input["client_name"].(string); ok && clientName != "" {
		args = append(args, clientName)
	} else {
		return nil, ErrMissingClientIDOrName
	}

	// At least one update field required
	hasUpdate := false
	if name, ok := input["name"].(string); ok && name != "" {
		args = append(args, "--name", name)
		hasUpdate = true
	}
	if email, ok := input["email"].(string); ok && email != "" {
		args = append(args, "--email", email)
		hasUpdate = true
	}
	if phone, ok := input["phone"].(string); ok && phone != "" {
		args = append(args, "--phone", phone)
		hasUpdate = true
	}
	if address, ok := input["address"].(string); ok && address != "" {
		args = append(args, "--address", address)
		hasUpdate = true
	}
	if active, ok := input["active"].(bool); ok {
		if active {
			args = append(args, "--activate")
		} else {
			args = append(args, "--deactivate")
		}
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, ErrMissingUpdateFields
	}

	return args, nil
}

func (b *CLIBridge) buildClientDeleteArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: client identifier
	if clientID, ok := input["client_id"].(string); ok && clientID != "" {
		args = append(args, clientID)
	} else if clientName, ok := input["client_name"].(string); ok && clientName != "" {
		args = append(args, clientName)
	} else {
		return nil, ErrMissingClientIDOrName
	}

	// Optional: force
	if force, ok := input["force"].(bool); ok && force {
		args = append(args, "--force")
	}

	return args, nil
}

func (b *CLIBridge) buildImportCSVArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: file_path
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("%w: file_path", ErrMissingRequired)
	}
	args = append(args, filePath)

	// Handle import_mode and required parameters
	importMode, hasMode := input["import_mode"].(string)
	if hasMode && importMode == "append_invoice" {
		// For append mode, we need an invoice ID
		if invoiceID, ok := input["invoice_id"].(string); ok && invoiceID != "" {
			args = append(args, "--invoice", invoiceID)
		} else if invoiceNumber, ok := input["invoice_number"].(string); ok && invoiceNumber != "" {
			args = append(args, "--invoice", invoiceNumber)
		} else {
			return nil, fmt.Errorf("%w: invoice_id or invoice_number is required for append_invoice mode", ErrMissingRequired)
		}
	} else {
		// Create new invoice mode - need client identifier
		if clientID, ok := input["client_id"].(string); ok && clientID != "" {
			args = append(args, "--client-id", clientID)
		} else if clientName, ok := input["client_name"].(string); ok && clientName != "" {
			args = append(args, "--client", clientName)
		} else if clientEmail, ok := input["client_email"].(string); ok && clientEmail != "" {
			args = append(args, "--client-email", clientEmail)
		} else {
			return nil, fmt.Errorf("%w: client_id, client_name, or client_email is required for new invoice", ErrMissingRequired)
		}

		// Optional: description for new invoice
		if description, ok := input["description"].(string); ok && description != "" {
			args = append(args, "--description", description)
		}

		// Optional: invoice_date for new invoice
		if invoiceDate, ok := input["invoice_date"].(string); ok && invoiceDate != "" {
			args = append(args, "--date", invoiceDate)
		}

		// Optional: due_days for new invoice
		if dueDays, ok := getIntValue(input["due_days"]); ok {
			args = append(args, "--due-days", fmt.Sprintf("%d", dueDays))
		}
	}

	// Common optional parameters for both modes
	if defaultRate, ok := getFloatValue(input["default_rate"]); ok {
		args = append(args, "--default-rate", fmt.Sprintf("%.2f", defaultRate))
	}

	if rateOverride, ok := input["rate_override"].(bool); ok && rateOverride {
		args = append(args, "--rate-override")
	}

	if dryRun, ok := input["dry_run"].(bool); ok && dryRun {
		args = append(args, "--dry-run")
	}

	if delimiter, ok := input["delimiter"].(string); ok && delimiter != "" {
		args = append(args, "--delimiter", delimiter)
	}

	if hasHeader, ok := input["has_header"].(bool); ok && !hasHeader {
		args = append(args, "--no-header")
	}

	if currency, ok := input["currency"].(string); ok && currency != "" {
		args = append(args, "--currency", currency)
	}

	return args, nil
}

func (b *CLIBridge) buildImportValidateArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: file_path
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("%w: file_path", ErrMissingRequired)
	}
	args = append(args, filePath)

	// Optional: delimiter
	if delimiter, ok := input["delimiter"].(string); ok && delimiter != "" {
		args = append(args, "--delimiter", delimiter)
	}

	// Optional: has_header
	if hasHeader, ok := input["has_header"].(bool); ok && !hasHeader {
		args = append(args, "--no-header")
	}

	// Optional: validate_rates
	if validateRates, ok := input["validate_rates"].(bool); ok && validateRates {
		args = append(args, "--validate-rates")
	}

	// Optional: validate_dates
	if validateDates, ok := input["validate_dates"].(bool); ok && validateDates {
		args = append(args, "--validate-dates")
	}

	// Optional: validate_business
	if validateBusiness, ok := input["validate_business"].(bool); ok && validateBusiness {
		args = append(args, "--validate-business")
	}

	// Optional: max_hours_per_day
	if maxHours, ok := getFloatValue(input["max_hours_per_day"]); ok {
		args = append(args, "--max-hours", fmt.Sprintf("%.1f", maxHours))
	}

	// Optional: min_rate
	if minRate, ok := getFloatValue(input["min_rate"]); ok {
		args = append(args, "--min-rate", fmt.Sprintf("%.2f", minRate))
	}

	// Optional: max_rate
	if maxRate, ok := getFloatValue(input["max_rate"]); ok {
		args = append(args, "--max-rate", fmt.Sprintf("%.2f", maxRate))
	}

	// Optional: check_duplicates
	if checkDuplicates, ok := input["check_duplicates"].(bool); ok && checkDuplicates {
		args = append(args, "--check-duplicates")
	}

	// Optional: check_weekends
	if checkWeekends, ok := input["check_weekends"].(bool); ok && checkWeekends {
		args = append(args, "--check-weekends")
	}

	// Optional: quick_validate
	if quickValidate, ok := input["quick_validate"].(bool); ok && quickValidate {
		args = append(args, "--quick")
	}

	// Optional: strict validation (legacy)
	if strict, ok := input["strict"].(bool); ok && strict {
		args = append(args, "--strict")
	}

	return args, nil
}

func (b *CLIBridge) buildImportPreviewArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: file_path
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("%w: file_path", ErrMissingRequired)
	}
	args = append(args, filePath)

	// For preview, we use the validate command with dry-run
	args = append(args, "--dry-run")

	// Optional: delimiter
	if delimiter, ok := input["delimiter"].(string); ok && delimiter != "" {
		args = append(args, "--delimiter", delimiter)
	}

	// Optional: has_header
	if hasHeader, ok := input["has_header"].(bool); ok && !hasHeader {
		args = append(args, "--no-header")
	}

	return args, nil
}

func (b *CLIBridge) buildGenerateHTMLArgs(input map[string]interface{}) ([]string, error) {
	args := b.getConfigArgs()

	// Required: invoice identifier (invoice_id or invoice_number)
	invoiceID, hasID := input["invoice_id"].(string)
	invoiceNumber, hasNumber := input["invoice_number"].(string)

	if (!hasID || invoiceID == "") && (!hasNumber || invoiceNumber == "") {
		return nil, fmt.Errorf("%w: either invoice_id or invoice_number is required", ErrMissingRequired)
	}

	// Prefer invoice_id if both are provided
	identifier := invoiceID
	if identifier == "" {
		identifier = invoiceNumber
	}

	// Add invoice identifier as positional argument (this will be added after config args)
	args = append(args, identifier)

	// Optional: template
	if template, ok := input["template"].(string); ok && template != "" {
		args = append(args, "--template", template)
	}

	// Optional: output path - only pass --output if explicitly provided
	// Let the CLI handle its own default path logic
	if outputPath, ok := input["output_path"].(string); ok && outputPath != "" {
		args = append(args, "--output", outputPath)
	}

	// Handle batch_invoices if provided (not supported by CLI, but we should handle gracefully)
	if batchInvoices, ok := input["batch_invoices"].([]interface{}); ok && len(batchInvoices) > 0 {
		return nil, ErrBatchNotSupported
	}

	// Skip unsupported MCP schema parameters that don't have CLI equivalents
	// These parameters are part of the MCP schema but not implemented in the CLI:
	// - company_name, custom_css, footer_text, include_logo, include_notes
	// - web_preview, return_html (these conflict with file-based output)
	_ = input["company_name"]
	_ = input["custom_css"]
	_ = input["footer_text"]
	_ = input["include_logo"]
	_ = input["include_notes"]
	_ = input["web_preview"]
	_ = input["return_html"]

	return args, nil
}

func (b *CLIBridge) buildGenerateSummaryArgs(input map[string]interface{}) ([]string, error) {
	var args []string
	args = append(args, "--output", "json") // Output JSON for MCP

	// Required: summary_type
	summaryType, ok := input["summary_type"].(string)
	if !ok || summaryType == "" {
		summaryType = "revenue" // Default
	}
	args = append(args, "--type", summaryType)

	// Optional: period
	if period, ok := input["period"].(string); ok && period != "" {
		args = append(args, "--period", period)
	}

	// Optional: from/to dates
	if fromDate, ok := input["from_date"].(string); ok && fromDate != "" {
		args = append(args, "--from", fromDate)
	}
	if toDate, ok := input["to_date"].(string); ok && toDate != "" {
		args = append(args, "--to", toDate)
	}

	// Optional: group by
	if groupBy, ok := input["group_by"].(string); ok && groupBy != "" {
		args = append(args, "--group-by", groupBy)
	}

	return args, nil
}

func (b *CLIBridge) buildExportDataArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: export_type
	exportType, ok := input["export_type"].(string)
	if !ok || exportType == "" {
		exportType = "invoices" // Default
	}
	args = append(args, "--type", exportType)

	// Required: format
	format, ok := input["format"].(string)
	if !ok || format == "" {
		format = "json" // Default
	}
	args = append(args, "--format", format)

	// Optional: filters
	if status, ok := input["status"].(string); ok && status != "" {
		args = append(args, "--status", status)
	}
	if clientName, ok := input["client_name"].(string); ok && clientName != "" {
		args = append(args, "--client", clientName)
	}
	if fromDate, ok := input["from_date"].(string); ok && fromDate != "" {
		args = append(args, "--from", fromDate)
	}
	if toDate, ok := input["to_date"].(string); ok && toDate != "" {
		args = append(args, "--to", toDate)
	}

	// Optional: output file
	if outputFile, ok := input["output_file"].(string); ok && outputFile != "" {
		args = append(args, "--output", outputFile)
	}

	return args, nil
}

func (b *CLIBridge) buildConfigShowArgs(input map[string]interface{}) ([]string, error) {
	var args []string
	args = append(args, "--output", "json") // Always output JSON for MCP

	// Optional: section
	if section, ok := input["section"].(string); ok && section != "" {
		args = append(args, "--section", section)
	}

	// Optional: show defaults
	if showDefaults, ok := input["show_defaults"].(bool); ok && showDefaults {
		args = append(args, "--show-defaults")
	}

	return args, nil
}

func (b *CLIBridge) buildConfigValidateArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Optional: config file path
	if configPath, ok := input["config_path"].(string); ok && configPath != "" {
		args = append(args, "--file", configPath)
	}

	// Optional: strict validation
	if strict, ok := input["strict"].(bool); ok && strict {
		args = append(args, "--strict")
	}

	return args, nil
}

func (b *CLIBridge) buildConfigInitArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Optional: template
	if template, ok := input["template"].(string); ok && template != "" {
		args = append(args, "--template", template)
	}

	// Optional: output path
	if outputPath, ok := input["output_path"].(string); ok && outputPath != "" {
		args = append(args, "--output", outputPath)
	}

	// Optional: force overwrite
	if force, ok := input["force"].(bool); ok && force {
		args = append(args, "--force")
	}

	// Optional: interactive mode
	if interactive, ok := input["interactive"].(bool); ok && interactive {
		args = append(args, "--interactive")
	}

	return args, nil
}

// Helper functions

func getFloatValue(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		// Try to parse string as float
		if strings.TrimSpace(val) == "" {
			return 0, false
		}
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err == nil
	default:
		return 0, false
	}
}

func getIntValue(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case string:
		// Try to parse string as int
		if strings.TrimSpace(val) == "" {
			return 0, false
		}
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		return i, err == nil
	default:
		return 0, false
	}
}
