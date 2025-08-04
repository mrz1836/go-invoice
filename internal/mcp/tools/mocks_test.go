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
	// Create arguments slice starting with message and append key-value pairs
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	// Create arguments slice starting with message and append key-value pairs
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	// Create arguments slice starting with message and append key-value pairs
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	// Create arguments slice starting with message and append key-value pairs
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

// MockInputValidator provides a mock implementation of InputValidator for testing
type MockInputValidator struct {
	mock.Mock
}

func (m *MockInputValidator) ValidateAgainstSchema(ctx context.Context, input map[string]interface{}, schema map[string]interface{}) error {
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
