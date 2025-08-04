# go-invoice MCP Integration - Implementation Plan

## Executive Summary

go-invoice MCP Integration extends the existing CLI-driven invoice management tool with Model Context Protocol (MCP) server capabilities, enabling seamless integration with Claude Desktop and Claude Code. This enhancement allows users to manage their entire invoicing workflow through natural language conversations while maintaining all the robustness and features of the existing CLI application.

**Key Architecture Decisions**:
- **MCP Server Separation**: Standalone MCP server binary separate from core go-invoice CLI
- **CLI Command Bridge**: Direct integration with existing CLI commands through process execution
- **Protocol Compliance**: Full MCP specification adherence for maximum compatibility
- **Tool-Based Interface**: Each CLI command exposed as individual MCP tools with proper schemas
- **Security First**: Input validation, sandboxing, and secure command execution
- **Configuration Integration**: Seamless integration with existing .env.config system
- **Context Awareness**: Full context.Context support throughout MCP operations

## Vision Statement

go-invoice MCP Integration embodies the principle of extending proven functionality without disruption. It bridges the gap between traditional CLI workflows and modern AI-assisted interfaces, providing users with the choice to manage invoices through either command-line operations or natural language conversations with Claude. The integration prioritizes:

- **Non-Disruptive Enhancement**: MCP server operates alongside existing CLI without modification
- **Natural Language Interface**: Convert complex CLI operations into conversational interactions
- **Workflow Continuity**: Existing CLI users can adopt MCP gradually at their own pace
- **AI-Powered Efficiency**: Leverage Claude's capabilities for intelligent invoice management
- **Unified Data Access**: Both CLI and MCP interfaces work with the same underlying data
- **Developer Experience**: Maintain the developer-first approach with enhanced accessibility

## System Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Claude User   │────▶│   MCP Server    │────▶│  go-invoice CLI │
│  (Natural Lang) │     │  (Protocol)     │     │  (Commands)     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                │                         │
                                ▼                         ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ Claude Desktop  │     │  Tool Schemas   │     │  JSON Storage   │
│  Integration    │     │   (MCP Tools)   │     │   (Invoices)    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                          │
                                                          ▼
                                                ┌─────────────────┐
                                                │ Business Data   │
                                                │ (Shared State)  │
                                                └─────────────────┘
```

The MCP server acts as a protocol bridge, translating natural language requests through Claude into structured CLI command executions, while maintaining complete data consistency and security.

## Implementation Roadmap

### Phase 0: Foundation Alignment (.github/AGENTS.md Compliance)
**Objective**: Ensure MCP implementation plan fully aligns with established conventions and standards

**Duration**: 30 minutes

**Implementation Agent**: Use Claude Code with go-expert-developer persona

**Key Alignment Areas:**
1. **Context-First Design**: All MCP operations must accept `context.Context` as first parameter
2. **Interface Philosophy**: Follow "accept interfaces, return concrete types" pattern for MCP handlers
3. **Error Handling Excellence**: Implement comprehensive error wrapping and context for MCP responses
4. **Testing Standards**: Use testify suite with table-driven tests and descriptive names
5. **Security Integration**: Include vulnerability scanning and dependency verification
6. **No Global State**: Enforce dependency injection patterns throughout MCP server

**Enhanced Architecture Principles:**
- Context flows through entire MCP call stack for cancellation and timeout support
- Consumer-driven interface design with minimal, focused MCP tool contracts
- Comprehensive error handling with actionable messages and proper wrapping
- Dependency injection eliminates global state and improves testability
- Security-first approach with automated vulnerability scanning

**Verification Steps:**
```bash
# Enhanced security and quality validation
govulncheck ./...
go mod verify
golangci-lint run
go test -race ./...
go test -cover ./...
go vet ./...
```

**Success Criteria:**
- ✅ All MCP function signatures include context.Context as first parameter
- ✅ MCP interfaces defined at point of use (consumer-driven design)
- ✅ Error messages provide clear context and actionable guidance
- ✅ Test coverage exceeds 90% using testify patterns
- ✅ No security vulnerabilities detected in dependencies
- ✅ All linting passes per .github/AGENTS.md standards
- ✅ Dependency injection used throughout (no global state)

---

### Phase 1: MCP Server Foundation and Protocol Implementation
**Objective**: Establish MCP server foundation with protocol compliance and CLI bridge interface

**Implementation Agent**: Use Claude Code with go-expert-developer persona for all Go code implementation

**Implementation Steps:**
1. Create standalone MCP server binary with protocol implementation
2. Implement CLI command bridge for safe command execution
3. Set up MCP protocol message handling with context support
4. Create configuration management for MCP server settings
5. Implement logging and error handling infrastructure

**Files to Create/Modify:**
- `cmd/go-invoice-mcp/main.go` - MCP server main entry point
- `internal/mcp/server.go` - MCP protocol server implementation
- `internal/mcp/bridge.go` - CLI command execution bridge
- `internal/mcp/config.go` - MCP server configuration
- `internal/mcp/handlers.go` - MCP message handlers
- `go.mod` - Update dependencies for MCP protocol support

**MCP Server Structure:**
```go
type MCPServer struct {
    logger     Logger
    bridge     CLIBridge
    config     *MCPConfig
    validator  RequestValidator
}

type CLIBridge interface {
    ExecuteCommand(ctx context.Context, cmd CommandRequest) (*CommandResponse, error)
    ValidateCommand(ctx context.Context, cmd string, args []string) error
}

type MCPConfig struct {
    Port        int
    Host        string
    CLIPath     string
    MaxTimeout  time.Duration
    LogLevel    string
}

// MCP server demonstrates context-first design and dependency injection
func NewMCPServer(logger Logger, bridge CLIBridge, config *MCPConfig, validator RequestValidator) *MCPServer {
    return &MCPServer{
        logger:    logger,
        bridge:    bridge,
        config:    config,
        validator: validator,
    }
}

