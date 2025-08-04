package executor

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("%w: %w", ErrCommandBuildFailed, err)
	}

	// Combine subcommands and args
	fullArgs := append(toolCmd.SubCommands, args...)

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
		SubCommands:   []string{"import"},
		BuildArgs:     b.buildImportCSVArgs,
		RequiresFiles: true,
		Timeout:       30 * time.Second,
	}

	b.toolCommands["import_validate"] = &ToolCommand{
		Tool:          "import_validate",
		Command:       b.cliPath,
		SubCommands:   []string{"import", "--validate"},
		BuildArgs:     b.buildImportValidateArgs,
		RequiresFiles: true,
		Timeout:       10 * time.Second,
	}

	b.toolCommands["import_preview"] = &ToolCommand{
		Tool:          "import_preview",
		Command:       b.cliPath,
		SubCommands:   []string{"import", "--preview"},
		BuildArgs:     b.buildImportPreviewArgs,
		RequiresFiles: true,
		ExpectJSON:    true,
		Timeout:       10 * time.Second,
	}

	// Generation tools
	b.toolCommands["generate_html"] = &ToolCommand{
		Tool:           "generate_html",
		Command:        b.cliPath,
		SubCommands:    []string{"generate"},
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

func (b *CLIBridge) buildInvoiceCreateArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: client_name
	clientName, ok := input["client_name"].(string)
	if !ok || clientName == "" {
		return nil, fmt.Errorf("%w: client_name", ErrMissingRequired)
	}
	args = append(args, "--client", clientName)

	// Optional parameters
	if projectName, ok := input["project_name"].(string); ok && projectName != "" {
		args = append(args, "--project", projectName)
	}
	if dueDate, ok := input["due_date"].(string); ok && dueDate != "" {
		args = append(args, "--due", dueDate)
	}
	if interactive, ok := input["interactive"].(bool); ok && interactive {
		args = append(args, "--interactive")
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceListArgs(input map[string]interface{}) ([]string, error) {
	var args []string
	args = append(args, "--json") // Always output JSON for MCP

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
	var args []string

	// Required: invoice_id
	invoiceID, ok := input["invoice_id"].(string)
	if !ok || invoiceID == "" {
		return nil, fmt.Errorf("%w: invoice_id", ErrMissingRequired)
	}
	args = append(args, invoiceID, "--json")

	return args, nil
}

func (b *CLIBridge) buildInvoiceUpdateArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: invoice_id
	invoiceID, ok := input["invoice_id"].(string)
	if !ok || invoiceID == "" {
		return nil, fmt.Errorf("%w: invoice_id", ErrMissingRequired)
	}
	args = append(args, invoiceID)

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
	if notes, ok := input["notes"].(string); ok && notes != "" {
		args = append(args, "--notes", notes)
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, ErrMissingUpdateFields
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceDeleteArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: invoice_id
	invoiceID, ok := input["invoice_id"].(string)
	if !ok || invoiceID == "" {
		return nil, fmt.Errorf("%w: invoice_id", ErrMissingRequired)
	}
	args = append(args, invoiceID)

	// Optional: force
	if force, ok := input["force"].(bool); ok && force {
		args = append(args, "--force")
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceAddItemArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: invoice_id
	invoiceID, ok := input["invoice_id"].(string)
	if !ok || invoiceID == "" {
		return nil, fmt.Errorf("%w: invoice_id", ErrMissingRequired)
	}
	args = append(args, invoiceID)

	// Required: description
	description, ok := input["description"].(string)
	if !ok || description == "" {
		return nil, fmt.Errorf("%w: description", ErrMissingRequired)
	}
	args = append(args, "--description", description)

	// Required: hours
	hours, ok := getFloatValue(input["hours"])
	if !ok {
		return nil, fmt.Errorf("%w: hours", ErrMissingRequired)
	}
	args = append(args, "--hours", fmt.Sprintf("%.2f", hours))

	// Required: rate
	rate, ok := getFloatValue(input["rate"])
	if !ok {
		return nil, fmt.Errorf("%w: rate", ErrMissingRequired)
	}
	args = append(args, "--rate", fmt.Sprintf("%.2f", rate))

	// Optional: date
	if date, ok := input["date"].(string); ok && date != "" {
		args = append(args, "--date", date)
	}

	return args, nil
}

func (b *CLIBridge) buildInvoiceRemoveItemArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: invoice_id
	invoiceID, ok := input["invoice_id"].(string)
	if !ok || invoiceID == "" {
		return nil, fmt.Errorf("%w: invoice_id", ErrMissingRequired)
	}
	args = append(args, invoiceID)

	// Required: item_id or item_index
	if itemID, ok := input["item_id"].(string); ok && itemID != "" {
		args = append(args, "--item-id", itemID)
	} else if itemIndex, ok := getIntValue(input["item_index"]); ok {
		args = append(args, "--index", fmt.Sprintf("%d", itemIndex))
	} else {
		return nil, ErrMissingItemIdentifier
	}

	return args, nil
}

func (b *CLIBridge) buildClientCreateArgs(input map[string]interface{}) ([]string, error) {
	var args []string

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
	var args []string
	args = append(args, "--json") // Always output JSON for MCP

	// Optional filters
	if active, ok := input["active"].(bool); ok {
		if active {
			args = append(args, "--active")
		} else {
			args = append(args, "--inactive")
		}
	}
	if search, ok := input["search"].(string); ok && search != "" {
		args = append(args, "--search", search)
	}

	return args, nil
}

func (b *CLIBridge) buildClientShowArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: client identifier (name, email, or id)
	if clientID, ok := input["client_id"].(string); ok && clientID != "" {
		args = append(args, clientID, "--json")
	} else if clientName, ok := input["client_name"].(string); ok && clientName != "" {
		args = append(args, clientName, "--json")
	} else if clientEmail, ok := input["client_email"].(string); ok && clientEmail != "" {
		args = append(args, clientEmail, "--json")
	} else {
		return nil, ErrMissingClientIdentifier
	}

	return args, nil
}

func (b *CLIBridge) buildClientUpdateArgs(input map[string]interface{}) ([]string, error) {
	var args []string

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
	var args []string

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
	var args []string

	// Required: file_path
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("%w: file_path", ErrMissingRequired)
	}
	args = append(args, filePath)

	// Optional: invoice_id or create new
	if invoiceID, ok := input["invoice_id"].(string); ok && invoiceID != "" {
		args = append(args, "--invoice", invoiceID)
	} else if createNew, ok := input["create_new"].(bool); ok && createNew {
		// Client name required for new invoice
		if clientName, ok := input["client_name"].(string); ok && clientName != "" {
			args = append(args, "--new-invoice", "--client", clientName)
		} else {
			return nil, ErrMissingClientNameForCreate
		}
	}

	// Optional: append mode
	if appendMode, ok := input["append"].(bool); ok && appendMode {
		args = append(args, "--append")
	}

	// Optional: dry run
	if dryRun, ok := input["dry_run"].(bool); ok && dryRun {
		args = append(args, "--dry-run")
	}

	return args, nil
}

