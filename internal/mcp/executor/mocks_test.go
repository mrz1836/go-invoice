package executor

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of the Logger interface for testing.
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, 1+len(keysAndValues))
	args = append(args, msg)
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, 1+len(keysAndValues))
	args = append(args, msg)
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, 1+len(keysAndValues))
	args = append(args, msg)
	args = append(args, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, 1+len(keysAndValues))
	args = append(args, msg)
	args = append(args, keysAndValues...)
	m.Called(args...)
}

// MockCommandValidator is a mock implementation of the CommandValidator interface for testing.
type MockCommandValidator struct {
	mock.Mock
}

func (m *MockCommandValidator) ValidateCommand(ctx context.Context, command string, args []string) error {
	arguments := m.Called(ctx, command, args)
	return arguments.Error(0)
}

func (m *MockCommandValidator) ValidateEnvironment(ctx context.Context, env map[string]string) error {
	arguments := m.Called(ctx, env)
	return arguments.Error(0)
}

func (m *MockCommandValidator) ValidatePath(ctx context.Context, path string) error {
	arguments := m.Called(ctx, path)
	return arguments.Error(0)
}

// MockFileHandler is a mock implementation of the FileHandler interface for testing.
type MockFileHandler struct {
	mock.Mock
}

func (m *MockFileHandler) PrepareWorkspace(ctx context.Context, files []FileReference) (workDir string, cleanup func(), err error) {
	arguments := m.Called(ctx, files)
	return arguments.String(0), arguments.Get(1).(func()), arguments.Error(2)
}

func (m *MockFileHandler) CollectOutputFiles(ctx context.Context, workDir string, patterns []string) ([]FileReference, error) {
	arguments := m.Called(ctx, workDir, patterns)
	return arguments.Get(0).([]FileReference), arguments.Error(1)
}

func (m *MockFileHandler) ValidateFile(ctx context.Context, path string) error {
	arguments := m.Called(ctx, path)
	return arguments.Error(0)
}

func (m *MockFileHandler) CreateTempFile(ctx context.Context, pattern string, content []byte) (string, error) {
	arguments := m.Called(ctx, pattern, content)
	return arguments.String(0), arguments.Error(1)
}

// MockOutputParser is a mock implementation of the OutputParser interface for testing.
type MockOutputParser struct {
	mock.Mock
}

func (m *MockOutputParser) ParseJSON(ctx context.Context, output string) (map[string]interface{}, error) {
	arguments := m.Called(ctx, output)
	return arguments.Get(0).(map[string]interface{}), arguments.Error(1)
}

func (m *MockOutputParser) ParseTable(ctx context.Context, output string) ([]map[string]string, error) {
	arguments := m.Called(ctx, output)
	return arguments.Get(0).([]map[string]string), arguments.Error(1)
}

func (m *MockOutputParser) ParseKeyValue(ctx context.Context, output string) (map[string]string, error) {
	arguments := m.Called(ctx, output)
	return arguments.Get(0).(map[string]string), arguments.Error(1)
}

func (m *MockOutputParser) ExtractError(ctx context.Context, stdout, stderr string, exitCode int) error {
	arguments := m.Called(ctx, stdout, stderr, exitCode)
	return arguments.Error(0)
}

// MockProgressTracker is a mock implementation of the ProgressTracker interface for testing.
type MockProgressTracker struct {
	mock.Mock
}

func (m *MockProgressTracker) StartOperation(ctx context.Context, operationID, description string, totalSteps int) (*Operation, error) {
	arguments := m.Called(ctx, operationID, description, totalSteps)
	return arguments.Get(0).(*Operation), arguments.Error(1)
}

func (m *MockProgressTracker) GetOperation(ctx context.Context, operationID string) (*Operation, error) {
	arguments := m.Called(ctx, operationID)
	return arguments.Get(0).(*Operation), arguments.Error(1)
}

func (m *MockProgressTracker) ListOperations(ctx context.Context) ([]*Operation, error) {
	arguments := m.Called(ctx)
	return arguments.Get(0).([]*Operation), arguments.Error(1)
}

func (m *MockProgressTracker) Subscribe(ctx context.Context, operationID string, callback ProgressFunc) error {
	arguments := m.Called(ctx, operationID, callback)
	return arguments.Error(0)
}

func (m *MockProgressTracker) Unsubscribe(ctx context.Context, operationID string, callback ProgressFunc) error {
	arguments := m.Called(ctx, operationID, callback)
	return arguments.Error(0)
}

// MockCommandExecutor is a mock implementation of the CommandExecutor interface for testing.
type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	arguments := m.Called(ctx, req)
	return arguments.Get(0).(*ExecutionResponse), arguments.Error(1)
}

func (m *MockCommandExecutor) ValidateCommand(ctx context.Context, command string, args []string) error {
	arguments := m.Called(ctx, command, args)
	return arguments.Error(0)
}

func (m *MockCommandExecutor) GetAllowedCommands(ctx context.Context) ([]string, error) {
	arguments := m.Called(ctx)
	return arguments.Get(0).([]string), arguments.Error(1)
}