func (s *MCPServer) HandleToolCall(ctx context.Context, req *MCPToolRequest) (*MCPToolResponse, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    if err := s.validator.ValidateRequest(ctx, req); err != nil {
        return nil, fmt.Errorf("invalid MCP request: %w", err)
    }
    
    cmdReq, err := s.convertToCommandRequest(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to convert MCP request: %w", err)
    }
    
    cmdResp, err := s.bridge.ExecuteCommand(ctx, cmdReq)
    if err != nil {
        return nil, fmt.Errorf("command execution failed: %w", err)
    }
    
    mcpResp := s.convertToMCPResponse(cmdResp)
    s.logger.Info("MCP tool executed successfully", "tool", req.Tool, "duration", cmdResp.Duration)
    
    return mcpResp, nil
}
```

**Verification Steps (Enhanced .github/AGENTS.md Compliance):**
```bash
# 1. Comprehensive security scanning
govulncheck ./...
go mod verify
gitleaks detect --source . --log-opts="--all" --verbose

# 2. Complete code quality validation
golangci-lint run
go vet ./...
gofumpt -l .
goimports -l .

# 3. Build MCP server with optimization
go build -ldflags="-s -w" -o go-invoice-mcp ./cmd/go-invoice-mcp

# 4. Comprehensive testing with coverage validation
go test -v -race -cover ./internal/mcp/...
go test -cover ./internal/mcp/... | grep -E "coverage: [0-9]+" | awk '{if ($2 < 90) exit 1}'

# 5. Context cancellation validation
timeout 5s go test -run TestMCPContextCancellation ./internal/mcp/...

# 6. Performance validation
go test -bench=. ./internal/mcp/...

# 7. Test MCP server startup with context
./go-invoice-mcp --config mcp-config.json --validate

# 8. Test CLI bridge functionality with security
./go-invoice-mcp --test-bridge --validate-security

# 9. Verify dependency injection (no global state)
grep -r "var.*=" internal/mcp/ | grep -v test | grep -v const | wc -l | awk '{if ($1 > 0) exit 1}'
grep -r "func init()" internal/mcp/ | wc -l | awk '{if ($1 > 0) exit 1}'
```

**Success Criteria (Measurable .github/AGENTS.md Compliance):**
- ✅ MCP server builds successfully with zero compilation errors and starts in <500ms
- ✅ Protocol implementation handles MCP messages with 100% context.Context parameter compliance
- ✅ CLI bridge executes commands with comprehensive input validation (zero injection vulnerabilities)
- ✅ Configuration loads with structured error messages including file path and line number
- ✅ Logging provides comprehensive debugging with operation context and execution duration
- ✅ 100% of MCP functions accept context.Context as first parameter (verified with grep)
- ✅ Zero global variables or init functions detected (verified with automated checks)
- ✅ Error handling uses fmt.Errorf wrapping pattern with operation context in 100% of cases
- ✅ Tests use testify.suite.Suite pattern with descriptive names (TestComponentOperationCondition format)
- ✅ Test coverage ≥ 90% for all MCP packages (verified with coverage threshold check)
- ✅ Zero race conditions detected (verified with -race flag)
- ✅ Zero security vulnerabilities in dependencies (verified with govulncheck)
- ✅ Zero secrets detected in code (verified with gitleaks)
- ✅ Zero linting violations (verified with golangci-lint, go vet, gofumpt, goimports)
- ✅ MCP server startup time <500ms and tool execution <2s average (verified with benchmarks)
- ✅ Context cancellation response time <100ms (verified with timeout tests)
- ✅ Final todo: Update the @plans/plan-02-status.md file with the results of the implementation

### Phase 2: MCP Tool Definitions and Schema Implementation
**Objective**: Define comprehensive MCP tools for all go-invoice CLI commands with proper schemas

**Implementation Agent**: Use Claude Code with go-expert-developer persona for Go implementation

**Implementation Steps:**
1. Create MCP tool schema definitions for all CLI commands
2. Implement tool parameter validation and type conversion
3. Add comprehensive help text and examples for each tool
4. Create tool discovery and listing functionality
5. Implement tool grouping and categorization

**Files to Create/Modify:**
- `internal/mcp/tools/` - MCP tool definitions directory
  - `invoice_tools.go` - Invoice management tools
  - `import_tools.go` - CSV import tools
  - `generate_tools.go` - Invoice generation tools
  - `config_tools.go` - Configuration management tools
  - `client_tools.go` - Client management tools
- `internal/mcp/schemas/` - JSON schema definitions
- `internal/mcp/validation.go` - Tool parameter validation
- `cmd/go-invoice-mcp/tools.json` - Tool registry for Claude Desktop

**MCP Tool Definitions:**
```go
type MCPTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
    Examples    []MCPToolExample       `json:"examples,omitempty"`
    Category    string                 `json:"category"`
    CLICommand  string                 `json:"cliCommand"`
    CLIArgs     []string               `json:"cliArgs"`
}

type MCPToolExample struct {
    Description string                 `json:"description"`
    Input       map[string]interface{} `json:"input"`
}

// Invoice management tools
var InvoiceCreateTool = MCPTool{
    Name:        "invoice_create",
    Description: "Create a new invoice with client information and metadata",
    Category:    "invoice_management",
    CLICommand:  "go-invoice",
    CLIArgs:     []string{"invoice", "create"},
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "client_name": map[string]interface{}{
                "type":        "string",
                "description": "Name of the client for this invoice",
                "minLength":   1,
            },
            "project_name": map[string]interface{}{
                "type":        "string",
                "description": "Optional project name for organization",
            },
            "due_date": map[string]interface{}{
                "type":        "string",
                "format":      "date",
                "description": "Invoice due date (YYYY-MM-DD format)",
            },
            "interactive": map[string]interface{}{
                "type":        "boolean",
                "description": "Use interactive mode for additional input",
                "default":     false,
            },
        },
        "required": []string{"client_name"},
    },
    Examples: []MCPToolExample{
        {
            Description: "Create invoice for Acme Corp",
            Input: map[string]interface{}{
                "client_name": "Acme Corp",
                "project_name": "Website Redesign",
                "due_date": "2025-09-01",
            },
        },
    },
}

