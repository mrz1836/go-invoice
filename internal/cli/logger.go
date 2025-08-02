// Package cli provides common command-line interface utilities for the go-invoice application.
package cli

import (
	"fmt"
	"log"
	"os"
)

// SimpleLogger provides basic logging functionality
type SimpleLogger struct {
	debug bool
}

// NewLogger creates a new simple logger
func NewLogger(debug bool) *SimpleLogger {
	return &SimpleLogger{debug: debug}
}

// Info logs an info message
func (l *SimpleLogger) Info(msg string, fields ...any) {
	log.Printf("[INFO] %s %s", msg, l.formatFields(fields...))
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, fields ...any) {
	log.Printf("[ERROR] %s %s", msg, l.formatFields(fields...))
}

// Debug logs a debug message if debug mode is enabled
func (l *SimpleLogger) Debug(msg string, fields ...any) {
	if l.debug {
		log.Printf("[DEBUG] %s %s", msg, l.formatFields(fields...))
	}
}

// Fatal logs a fatal message and exits
func (l *SimpleLogger) Fatal(msg string, fields ...any) {
	log.Printf("[FATAL] %s %s", msg, l.formatFields(fields...))
	os.Exit(1)
}

// Print prints a message to stdout without any formatting
func (l *SimpleLogger) Print(msg string) {
	fmt.Print(msg) //nolint:forbidigo // Console output for CLI
}

// Printf prints a formatted message to stdout
func (l *SimpleLogger) Printf(format string, args ...any) {
	fmt.Printf(format, args...) //nolint:forbidigo // Console output for CLI
}

// Println prints a message to stdout with a newline
func (l *SimpleLogger) Println(msg string) {
	fmt.Println(msg) //nolint:forbidigo // Console output for CLI
}

// formatFields formats key-value pairs for logging
func (l *SimpleLogger) formatFields(fields ...any) string {
	if len(fields) == 0 {
		return ""
	}

	var result string
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			result += fmt.Sprintf(" %v=%v", fields[i], fields[i+1])
		}
	}
	return result
}
