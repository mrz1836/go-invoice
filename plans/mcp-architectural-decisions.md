# MCP Server Architectural Decisions

## Decision Summary

Architectural decisions for the go-invoice MCP integration with rationale and alternatives considered.

## ADR-001: Standalone MCP Server Binary

**Decision**: Create separate `go-invoice-mcp` binary instead of adding MCP support to existing CLI.

**Rationale**:
- **Non-Disruptive**: Existing CLI users unaffected
- **Single Responsibility**: MCP server focused solely on protocol implementation
- **Independent Deployment**: MCP server can be updated independently
- **Process Isolation**: Security boundary between MCP and core CLI

**Alternatives Considered**:
- Embed MCP server in existing CLI (rejected: complexity, mixing concerns)
- Plugin architecture (rejected: complexity, Go plugin limitations)

**Implementation**: `cmd/go-invoice-mcp/main.go` with independent build target

---

## ADR-002: CLI Command Bridge Pattern

**Decision**: Execute existing CLI commands via process execution rather than code integration.

**Rationale**:
- **Zero Code Duplication**: Reuse all existing CLI functionality
- **Security Isolation**: Command execution sandboxed
- **Compatibility**: Works with any CLI version
- **Simplicity**: Clear separation of concerns

**Alternatives Considered**:
- Direct code integration (rejected: tight coupling, complexity)
- Shared library approach (rejected: API surface management)

**Implementation**: `internal/mcp/bridge.go` with `exec.CommandContext`

---

## ADR-003: Context-First Design Throughout

**Decision**: All MCP operations accept `context.Context` as first parameter.

**Rationale**:
- **Cancellation Support**: Proper timeout and cancellation handling
- **Request Tracing**: End-to-end request context preservation
- **Resource Management**: Automatic cleanup on cancellation
- **Go Idioms**: Follows established Go patterns

**Implementation**:
```go
func (s *MCPServer) HandleToolCall(ctx context.Context, req *MCPToolRequest) (*MCPToolResponse, error)
```

---

## ADR-004: Consumer-Driven Interface Design

**Decision**: Define interfaces in consuming packages with minimal surface area.

**Rationale**:
- **Dependency Inversion**: Consumers define what they need
- **Testing**: Easy to mock and test
- **Flexibility**: Implementation can change without affecting consumers
- **Go Best Practice**: "Accept interfaces, return concrete types"

**Implementation**:
```go
// In internal/mcp/server.go (consumer)
type CLIBridge interface {
    ExecuteCommand(ctx context.Context, cmd CommandRequest) (*CommandResponse, error)
    ValidateCommand(ctx context.Context, cmd string, args []string) error
}
```

---

## ADR-005: Dependency Injection Pattern

**Decision**: Use constructor-based dependency injection with zero global state.

**Rationale**:
- **Testability**: Easy to inject mocks for testing
- **Configurability**: Runtime configuration of dependencies
- **Thread Safety**: No shared mutable state
- **Go Patterns**: Follows existing codebase patterns

**Implementation**:
```go
func NewMCPServer(logger Logger, bridge CLIBridge, config *MCPConfig) *MCPServer
```

---

## ADR-006: Comprehensive Error Wrapping

**Decision**: Wrap all errors with operation context using `fmt.Errorf`.

**Rationale**:
- **Debugging**: Clear error chains for troubleshooting
- **Context Preservation**: Operation details preserved
- **Actionable Messages**: Users get specific guidance
- **Go 1.13+ Features**: Leverages error wrapping capabilities

**Implementation**:
```go
if err := s.bridge.ExecuteCommand(ctx, cmdReq); err != nil {
    return fmt.Errorf("MCP command execution failed for tool %s: %w", toolName, err)
}
```

---

## ADR-007: JSON Schema Validation

**Decision**: Use JSON Schema for MCP tool parameter validation.

**Rationale**:
- **Standard Format**: Industry standard for API validation
- **Self-Documenting**: Schema serves as documentation
- **Client Generation**: Schemas can generate client code
- **Comprehensive**: Supports complex validation rules

**Implementation**: Schema definitions in `internal/mcp/schemas/`

---

## ADR-008: Dual Transport Support

**Decision**: Support both HTTP (Claude Desktop) and stdio (Claude Code) transports.