var CSVImportTool = MCPTool{
    Name:        "import_csv",
    Description: "Import work hours from CSV timesheet into an invoice",
    Category:    "data_import",
    CLICommand:  "go-invoice",
    CLIArgs:     []string{"import"},
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "file_path": map[string]interface{}{
                "type":        "string",
                "description": "Path to CSV file containing timesheet data",
                "pattern":     ".*\\.csv$",
            },
            "invoice_id": map[string]interface{}{
                "type":        "string",
                "description": "Invoice ID to import hours into",
                "minLength":   1,
            },
            "append": map[string]interface{}{
                "type":        "boolean",
                "description": "Append to existing work items instead of replacing",
                "default":     false,
            },
            "dry_run": map[string]interface{}{
                "type":        "boolean",
                "description": "Validate CSV without making changes",
                "default":     false,
            },
        },
        "required": []string{"file_path", "invoice_id"},
    },
}

// Tool registry with consumer-driven interface
type ToolRegistry interface {
    GetTool(ctx context.Context, name string) (*MCPTool, error)
    ListTools(ctx context.Context, category string) ([]*MCPTool, error)
    ValidateToolInput(ctx context.Context, toolName string, input map[string]interface{}) error
}

type DefaultToolRegistry struct {
    tools     map[string]*MCPTool
    validator InputValidator
    logger    Logger
}

func NewToolRegistry(validator InputValidator, logger Logger) *DefaultToolRegistry {
    registry := &DefaultToolRegistry{
        tools:     make(map[string]*MCPTool),
        validator: validator,
        logger:    logger,
    }
    
    // Register all tools
    registry.registerTool(&InvoiceCreateTool)
    registry.registerTool(&CSVImportTool)
    // ... register other tools
    
    return registry
}

func (r *DefaultToolRegistry) ValidateToolInput(ctx context.Context, toolName string, input map[string]interface{}) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    tool, exists := r.tools[toolName]
    if !exists {
        return fmt.Errorf("tool not found: %s", toolName)
    }
    
    if err := r.validator.ValidateAgainstSchema(ctx, input, tool.InputSchema); err != nil {
        return fmt.Errorf("input validation failed for tool %s: %w", toolName, err)
    }
    
    return nil
}
```

**Verification Steps:**
```bash
# 1. Run security and quality checks
govulncheck ./...
go mod verify
golangci-lint run

# 2. Run comprehensive tests with testify
go test -v -race -cover ./internal/mcp/tools/...
go test -v -race -cover ./internal/mcp/schemas/...

# 3. Test tool discovery
./go-invoice-mcp --list-tools

# 4. Test tool schema validation
./go-invoice-mcp --validate-tool invoice_create --input '{"client_name":"Test"}'

# 5. Test all tool categories
./go-invoice-mcp --list-tools --category invoice_management

# 6. Test tool execution with context
./go-invoice-mcp --execute-tool invoice_create --input '{"client_name":"Acme Corp"}'

# 7. Test context cancellation on tool operations
timeout 2s ./go-invoice-mcp --execute-tool import_csv --input '{"file_path":"large.csv"}'
```

**Success Criteria:**
- ✅ All CLI commands have corresponding MCP tools with proper schemas
- ✅ Tool parameter validation works correctly with clear error messages
- ✅ Tool discovery and listing functions properly with categorization
- ✅ Schema validation catches invalid inputs with actionable feedback
- ✅ Help text and examples provide clear usage guidance
- ✅ All tool operations accept context.Context for cancellation support
- ✅ Consumer-driven interfaces defined at point of use
- ✅ Dependency injection used throughout (no global state)
- ✅ Tests use testify suite with table-driven patterns
- ✅ Context cancellation works correctly for tool operations
- ✅ No security vulnerabilities in tool processing dependencies
- ✅ All linting and race condition checks pass
- ✅ Final todo: Update the @plans/plan-02-status.md file with the results of the implementation

### Phase 3: Command Execution and Response Processing
**Objective**: Implement secure command execution bridge and response processing system

**Implementation Agent**: Use Claude Code with go-expert-developer persona

**Implementation Steps:**
1. Implement secure CLI command execution with sandboxing
2. Create command output parsing and response formatting
3. Add error handling and recovery for command failures
4. Implement file handling for CSV imports and HTML generation
5. Add progress reporting for long-running operations

**Files to Create/Modify:**
- `internal/mcp/executor/` - Command execution engine
  - `bridge.go` - CLI command execution bridge
  - `parser.go` - Command output parsing
  - `security.go` - Command validation and sandboxing
  - `files.go` - File handling for imports and exports
- `internal/mcp/responses/` - Response formatting
- `internal/mcp/progress.go` - Progress reporting for long operations
- `cmd/go-invoice-mcp/sandbox.json` - Command execution sandbox config

**Command Execution System:**
```go
type CommandExecutor interface {
    Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error)
    ValidateCommand(ctx context.Context, command string, args []string) error
    GetAllowedCommands(ctx context.Context) ([]string, error)
}

type ExecutionRequest struct {
    Command     string            `json:"command"`
    Args        []string          `json:"args"`
    WorkingDir  string            `json:"workingDir,omitempty"`
    Environment map[string]string `json:"environment,omitempty"`
    Timeout     time.Duration     `json:"timeout,omitempty"`
    InputFiles  []FileReference   `json:"inputFiles,omitempty"`
}

type ExecutionResponse struct {
    ExitCode    int               `json:"exitCode"`
    Stdout      string            `json:"stdout"`
    Stderr      string            `json:"stderr"`
    Duration    time.Duration     `json:"duration"`
    OutputFiles []FileReference   `json:"outputFiles,omitempty"`
    Error       string            `json:"error,omitempty"`
}

type FileReference struct {
    Path        string `json:"path"`
    ContentType string `json:"contentType"`
    Size        int64  `json:"size"`
    Checksum    string `json:"checksum,omitempty"`
}

type SecureExecutor struct {
    logger       Logger
    validator    CommandValidator
    sandbox      SandboxConfig
    fileHandler  FileHandler
    allowedCmds  map[string]bool
}

func NewSecureExecutor(logger Logger, validator CommandValidator, sandbox SandboxConfig, fileHandler FileHandler) *SecureExecutor {
    return &SecureExecutor{
        logger:      logger,
        validator:   validator,
        sandbox:     sandbox,
        fileHandler: fileHandler,
        allowedCmds: initAllowedCommands(),
    }
}

