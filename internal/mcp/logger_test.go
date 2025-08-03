package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LoggerTestSuite struct {
	suite.Suite

	testLogger *TestLogger
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (s *LoggerTestSuite) SetupTest() {
	s.testLogger = NewTestLogger()
}

func (s *LoggerTestSuite) TestLoggerLevels() {
	s.testLogger.Debug("debug message", "key", "value")
	s.testLogger.Info("info message", "count", 42)
	s.testLogger.Warn("warn message", "warning", true)
	s.testLogger.Error("error message", "error", "something went wrong")

	messages := s.testLogger.GetMessages()
	s.Len(messages, 4)

	s.Equal("DEBUG", messages[0].Level)
	s.Equal("debug message", messages[0].Message)
	s.Equal("value", messages[0].KVPairs["key"])

	s.Equal("INFO", messages[1].Level)
	s.Equal("info message", messages[1].Message)
	s.Equal(42, messages[1].KVPairs["count"])

	s.Equal("WARN", messages[2].Level)
	s.Equal("warn message", messages[2].Message)
	s.Equal(true, messages[2].KVPairs["warning"])

	s.Equal("ERROR", messages[3].Level)
	s.Equal("error message", messages[3].Message)
	s.Equal("something went wrong", messages[3].KVPairs["error"])
}

func (s *LoggerTestSuite) TestHasMessage() {
	s.testLogger.Info("test message")
	s.testLogger.Error("error occurred")

	s.True(s.testLogger.HasMessage("INFO", "test message"))
	s.True(s.testLogger.HasMessage("ERROR", "error occurred"))
	s.False(s.testLogger.HasMessage("DEBUG", "test message"))
	s.False(s.testLogger.HasMessage("INFO", "nonexistent"))
}

func (s *LoggerTestSuite) TestHasMessageWithKV() {
	s.testLogger.Info("operation completed", "duration", "100ms", "success", true)
	s.testLogger.Error("operation failed", "error", "timeout", "retry", false)

	s.True(s.testLogger.HasMessageWithKV("INFO", "operation completed", "duration", "100ms"))
	s.True(s.testLogger.HasMessageWithKV("INFO", "operation completed", "success", true))
	s.True(s.testLogger.HasMessageWithKV("ERROR", "operation failed", "error", "timeout"))
	s.True(s.testLogger.HasMessageWithKV("ERROR", "operation failed", "retry", false))

	s.False(s.testLogger.HasMessageWithKV("INFO", "operation completed", "duration", "200ms"))
	s.False(s.testLogger.HasMessageWithKV("INFO", "operation completed", "nonexistent", "value"))
	s.False(s.testLogger.HasMessageWithKV("DEBUG", "operation completed", "duration", "100ms"))
}

func (s *LoggerTestSuite) TestClearMessages() {
	s.testLogger.Info("message 1")
	s.testLogger.Error("message 2")

	s.Len(s.testLogger.GetMessages(), 2)

	s.testLogger.Clear()
	s.Empty(s.testLogger.GetMessages())
}

func (s *LoggerTestSuite) TestLoggerWithKeyValuePairs() {
	s.testLogger.Info("processing request",
		"method", "POST",
		"path", "/api/test",
		"duration", "150ms",
		"status", 200)

	messages := s.testLogger.GetMessages()
	s.Require().Len(messages, 1)

	msg := messages[0]
	s.Equal("INFO", msg.Level)
	s.Equal("processing request", msg.Message)
	s.Equal("POST", msg.KVPairs["method"])
	s.Equal("/api/test", msg.KVPairs["path"])
	s.Equal("150ms", msg.KVPairs["duration"])
	s.Equal(200, msg.KVPairs["status"])
}

func (s *LoggerTestSuite) TestLoggerWithOddNumberOfKeyValues() {
	// Test odd number of arguments (missing value for last key)
	s.testLogger.Info("test message", "key1", "value1", "key2")

	messages := s.testLogger.GetMessages()
	s.Require().Len(messages, 1)

	msg := messages[0]
	s.Equal("value1", msg.KVPairs["key1"])
	s.NotContains(msg.KVPairs, "key2") // Should not include key without value
}

func TestDefaultLoggerCreation(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected LogLevel
	}{
		{
			name:     "DebugLevel",
			level:    "debug",
			expected: LogLevelDebug,
		},
		{
			name:     "InfoLevel",
			level:    "info",
			expected: LogLevelInfo,
		},
		{
			name:     "WarnLevel",
			level:    "warn",
			expected: LogLevelWarn,
		},
		{
			name:     "ErrorLevel",
			level:    "error",
			expected: LogLevelError,
		},
		{
			name:     "InvalidLevel",
			level:    "invalid",
			expected: LogLevelInfo, // Should default to info
		},
		{
			name:     "EmptyLevel",
			level:    "",
			expected: LogLevelInfo, // Should default to info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			assert.NotNil(t, logger)

			// Verify logger is of correct type
			defaultLogger, ok := logger.(*DefaultLogger)
			assert.True(t, ok)
			assert.Equal(t, tt.expected, defaultLogger.level)
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"DEBUG", LogLevelDebug},
		{"info", LogLevelInfo},
		{"INFO", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"WARN", LogLevelWarn},
		{"warning", LogLevelWarn},
		{"WARNING", LogLevelWarn},
		{"error", LogLevelError},
		{"ERROR", LogLevelError},
		{"invalid", LogLevelInfo},
		{"", LogLevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkTestLoggerCapture(b *testing.B) {
	logger := NewTestLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "timestamp", "2024-01-01T00:00:00Z")
	}
}

func BenchmarkDefaultLoggerInfo(b *testing.B) {
	logger := NewLogger("info")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "timestamp", "2024-01-01T00:00:00Z")
	}
}

func BenchmarkTestLoggerHasMessage(b *testing.B) {
	logger := NewTestLogger()

	// Populate with some messages
	for i := 0; i < 100; i++ {
		logger.Info("test message", "id", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.HasMessage("INFO", "test message")
	}
}