**Rationale**:
- **Platform Coverage**: Both Claude platforms supported
- **Protocol Compliance**: MCP specification supports both
- **User Choice**: Users can choose preferred platform
- **Future Proofing**: Ready for new transport types

**Implementation**: Transport detection and configuration in `internal/mcp/transport.go`

---

## ADR-009: Testify Suite Testing Pattern

**Decision**: Use testify suite pattern for all MCP tests.

**Rationale**:
- **Consistency**: Matches existing codebase patterns
- **Setup/Teardown**: Clean test lifecycle management
- **Assertions**: Rich assertion library
- **Table-Driven**: Supports table-driven test patterns

**Implementation**:
```go
type MCPServerTestSuite struct {
    suite.Suite
    ctx     context.Context
    server  *MCPServer
    // test dependencies
}
```

---

## ADR-010: Security-First Approach

**Decision**: Implement comprehensive security validation at every layer.

**Rationale**:
- **Defense in Depth**: Multiple validation layers
- **Input Sanitization**: All inputs validated and sanitized
- **Command Whitelisting**: Only allowed commands executed
- **Resource Limits**: Prevent resource exhaustion

**Implementation**:
- Input validation via JSON schemas
- Command whitelisting in CLI bridge
- File system sandboxing
- Resource and timeout limits

---

## ADR-011: Configuration as Code

**Decision**: Use structured configuration files with validation.

**Rationale**:
- **Versioning**: Configuration can be version controlled
- **Validation**: Structure prevents configuration errors
- **Documentation**: Self-documenting configuration
- **Flexibility**: Easy to extend and modify

**Implementation**: JSON/YAML configuration with schema validation

---

## ADR-012: Logging with Structured Context

**Decision**: Use structured logging with operation context.

**Rationale**:
- **Debugging**: Easy to trace operations
- **Monitoring**: Metrics and alerting support
- **Performance**: Track operation timing
- **Troubleshooting**: Rich context for issues

**Implementation**:
```go
s.logger.Info("MCP tool executed", "tool", toolName, "duration", duration, "success", true)
```

---

## Cross-Cutting Concerns

### Performance
- **Startup Time**: Target <500ms server startup
- **Tool Execution**: Target <2s average execution time
- **Memory Usage**: Minimal heap allocations
- **Concurrency**: Support 5 simultaneous operations

### Security
- **Input Validation**: JSON schema validation for all inputs
- **Command Sandboxing**: Restricted command execution
- **File Access**: Limited file system access
- **Resource Limits**: CPU and memory constraints

### Reliability
- **Error Recovery**: Graceful handling of CLI failures
- **Circuit Breaker**: Protection against cascading failures
- **Health Checks**: Comprehensive system health monitoring
- **Graceful Shutdown**: Clean resource cleanup

### Maintainability
- **Code Organization**: Clear module boundaries
- **Documentation**: Comprehensive inline documentation
- **Testing**: High coverage with integration tests
- **Monitoring**: Comprehensive logging and metrics

## Implementation Guidelines

### Code Patterns
```go
// 1. Context-first functions
func Operation(ctx context.Context, params...) (result, error)

// 2. Error wrapping
return fmt.Errorf("operation failed for %s: %w", context, err)

// 3. Dependency injection
func NewComponent(deps ...Interface) *Component

// 4. Interface definition
type ComponentInterface interface {
    Method(ctx context.Context, params...) (result, error)
}
```

### Testing Patterns
```go
// 1. Testify suite
type ComponentTestSuite struct {
    suite.Suite
    // dependencies
}

// 2. Descriptive test names
func (suite *Suite) TestComponentOperationWithValidInputReturnsSuccess()

// 3. Table-driven tests
tests := []struct{
    name string
    input interface{}
    expected interface{}
    wantErr bool
}{ /* test cases */ }
```

### Security Patterns
```go
// 1. Input validation
if err := validator.Validate(input); err != nil {
    return fmt.Errorf("invalid input: %w", err)
}

// 2. Context timeout
ctx, cancel := context.WithTimeout(ctx, maxDuration)
defer cancel()

// 3. Resource cleanup
defer func() {
    if err := cleanup(); err != nil {
        logger.Error("cleanup failed", "error", err)
    }
}()
```

These architectural decisions provide a solid foundation for implementing a robust, secure, and maintainable MCP integration that follows Go best practices and integrates seamlessly with the existing go-invoice codebase.