func (e *SecureExecutor) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Validate command is allowed
    if err := e.ValidateCommand(ctx, req.Command, req.Args); err != nil {
        return nil, fmt.Errorf("command validation failed: %w", err)
    }
    
    // Prepare execution environment
    execCtx, cancel := context.WithTimeout(ctx, req.Timeout)
    defer cancel()
    
    // Handle input files
    workDir, cleanup, err := e.fileHandler.PrepareWorkspace(execCtx, req.InputFiles)
    if err != nil {
        return nil, fmt.Errorf("workspace preparation failed: %w", err)
    }
    defer cleanup()
    
    // Execute command with sandboxing
    cmd := exec.CommandContext(execCtx, req.Command, req.Args...)
    cmd.Dir = workDir
    cmd.Env = e.buildEnvironment(req.Environment)
    
    start := time.Now()
    stdout, stderr, err := e.executeWithCapture(cmd)
    duration := time.Since(start)
    
    response := &ExecutionResponse{
        ExitCode: e.getExitCode(err),
        Stdout:   string(stdout),
        Stderr:   string(stderr),
        Duration: duration,
    }
    
    if err != nil {
        response.Error = err.Error()
        e.logger.Error("command execution failed", "command", req.Command, "error", err)
    } else {
        e.logger.Info("command executed successfully", "command", req.Command, "duration", duration)
    }
    
    // Handle output files
    outputFiles, err := e.fileHandler.CollectOutputFiles(execCtx, workDir)
    if err != nil {
        e.logger.Warn("failed to collect output files", "error", err)
    } else {
        response.OutputFiles = outputFiles
    }
    
    return response, nil
}

// Response processor interface defined at point of use
type ResponseProcessor interface {
    ProcessCommandOutput(ctx context.Context, resp *ExecutionResponse) (*MCPToolResponse, error)
    FormatErrorResponse(ctx context.Context, err error) (*MCPToolResponse, error)
}

type DefaultResponseProcessor struct {
    logger Logger
    parser OutputParser
}

func (p *DefaultResponseProcessor) ProcessCommandOutput(ctx context.Context, resp *ExecutionResponse) (*MCPToolResponse, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    if resp.ExitCode != 0 {
        return p.FormatErrorResponse(ctx, fmt.Errorf("command failed with exit code %d: %s", resp.ExitCode, resp.Stderr))
    }
    
    // Parse structured output from CLI
    content, err := p.parser.ParseOutput(ctx, resp.Stdout, resp.OutputFiles)
    if err != nil {
        return nil, fmt.Errorf("failed to parse command output: %w", err)
    }
    
    mcpResp := &MCPToolResponse{
        Content: []MCPContent{
            {
                Type: "text",
                Text: content.Summary,
            },
        },
        IsError: false,
    }
    
    // Add file attachments if present
    for _, file := range resp.OutputFiles {
        mcpResp.Content = append(mcpResp.Content, MCPContent{
            Type:     "resource",
            Resource: file.Path,
            MimeType: file.ContentType,
        })
    }
    
    return mcpResp, nil
}
```

**Verification Steps:**
```bash
# 1. Run security and quality checks
govulncheck ./...
go mod verify
golangci-lint run

# 2. Run comprehensive tests with testify
go test -v -race -cover ./internal/mcp/executor/...
go test -v -race -cover ./internal/mcp/responses/...

# 3. Test command execution with context
./go-invoice-mcp --test-execution invoice create --client "Test Client"

# 4. Test file handling for CSV import
./go-invoice-mcp --test-file-handling import --file examples/timesheet.csv

# 5. Test command validation and sandboxing
./go-invoice-mcp --test-security-validation

# 6. Test error handling and recovery
./go-invoice-mcp --test-error-handling

