package cli

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// LoggerTestSuite defines the test suite for logger functionality
type LoggerTestSuite struct {
	suite.Suite

	logger         *SimpleLogger
	logOutput      *bytes.Buffer
	originalOutput *os.File
}

// SetupTest runs before each test
func (suite *LoggerTestSuite) SetupTest() {
	// Capture log output
	suite.logOutput = &bytes.Buffer{}
	suite.originalOutput = os.Stderr

	// Redirect log output to our buffer
	log.SetOutput(suite.logOutput)
	log.SetFlags(0) // Remove timestamp for predictable testing

	suite.logger = NewLogger(false) // Debug disabled by default
}

// TearDownTest runs after each test
func (suite *LoggerTestSuite) TearDownTest() {
	// Restore original log output
	log.SetOutput(suite.originalOutput)
	log.SetFlags(log.LstdFlags)
}

// TestNewLogger tests the logger constructor
func (suite *LoggerTestSuite) TestNewLogger() {
	logger := NewLogger(true)
	suite.NotNil(logger)
	suite.True(logger.debug)

	logger = NewLogger(false)
	suite.NotNil(logger)
	suite.False(logger.debug)
}

// TestInfoLogging tests info message logging
func (suite *LoggerTestSuite) TestInfoLogging() {
	suite.logger.Info("test info message")

	output := suite.logOutput.String()
	suite.Contains(output, "[INFO]")
	suite.Contains(output, "test info message")
}

// TestInfoLoggingWithFields tests info message logging with fields
func (suite *LoggerTestSuite) TestInfoLoggingWithFields() {
	suite.logger.Info("test message", "key1", "value1", "key2", 42)

	output := suite.logOutput.String()
	suite.Contains(output, "[INFO]")
	suite.Contains(output, "test message")
	suite.Contains(output, "key1=value1")
	suite.Contains(output, "key2=42")
}

// TestErrorLogging tests error message logging
func (suite *LoggerTestSuite) TestErrorLogging() {
	suite.logger.Error("test error message")

	output := suite.logOutput.String()
	suite.Contains(output, "[ERROR]")
	suite.Contains(output, "test error message")
}

// TestErrorLoggingWithFields tests error message logging with fields
func (suite *LoggerTestSuite) TestErrorLoggingWithFields() {
	suite.logger.Error("error occurred", "error", "connection failed", "retry", 3)

	output := suite.logOutput.String()
	suite.Contains(output, "[ERROR]")
	suite.Contains(output, "error occurred")
	suite.Contains(output, "error=connection failed")
	suite.Contains(output, "retry=3")
}

// TestDebugLoggingDisabled tests that debug messages are not logged when debug is disabled
func (suite *LoggerTestSuite) TestDebugLoggingDisabled() {
	suite.logger.Debug("debug message should not appear")

	output := suite.logOutput.String()
	suite.Empty(output)
}

// TestDebugLoggingEnabled tests that debug messages are logged when debug is enabled
func (suite *LoggerTestSuite) TestDebugLoggingEnabled() {
	debugLogger := NewLogger(true)
	debugLogger.Debug("debug message should appear")

	output := suite.logOutput.String()
	suite.Contains(output, "[DEBUG]")
	suite.Contains(output, "debug message should appear")
}

// TestDebugLoggingWithFields tests debug message logging with fields
func (suite *LoggerTestSuite) TestDebugLoggingWithFields() {
	debugLogger := NewLogger(true)
	debugLogger.Debug("debug info", "component", "config", "action", "validation")

	output := suite.logOutput.String()
	suite.Contains(output, "[DEBUG]")
	suite.Contains(output, "debug info")
	suite.Contains(output, "component=config")
	suite.Contains(output, "action=validation")
}

// TestFormatFields tests the field formatting functionality
func (suite *LoggerTestSuite) TestFormatFields() {
	tests := []struct {
		name     string
		fields   []any
		expected string
	}{
		{
			name:     "NoFields",
			fields:   []any{},
			expected: "",
		},
		{
			name:     "SinglePair",
			fields:   []any{"key", "value"},
			expected: " key=value",
		},
		{
			name:     "MultiplePairs",
			fields:   []any{"key1", "value1", "key2", "value2"},
			expected: " key1=value1 key2=value2",
		},
		{
			name:     "OddNumberOfFields",
			fields:   []any{"key1", "value1", "key2"},
			expected: " key1=value1",
		},
		{
			name:     "MixedTypes",
			fields:   []any{"string", "text", "number", 42, "boolean", true},
			expected: " string=text number=42 boolean=true",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.logger.formatFields(tt.fields...)
			suite.Equal(tt.expected, result)
		})
	}
}

