package mcp

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// LogLevelDebug enables debug level logging
	LogLevelDebug LogLevel = iota
	// LogLevelInfo enables info level logging
	LogLevelInfo
	// LogLevelWarn enables warning level logging
	LogLevelWarn
	// LogLevelError enables error level logging
	LogLevelError
)

// DefaultLogger implements the Logger interface with structured logging
type DefaultLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string) Logger {
	logLevel := parseLogLevel(level)
	logger := log.New(os.Stderr, "", log.LstdFlags)

	return &DefaultLogger{
		level:  logLevel,
		logger: logger,
	}
}

// Debug logs a debug message with key-value pairs
func (l *DefaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", msg, keysAndValues...)
	}
}

// Info logs an info message with key-value pairs
func (l *DefaultLogger) Info(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", msg, keysAndValues...)
	}
}

// Warn logs a warning message with key-value pairs
func (l *DefaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", msg, keysAndValues...)
	}
}

// Error logs an error message with key-value pairs
func (l *DefaultLogger) Error(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelError {
		l.log("ERROR", msg, keysAndValues...)
	}
}

// log is the internal logging method that formats structured key-value pairs
func (l *DefaultLogger) log(level, msg string, keysAndValues ...interface{}) {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")

	var pairs []string
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			value := fmt.Sprintf("%v", keysAndValues[i+1])
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
		}
	}

	var logLine string
	if len(pairs) > 0 {
		logLine = fmt.Sprintf("%s [%s] %s %s", timestamp, level, msg, strings.Join(pairs, " "))
	} else {
		logLine = fmt.Sprintf("%s [%s] %s", timestamp, level, msg)
	}

	l.logger.Println(logLine)
}

// parseLogLevel converts string log level to LogLevel enum
func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// TestLogger is a logger implementation for testing that captures log messages
type TestLogger struct {
	mu       sync.RWMutex
	messages []LogMessage
}

// LogMessage represents a captured log message
type LogMessage struct {
	Level   string
	Message string
	KVPairs map[string]interface{}
}

// NewTestLogger creates a new test logger
func NewTestLogger() *TestLogger {
	return &TestLogger{
		messages: make([]LogMessage, 0),
	}
}

// Debug captures a debug message
func (t *TestLogger) Debug(msg string, keysAndValues ...interface{}) {
	t.capture("DEBUG", msg, keysAndValues...)
}

// Info captures an info message
func (t *TestLogger) Info(msg string, keysAndValues ...interface{}) {
	t.capture("INFO", msg, keysAndValues...)
}

// Warn captures a warning message
func (t *TestLogger) Warn(msg string, keysAndValues ...interface{}) {
	t.capture("WARN", msg, keysAndValues...)
}

// Error captures an error message
func (t *TestLogger) Error(msg string, keysAndValues ...interface{}) {
	t.capture("ERROR", msg, keysAndValues...)
}

// capture stores a log message for testing
func (t *TestLogger) capture(level, msg string, keysAndValues ...interface{}) {
	kvPairs := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			kvPairs[key] = keysAndValues[i+1]
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.messages = append(t.messages, LogMessage{
		Level:   level,
		Message: msg,
		KVPairs: kvPairs,
	})
}

// GetMessages returns all captured messages
func (t *TestLogger) GetMessages() []LogMessage {
	t.mu.RLock()
	defer t.mu.RUnlock()
	messages := make([]LogMessage, len(t.messages))
	copy(messages, t.messages)
	return messages
}

// Clear clears all captured messages
func (t *TestLogger) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messages = t.messages[:0]
}

// HasMessage checks if a message with the given level and text exists
func (t *TestLogger) HasMessage(level, message string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, msg := range t.messages {
		if msg.Level == level && msg.Message == message {
			return true
		}
	}
	return false
}

// HasMessageWithKV checks if a message exists with specific key-value pairs
func (t *TestLogger) HasMessageWithKV(level, message string, key string, value interface{}) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, msg := range t.messages {
		if msg.Level == level && msg.Message == message {
			if v, exists := msg.KVPairs[key]; exists && v == value {
				return true
			}
		}
	}
	return false
}