# 7. Test context cancellation on long operations
timeout 3s ./go-invoice-mcp --test-long-operation
```

**Success Criteria:**
- ✅ Command execution works securely with proper validation and context support
- ✅ File handling supports CSV imports and HTML generation with validation
- ✅ Error handling provides clear, actionable feedback with proper wrapping
- ✅ Response processing formats output correctly for MCP consumption
- ✅ Sandboxing prevents unauthorized command execution
- ✅ All execution operations accept context.Context for cancellation support
- ✅ Consumer-driven interfaces used for executor abstraction
- ✅ Dependency injection eliminates global state
- ✅ Tests use testify suite with comprehensive security testing
- ✅ Context cancellation works for command execution
- ✅ Progress reporting works for long-running operations
- ✅ No security vulnerabilities in execution dependencies
- ✅ Race condition testing passes for concurrent executions
- ✅ Final todo: Update the @plans/plan-02-status.md file with the results of the implementation

### Phase 4: Claude Desktop and Claude Code Integration
**Objective**: Create comprehensive integration configuration and setup for both Claude Desktop and Claude Code platforms

**Implementation Agent**: Use Claude Code for integration setup and documentation

**Implementation Steps:**
1. Create Claude Desktop MCP configuration files (HTTP transport)
2. Create Claude Code MCP configuration files (stdio transport)
3. Implement dual-transport MCP server with auto-detection
4. Add connection testing and health checks for both platforms
5. Create setup documentation and troubleshooting guide for local-only operation
6. Implement logging and monitoring for both Claude platforms

**Files to Create/Modify:**
- `configs/claude-desktop/` - Claude Desktop integration files
  - `mcp_servers.json` - MCP server configuration for Claude Desktop (HTTP)
  - `tools_config.json` - Tool-specific configuration
- `configs/claude-code/` - Claude Code integration files
  - `mcp_config.json` - MCP server configuration for Claude Code (stdio)
  - `project_config.json` - Project-scope configuration template
- `scripts/setup-claude-integration.sh` - Unified setup script for both platforms
- `scripts/setup-claude-code-integration.sh` - Claude Code specific setup
- `docs/claude-integration.md` - Comprehensive integration documentation
- `internal/mcp/health.go` - Health check and monitoring
- `internal/mcp/transport.go` - Dual transport support (stdio + HTTP)
- `configs/logging.yaml` - Logging configuration for Claude interactions

**Claude Desktop Integration (HTTP Transport):**
```json
// configs/claude-desktop/mcp_servers.json
{
  "mcpServers": {
    "go-invoice": {
      "command": "go-invoice-mcp",
      "args": ["--http", "--config", "~/.go-invoice/mcp-config.json"],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "info",
        "MCP_TRANSPORT": "http"
      }
    }
  }
}
```

**Claude Code Integration (stdio Transport):**
```json
// configs/claude-code/mcp_config.json
{
  "mcpServers": {
    "go-invoice": {
      "command": "go-invoice-mcp",
      "args": ["--stdio"],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "info",
        "MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

**Project-Scope Claude Code Configuration:**
```json
// .claude_config.json (project root)
{
  "mcp": {
    "servers": {
      "go-invoice": {
        "command": "./bin/go-invoice-mcp",
        "args": ["--stdio", "--config", "./invoice-config.json"],
        "workingDirectory": ".",
        "env": {
          "GO_INVOICE_PROJECT": "./",
          "MCP_TRANSPORT": "stdio"
        }
      }
    }
  }
}
```

**Dual Transport MCP Server:**
```go
// Transport detection and configuration
type TransportConfig struct {
    Type    string `json:"type"`    // "stdio" | "http" 
    Host    string `json:"host,omitempty"`
    Port    int    `json:"port,omitempty"`
    Timeout string `json:"timeout"`
}

type MCPServer struct {
    logger     Logger
    bridge     CLIBridge
    config     *MCPConfig
    transport  Transport
    validator  RequestValidator
}

func NewMCPServer(logger Logger, bridge CLIBridge, config *MCPConfig) *MCPServer {
    // Auto-detect transport based on environment/args
    transport := detectTransport()
    
    return &MCPServer{
        logger:    logger,
        bridge:    bridge,
        config:    config,
        transport: transport,
        validator: NewRequestValidator(),
    }
}

func detectTransport() Transport {
    // Check command line args or environment
    if hasStdioArgs() || os.Getenv("MCP_TRANSPORT") == "stdio" {
        return NewStdioTransport()
    }
    return NewHTTPTransport()
}
```

```go
// Health check and monitoring system
type HealthChecker interface {
    CheckHealth(ctx context.Context) (*HealthStatus, error)
    StartHealthMonitoring(ctx context.Context, interval time.Duration) error
}

type HealthStatus struct {
    Status      string            `json:"status"`
    Version     string            `json:"version"`
    Uptime      time.Duration     `json:"uptime"`
    CLIStatus   string            `json:"cliStatus"`
    StorageOK   bool              `json:"storageOk"`
    Checks      []HealthCheck     `json:"checks"`
    LastError   string            `json:"lastError,omitempty"`
    Metadata    map[string]string `json:"metadata"`
}

type HealthCheck struct {
    Name     string        `json:"name"`
    Status   string        `json:"status"`
    Duration time.Duration `json:"duration"`
    Message  string        `json:"message,omitempty"`
}

type DefaultHealthChecker struct {
    logger     Logger
    executor   CommandExecutor
    storage    StorageChecker
    startTime  time.Time
}

func NewHealthChecker(logger Logger, executor CommandExecutor, storage StorageChecker) *DefaultHealthChecker {
    return &DefaultHealthChecker{
        logger:    logger,
        executor:  executor,
        storage:   storage,
        startTime: time.Now(),
    }
}

func (h *DefaultHealthChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    status := &HealthStatus{
        Status:    "unknown",
        Version:   buildinfo.Version,
        Uptime:    time.Since(h.startTime),
        Checks:    []HealthCheck{},
        Metadata:  make(map[string]string),
    }
    
    // Check CLI availability
    cliCheck := h.checkCLIAvailability(ctx)
    status.Checks = append(status.Checks, cliCheck)
    status.CLIStatus = cliCheck.Status
    
    // Check storage accessibility
    storageCheck := h.checkStorageHealth(ctx)
    status.Checks = append(status.Checks, storageCheck)
    status.StorageOK = storageCheck.Status == "healthy"
    
    // Determine overall status
    allHealthy := true
    for _, check := range status.Checks {
        if check.Status != "healthy" {
            allHealthy = false
            if status.LastError == "" {
                status.LastError = check.Message
            }
        }
    }
    
    if allHealthy {
        status.Status = "healthy"
    } else {
        status.Status = "unhealthy"
    }
    
    h.logger.Info("health check completed", "status", status.Status, "checks", len(status.Checks))
    return status, nil
}

// Connection manager for Claude Desktop integration
type ConnectionManager interface {
    EstablishConnection(ctx context.Context) error
    HandleClaudeRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error)
    CloseConnection(ctx context.Context) error
}

type DefaultConnectionManager struct {
    logger       Logger
    toolRegistry ToolRegistry
    executor     CommandExecutor
    health       HealthChecker
    config       *MCPConfig
}

func (c *DefaultConnectionManager) HandleClaudeRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    c.logger.Info("handling Claude request", "method", req.Method, "tool", req.Params.Name)
    
    switch req.Method {
    case "tools/list":
        return c.handleToolsList(ctx, req)
    case "tools/call":
        return c.handleToolCall(ctx, req)
    case "ping":
        return c.handlePing(ctx, req)
    default:
        return nil, fmt.Errorf("unsupported method: %s", req.Method)
    }
}
```

**Unified Setup Script:**
```bash
#!/bin/bash
# scripts/setup-claude-integration.sh

set -euo pipefail

PLATFORM="${1:-both}"  # both, desktop, code

echo "Setting up go-invoice MCP integration for: $PLATFORM"

# Check if go-invoice-mcp is built
if ! command -v go-invoice-mcp &> /dev/null; then
    echo "Building go-invoice-mcp..."
    go build -o ~/bin/go-invoice-mcp ./cmd/go-invoice-mcp
    export PATH="$HOME/bin:$PATH"
fi

# Create MCP configuration directory
MCP_CONFIG_DIR="$HOME/.go-invoice"
mkdir -p "$MCP_CONFIG_DIR"

setup_claude_desktop() {
    echo "Setting up Claude Desktop integration (HTTP transport)..."
    
    # Create Claude Desktop config directory
    mkdir -p "$HOME/.claude"
    
    # Copy Claude Desktop configuration
    cp configs/claude-desktop/mcp_servers.json "$HOME/.claude/"
    
    # Test HTTP transport
    echo "Testing Claude Desktop MCP server..."
    go-invoice-mcp --http --config "$MCP_CONFIG_DIR/mcp-config.json" --test
    
    echo "✅ Claude Desktop setup complete!"
    echo "📝 Restart Claude Desktop to load the new MCP server."
}

setup_claude_code() {
    echo "Setting up Claude Code integration (stdio transport)..."
    
    # Copy Claude Code configuration  
    cp configs/claude-code/mcp_config.json "$MCP_CONFIG_DIR/"
    
    # Test stdio transport
    echo "Testing Claude Code MCP server..."
    echo '{"method":"ping","id":1}' | go-invoice-mcp --stdio --config "$MCP_CONFIG_DIR/mcp-config.json"
    
    echo "✅ Claude Code setup complete!"
    echo "📝 Add the MCP server configuration to your Claude Code settings."
    echo "🔧 Configuration file: ~/.go-invoice/mcp_config.json"
}

# Copy base configuration
cp configs/mcp-config.json "$MCP_CONFIG_DIR/"

# Setup based on platform selection
case "$PLATFORM" in
    "desktop")
        setup_claude_desktop
        ;;
    "code") 
        setup_claude_code
        ;;
    "both"|*)
        setup_claude_desktop
        setup_claude_code
        echo "💬 You can now use natural language to manage invoices in both Claude platforms!"
        ;;
esac

echo "🎉 MCP integration setup complete!"
```

**Claude Code Specific Setup Script:**
```bash
#!/bin/bash
# scripts/setup-claude-code-integration.sh

set -euo pipefail

echo "Setting up go-invoice MCP integration for Claude Code..."

# Build MCP server if needed
if ! command -v go-invoice-mcp &> /dev/null; then
    echo "Building go-invoice-mcp..."
    go build -o ~/bin/go-invoice-mcp ./cmd/go-invoice-mcp
fi

# Create configuration directories
MCP_CONFIG_DIR="$HOME/.go-invoice"
mkdir -p "$MCP_CONFIG_DIR"

# Copy Claude Code specific configuration
cp configs/claude-code/mcp_config.json "$MCP_CONFIG_DIR/"
cp configs/mcp-config.json "$MCP_CONFIG_DIR/"

# Test stdio transport
echo "Testing Claude Code MCP server (stdio)..."
echo '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
    go-invoice-mcp --stdio --config "$MCP_CONFIG_DIR/mcp-config.json"

if [ $? -eq 0 ]; then
    echo "✅ Claude Code MCP server test successful!"
    echo ""
    echo "Next steps:"
    echo "1. Add this server to your Claude Code MCP configuration:"
    echo "   Server config: ~/.go-invoice/mcp_config.json"
    echo "2. Restart Claude Code to load the MCP server"
    echo "3. Test with: /mcp__go_invoice__show_config"
    echo ""
    echo "Available slash commands:"
    echo "  /mcp__go_invoice__create_invoice"
    echo "  /mcp__go_invoice__list_invoices"
    echo "  /mcp__go_invoice__import_csv"
    echo "  /mcp__go_invoice__generate_html"
else
    echo "❌ MCP server test failed. Check configuration and try again."
    exit 1
fi
```

**Verification Steps:**
```bash
# 1. Run security and quality checks (local-only focus)
govulncheck ./...
go mod verify
golangci-lint run

# 2. Run integration tests for both platforms
go test -v -race -cover ./internal/mcp/health/...
go test -v -race -cover ./internal/mcp/transport/...

# 3. Test unified setup script for both platforms
./scripts/setup-claude-integration.sh both

# 4. Test Claude Desktop integration (HTTP transport)
./scripts/setup-claude-integration.sh desktop
go-invoice-mcp --http --health-check

# 5. Test Claude Code integration (stdio transport)  
./scripts/setup-claude-code-integration.sh
echo '{"method":"ping","id":1}' | go-invoice-mcp --stdio

# 6. Test dual transport functionality
go-invoice-mcp --stdio --test-stdio
go-invoice-mcp --http --test-http

# 7. Test local security validation (no OAuth needed)
go-invoice-mcp --validate-security --local-only

# 8. Test connection with both platforms (manual)
# Start Claude Desktop and test MCP tools
# Start Claude Code and test slash commands

# 9. Test logging and monitoring for both transports
go-invoice-mcp --test-monitoring --transport=stdio
go-invoice-mcp --test-monitoring --transport=http

# 10. Test error recovery and reconnection
go-invoice-mcp --test-resilience --stdio
go-invoice-mcp --test-resilience --http
```

**Success Criteria:**
- ✅ Both Claude Desktop and Claude Code integration configure correctly with proper MCP server registration
- ✅ Dual transport support (stdio for Claude Code, HTTP for Claude Desktop) works seamlessly
- ✅ Health checks validate MCP server and CLI availability for both platforms with context support
- ✅ Setup scripts automate integration configuration successfully for both platforms
- ✅ Connection management handles requests reliably from both Claude platforms with error recovery
- ✅ Logging provides comprehensive debugging information for both transport types
- ✅ All integration operations accept context.Context for cancellation support
- ✅ Monitoring tracks performance and connection health for both platforms
- ✅ Error handling provides clear guidance for setup issues on both platforms
- ✅ Documentation covers setup, troubleshooting, and common workflows for both platforms
- ✅ Local-only security validation ensures safe integration (no network dependencies)
- ✅ Prompt injection protection prevents malicious command execution
- ✅ File system sandboxing restricts access to invoice directories only
- ✅ Tests use testify suite with integration testing for both transports
- ✅ No security vulnerabilities in integration dependencies
- ✅ Claude Code slash commands work correctly (/mcp__go_invoice__*)
- ✅ Resource mentions work in Claude Code (@invoice:, @client:, @timesheet:)
- ✅ Project-scope configuration works for Claude Code
- ✅ Final todo: Update the @plans/plan-02-status.md file with the results of the implementation

### Phase 5: Testing and Documentation
**Objective**: Ensure code quality with comprehensive testing and documentation for MCP integration

**Implementation Agent**: Use Claude Code with testing expertise

**Implementation Steps:**
1. Write unit tests for all MCP packages
2. Create integration tests for complete MCP workflows
3. Add example conversations and use cases
4. Write comprehensive MCP integration documentation
5. Create troubleshooting guide and FAQ
6. Update the already good README.md with MCP integration details

**Files to Create/Modify:**
- `*_test.go` - Unit tests for all MCP packages
- `test/mcp_integration_test.go` - MCP integration test suite
- `docs/mcp-integration.md` - Comprehensive MCP documentation
- `docs/claude-workflows.md` - Example workflows and conversations
- `examples/mcp/` - Example MCP configurations and use cases
- `docs/troubleshooting-mcp.md` - MCP troubleshooting guide

**Testing Strategy:**
```go
// Unit test example using testify suite for MCP tools
func TestMCPToolExecution(t *testing.T) {
    tests := []struct {
        name        string
        tool        string
        input       map[string]interface{}
        expected    MCPToolResponse
        expectError bool
    }{
        {
            name: "InvoiceCreateWithValidInput",
            tool: "invoice_create",
            input: map[string]interface{}{
                "client_name":  "Test Client",
                "project_name": "Test Project",
                "due_date":     "2025-09-01",
            },
            expected: MCPToolResponse{
                Content: []MCPContent{
                    {
                        Type: "text",
                        Text: "Invoice created successfully",
                    },
                },
                IsError: false,
            },
            expectError: false,
        },
        {
            name: "InvoiceCreateWithMissingClient",
            tool: "invoice_create",
            input: map[string]interface{}{
                "project_name": "Test Project",
            },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            
            // Setup test dependencies with dependency injection
            executor := NewMockExecutor()
            registry := NewToolRegistry(NewValidator(), NewLogger())
            processor := NewResponseProcessor(NewLogger(), NewOutputParser())
            
            server := NewMCPServer(NewLogger(), executor, &MCPConfig{}, registry, processor)
            
            req := &MCPToolRequest{
                Tool:   tt.tool,
                Input:  tt.input,
            }
            
            resp, err := server.HandleToolCall(ctx, req)
            
            if tt.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.expected.IsError, resp.IsError)
                assert.NotEmpty(t, resp.Content)
            }
        })
    }
}

// Integration test example using testify suite
func TestFullMCPWorkflow(t *testing.T) {
    ctx := context.Background()
    
    // Setup test environment
    tempDir := t.TempDir()
    config := &MCPConfig{
        CLIPath:    "go-invoice",
        WorkingDir: tempDir,
        MaxTimeout: 30 * time.Second,
    }
    
    // Initialize MCP server with real dependencies
    executor := NewSecureExecutor(NewLogger(), NewValidator(), SandboxConfig{}, NewFileHandler())
    registry := NewToolRegistry(NewValidator(), NewLogger())
    server := NewMCPServer(NewLogger(), executor, config, registry, NewResponseProcessor(NewLogger(), NewOutputParser()))
    
    // Test 1: Create client
    clientReq := &MCPToolRequest{
        Tool: "client_create",
        Input: map[string]interface{}{
            "name":    "Integration Test Client",
            "email":   "test@example.com",
            "address": "123 Test St",
        },
    }
    
    clientResp, err := server.HandleToolCall(ctx, clientReq)
    require.NoError(t, err)
    require.False(t, clientResp.IsError)
    
    // Test 2: Create invoice
    invoiceReq := &MCPToolRequest{
        Tool: "invoice_create",
        Input: map[string]interface{}{
            "client_name": "Integration Test Client",
            "project_name": "Integration Testing",
        },
    }
    
    invoiceResp, err := server.HandleToolCall(ctx, invoiceReq)
    require.NoError(t, err)
    require.False(t, invoiceResp.IsError)
    
    // Extract invoice ID from response
    invoiceID := extractInvoiceID(invoiceResp.Content)
    require.NotEmpty(t, invoiceID)
    
    // Test 3: Import CSV data
    csvContent := "Date,Hours,Rate,Description\n2025-08-01,8,100,Development work\n2025-08-02,6,100,Testing"
    csvFile := filepath.Join(tempDir, "test-timesheet.csv")
    err = os.WriteFile(csvFile, []byte(csvContent), 0644)
    require.NoError(t, err)
    
    importReq := &MCPToolRequest{
        Tool: "import_csv",
        Input: map[string]interface{}{
            "file_path":  csvFile,
            "invoice_id": invoiceID,
        },
    }
    
    importResp, err := server.HandleToolCall(ctx, importReq)
    require.NoError(t, err)
    require.False(t, importResp.IsError)
    
    // Test 4: Generate invoice
    generateReq := &MCPToolRequest{
        Tool: "generate_invoice",
        Input: map[string]interface{}{
            "invoice_id": invoiceID,
            "template":   "default",
        },
    }
    
    generateResp, err := server.HandleToolCall(ctx, generateReq)
    require.NoError(t, err)
    require.False(t, generateResp.IsError)
    
    // Verify HTML file was generated
    htmlFiles := findFilesByExtension(generateResp.Content, ".html")
    require.NotEmpty(t, htmlFiles, "HTML invoice should be generated")
}
```

**Documentation Structure:**
```markdown
# docs/mcp-integration.md

# go-invoice MCP Integration

## Overview
This document provides comprehensive guidance for using go-invoice with Claude Desktop through the Model Context Protocol (MCP) integration.

## Installation

### Prerequisites
- go-invoice CLI installed and configured
- Claude Desktop application
- Go 1.24+ for building from source

### Setup Steps
1. Build the MCP server: `go build -o go-invoice-mcp ./cmd/go-invoice-mcp`
2. Run setup script: `./scripts/setup-claude-integration.sh`
3. Restart Claude Desktop
4. Test integration: "List my invoices" in Claude

## Available Tools

### Invoice Management
- `invoice_create`: Create new invoices
- `invoice_list`: List and filter invoices
- `invoice_show`: Display invoice details
- `invoice_update`: Modify invoice metadata
- `invoice_delete`: Remove invoices

### Data Import/Export
- `import_csv`: Import timesheet data
- `generate_invoice`: Create HTML invoices
- `export_data`: Export invoice data

### Configuration
- `config_show`: Display current configuration
- `config_validate`: Validate configuration

## Example Conversations

### Creating and Managing Invoices
```
User: "Create an invoice for Acme Corp for the website project due September 1st"
Claude: I'll create a new invoice for Acme Corp. Let me set that up for you.

[Creates invoice using invoice_create tool]

User: "Import the hours from my timesheet.csv file into that invoice"
Claude: I'll import the timesheet data into the invoice I just created.

[Uses import_csv tool to add work items]

User: "Generate the final invoice as HTML"
Claude: I'll generate the HTML invoice for you now.

[Uses generate_invoice tool to create printable invoice]
```

### Monthly Billing Workflow
```
User: "Show me all unpaid invoices from last month and generate them"
Claude: Let me find all unpaid invoices from last month and generate the HTML versions.

[Uses invoice_list with filters, then generate_invoice for each]
```

## Troubleshooting
See docs/troubleshooting-mcp.md for common issues and solutions.
```

**Verification Steps:**
```bash
# 1. Run comprehensive security and quality checks
govulncheck ./...
go mod verify
gitleaks detect --source . --log-opts="--all" --verbose

# 2. Run unit tests with testify and race detection
go test -v -race -cover ./internal/mcp/...

# 3. Run tests with coverage threshold validation
go test -cover ./internal/mcp/... | grep -E "coverage: [0-9]+" | awk '{if ($2 < 90) exit 1}'

# 4. Run MCP integration tests
go test -v -race ./test/mcp_integration_test.go

# 5. Run comprehensive linting per .github/AGENTS.md standards
golangci-lint run
go vet ./...
gofumpt -l .
goimports -l .

# 6. Test MCP server with Claude Desktop (manual)
# Start Claude Desktop and test natural language interactions

# 7. Test context cancellation across MCP operations
go test -v -run TestMCPContext ./...

# 8. Verify no global state or init functions
grep -r "var.*=" internal/mcp/ | grep -v test | grep -v const
grep -r "func init()" internal/mcp/
```

**Success Criteria:**
- ✅ Test coverage exceeds 90% using testify suite patterns for MCP components
- ✅ All critical MCP workflows have comprehensive tests with edge cases
- ✅ Documentation is clear, complete, and follows .github/AGENTS.md standards
- ✅ Examples demonstrate key MCP workflows with context handling
- ✅ Integration tests validate complete Claude Desktop workflows
- ✅ All tests use testify assertions and table-driven patterns
- ✅ Context cancellation tested across all MCP operations
- ✅ Error handling tested with proper wrapping verification
- ✅ No global state or init functions detected in MCP codebase
- ✅ Race condition testing passes for all concurrent MCP operations
- ✅ Dependency injection patterns verified in all MCP tests
- ✅ Security vulnerabilities scan clean (govulncheck)
- ✅ Secret detection passes (gitleaks)
- ✅ All linting and formatting per .github/AGENTS.md standards passes
- ✅ Final todo: Update the @plans/plan-02-status.md file with the results of the implementation

## Configuration Examples

### MCP Server Configuration
```json
{
  "server": {
    "host": "localhost",
    "port": 0,
    "timeout": "30s",
    "logLevel": "info"
  },
  "cli": {
    "path": "go-invoice",
    "workingDir": "~/.go-invoice",
    "maxTimeout": "60s"
  },
  "security": {
    "allowedCommands": [
      "go-invoice",
      "config",
      "invoice",
      "import",
      "generate"
    ],
    "sandboxEnabled": true,
    "fileAccessRestricted": true
  },
  "tools": {
    "enableAll": true,
    "categories": [
      "invoice_management",
      "data_import",
      "configuration"
    ]
  }
}
```

### Claude Desktop Integration
```json
{
  "mcpServers": {
    "go-invoice": {
      "command": "go-invoice-mcp",
      "args": ["--config", "~/.go-invoice/mcp-config.json"],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "info"
      }
    }
  }
}
```

## Implementation Timeline

- **Session 1**: MCP Foundation (Phase 0-1) - 3-4 hours
- **Session 2**: Tool Definitions (Phase 2) - 3-4 hours  
- **Session 3**: Command Execution (Phase 3) - 3-4 hours
- **Session 4**: Claude Integration (Phase 4) - 2-3 hours
- **Session 5**: Testing and Documentation (Phase 5) - 3-4 hours

Total estimated time: 14-19 hours across 5 focused sessions

## Success Metrics

### Functionality
- **Tool Coverage**: All CLI commands exposed as MCP tools
- **Response Time**: MCP tool execution within 2 seconds average
- **Data Integrity**: No data corruption across MCP operations
- **Natural Language**: Complex workflows completable through conversation

### Compatibility
- **Claude Desktop**: Full integration with current Claude Desktop versions
- **Protocol Compliance**: 100% MCP specification adherence
- **Platform Support**: Works on Linux, macOS, and Windows
- **CLI Compatibility**: Zero impact on existing CLI functionality

### Developer Experience
- **Setup Time**: From installation to first conversation in under 10 minutes
- **Documentation**: Every MCP tool documented with examples
- **Error Messages**: Clear, actionable error messages with suggested fixes
- **Extensibility**: New tools can be added in under 50 lines of code

## Conclusion

go-invoice MCP Integration represents a seamless bridge between traditional CLI workflows and modern AI-assisted interfaces. By leveraging the Model Context Protocol, users can now manage their entire invoicing workflow through natural language conversations with Claude Desktop while maintaining all the robustness, security, and features of the existing CLI application.

**Key innovations in this MCP integration:**
- **.github/AGENTS.md Compliance**: Full alignment with established engineering standards
- **Non-Disruptive Architecture**: MCP server operates independently of existing CLI
- **Context-First Design**: All MCP operations support cancellation and timeout
- **Security-First Approach**: Comprehensive command validation and sandboxing
- **Natural Language Interface**: Complex workflows become conversational
- **Comprehensive Tool Coverage**: Every CLI command exposed as MCP tool
- **Production Ready**: Full testing, documentation, and monitoring

This implementation exemplifies established patterns:
- **Protocol Compliance** following MCP specification exactly
- **Interface-Based Design** for pluggable MCP components
- **Context-Aware Operations** for proper cancellation and timeout handling
- **Structured Logging** for debugging and monitoring MCP interactions
- **Configuration as Code** with validation and defaults
- **No Global State** - dependency injection throughout
- **Error Handling Excellence** - comprehensive wrapping and context
- **Security Integration** - govulncheck, sandboxing, validation
- **Testing Standards** - testify suite with integration tests

go-invoice MCP Integration positions the tool as a leader in AI-assisted invoice management, providing users with the flexibility to choose between traditional CLI interactions and modern conversational interfaces powered by Claude.