func (b *CLIBridge) buildImportValidateArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: file_path
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("%w: file_path", ErrMissingRequired)
	}
	args = append(args, filePath)

	// Optional: strict validation
	if strict, ok := input["strict"].(bool); ok && strict {
		args = append(args, "--strict")
	}

	return args, nil
}

func (b *CLIBridge) buildImportPreviewArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Required: file_path
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("%w: file_path", ErrMissingRequired)
	}
	args = append(args, filePath, "--json")

	// Optional: limit rows
	if limit, ok := getIntValue(input["limit"]); ok {
		args = append(args, "--limit", fmt.Sprintf("%d", limit))
	}

	return args, nil
}

func (b *CLIBridge) buildGenerateHTMLArgs(input map[string]interface{}) ([]string, error) {
	var args []string

	// Optional: specific invoice IDs or all
	if invoiceIDs, ok := input["invoice_ids"].([]interface{}); ok && len(invoiceIDs) > 0 {
		for _, id := range invoiceIDs {
			if idStr, ok := id.(string); ok {
				args = append(args, idStr)
			}
		}
	} else if invoiceID, ok := input["invoice_id"].(string); ok && invoiceID != "" {
		args = append(args, invoiceID)
	} else {
		// Generate all invoices
		args = append(args, "--all")
	}

	// Optional: template
	if template, ok := input["template"].(string); ok && template != "" {
		args = append(args, "--template", template)
	}

	// Optional: output directory
	if outputDir, ok := input["output_dir"].(string); ok && outputDir != "" {
		args = append(args, "--output", outputDir)
	}

	return args, nil
}

func (b *CLIBridge) buildGenerateSummaryArgs(input map[string]interface{}) ([]string, error) {
	var args []string
	args = append(args, "--json") // Output JSON for MCP

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
	args = append(args, "--json") // Always output JSON for MCP

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