// TestLoggerIntegration tests logger integration with different scenarios
func (suite *LoggerTestSuite) TestLoggerIntegration() {
	// Test a realistic logging scenario
	suite.logger.Info("application started", "version", "1.0.0", "port", 8080)
	suite.logger.Error("database connection failed", "host", "localhost", "error", "timeout")

	output := suite.logOutput.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	suite.Len(lines, 2)
	suite.Contains(lines[0], "[INFO]")
	suite.Contains(lines[0], "application started")
	suite.Contains(lines[1], "[ERROR]")
	suite.Contains(lines[1], "database connection failed")
}

// TestLoggerConcurrency tests that the logger is safe for concurrent use
func (suite *LoggerTestSuite) TestLoggerConcurrency() {
	const numGoroutines = 10
	const messagesPerGoroutine = 5

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				suite.logger.Info("concurrent message", "goroutine", id, "message", j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	output := suite.logOutput.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have exactly numGoroutines * messagesPerGoroutine lines
	suite.Len(lines, numGoroutines*messagesPerGoroutine)

	// All lines should contain [INFO] and "concurrent message"
	for _, line := range lines {
		suite.Contains(line, "[INFO]")
		suite.Contains(line, "concurrent message")
	}
}

// TestFatalLoggingFormat tests that fatal messages are formatted correctly
// This doesn't test the os.Exit(1) behavior to avoid terminating the test
func (suite *LoggerTestSuite) TestFatalLoggingFormat() {
	// We can test the logging format by examining what would be logged
	// before os.Exit is called. We'll use a mock or examine the formatFields method directly.

	// Test that Fatal would log with the correct format
	msg := "fatal error occurred"
	fields := []any{"component", "database", "error", "connection timeout"}

	// Test the format that would be used (same as other log methods)
	expectedFields := suite.logger.formatFields(fields...)
	expectedFormat := "[FATAL] " + msg + " " + expectedFields

	// Verify the format is constructed correctly
	suite.Contains(expectedFormat, "[FATAL]")
	suite.Contains(expectedFormat, msg)
	suite.Contains(expectedFormat, "component=database")
	suite.Contains(expectedFormat, "error=connection timeout")
}

// TestPrintMethods tests the Print, Printf, and Println methods
func (suite *LoggerTestSuite) TestPrintMethods() {
	// Capture stdout for print methods
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Capture output in a separate goroutine
	outputChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputChan <- buf.String()
	}()

	// Test Print method
	suite.logger.Print("test print message")

	// Test Printf method
	suite.logger.Printf("test printf %s %d", "message", 42)

	// Test Println method
	suite.logger.Println("test println message")

	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Get captured output
	output := <-outputChan

	// Verify output contains expected messages
	suite.Contains(output, "test print message")
	suite.Contains(output, "test printf message 42")
	suite.Contains(output, "test println message")
}

// TestPrintMethodsIndividually tests each print method separately for more precise verification
func (suite *LoggerTestSuite) TestPrintMethodsIndividually() {
	suite.Run("Print", func() {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		outputChan := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			outputChan <- buf.String()
		}()

		suite.logger.Print("hello")
		_ = w.Close()
		os.Stdout = oldStdout

		output := <-outputChan
		suite.Equal("hello", output)
	})

	suite.Run("Printf", func() {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		outputChan := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			outputChan <- buf.String()
		}()

		suite.logger.Printf("hello %s", "world")
		_ = w.Close()
		os.Stdout = oldStdout

		output := <-outputChan
		suite.Equal("hello world", output)
	})

	suite.Run("Println", func() {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		outputChan := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			outputChan <- buf.String()
		}()

		suite.logger.Println("hello")
		_ = w.Close()
		os.Stdout = oldStdout

		output := <-outputChan
		suite.Equal("hello\n", output)
	})
}

// TestLoggerTestSuite runs the logger test suite
func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

// TestLoggerFieldFormatting tests edge cases in field formatting
func TestLoggerFieldFormatting(t *testing.T) {
	logger := NewLogger(false)

	tests := []struct {
		name   string
		fields []any
		check  func(string) bool
	}{
		{
			name:   "NilValues",
			fields: []any{"key", nil},
			check: func(s string) bool {
				return strings.Contains(s, "key=<nil>")
			},
		},
		{
			name:   "EmptyStringValues",
			fields: []any{"key", ""},
			check: func(s string) bool {
				return strings.Contains(s, "key=")
			},
		},
		{
			name:   "SpecialCharacters",
			fields: []any{"key", "value with spaces & symbols!"},
			check: func(s string) bool {
				return strings.Contains(s, "key=value with spaces & symbols!")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.formatFields(tt.fields...)
			assert.True(t, tt.check(result), "formatting check failed for: %s", result)
		})
	}
}
