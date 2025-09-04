package tools

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockLogger provides a mock implementation of the Logger interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	// No-op for testing - just record that it was called
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	// No-op for testing - just record that it was called
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	// No-op for testing - just record that it was called
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	// No-op for testing - just record that it was called
}

// MockInputValidator provides a mock implementation of InputValidator for testing
type MockInputValidator struct {
	mock.Mock
}

func (m *MockInputValidator) ValidateAgainstSchema(ctx context.Context, input, schema map[string]interface{}) error {
	args := m.Called(ctx, input, schema)
	return args.Error(0)
}

func (m *MockInputValidator) ValidateRequired(ctx context.Context, input map[string]interface{}, requiredFields []string) error {
	args := m.Called(ctx, input, requiredFields)
	return args.Error(0)
}

func (m *MockInputValidator) ValidateFormat(ctx context.Context, fieldName string, value interface{}, format string) error {
	args := m.Called(ctx, fieldName, value, format)
	return args.Error(0)
}

func (m *MockInputValidator) BuildValidationError(ctx context.Context, fieldPath, message string, suggestions []string) error {
	args := m.Called(ctx, fieldPath, message, suggestions)
	return args.Error(0)
}

// MockToolRegistry provides a mock implementation of ToolRegistry for testing
type MockToolRegistry struct {
	mock.Mock

	// Mock storage for testing
	tools map[CategoryType][]*MCPTool
}

func NewMockToolRegistry() *MockToolRegistry {
	return &MockToolRegistry{
		tools: make(map[CategoryType][]*MCPTool),
	}
}

func (m *MockToolRegistry) RegisterTool(ctx context.Context, tool *MCPTool) error {
	args := m.Called(ctx, tool)
	if args.Error(0) == nil {
		m.tools[tool.Category] = append(m.tools[tool.Category], tool)
	}
	return args.Error(0)
}

func (m *MockToolRegistry) GetTool(ctx context.Context, name string) (*MCPTool, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*MCPTool), args.Error(1)
}

func (m *MockToolRegistry) ListTools(ctx context.Context, category CategoryType) ([]*MCPTool, error) {
	// Just use direct storage lookup without mocking framework complexity
	// If specific category requested, return tools for that category
	if category != "" {
		if tools, exists := m.tools[category]; exists {
			return tools, nil
		}
		return []*MCPTool{}, nil
	}

	// If empty category, return all tools
	var allTools []*MCPTool
	for _, categoryTools := range m.tools {
		allTools = append(allTools, categoryTools...)
	}
	return allTools, nil
}

func (m *MockToolRegistry) ValidateToolInput(ctx context.Context, toolName string, input map[string]interface{}) error {
	args := m.Called(ctx, toolName, input)
	return args.Error(0)
}

func (m *MockToolRegistry) GetCategories(ctx context.Context) ([]CategoryType, error) {
	args := m.Called(ctx)
	return args.Get(0).([]CategoryType), args.Error(1)
}

// SetListToolsResponse sets up mock response for ListTools calls
func (m *MockToolRegistry) SetListToolsResponse(category CategoryType, tools []*MCPTool) {
	m.tools[category] = tools
}

// NewMockLogger creates a new mock logger that doesn't require expectations
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}
