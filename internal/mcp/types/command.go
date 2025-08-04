// Package types provides specialized data structures for MCP command execution.
// The package name "types" is appropriate here as it contains type definitions
// for command request/response structures used throughout the MCP protocol.
package types

import (
	"time"
)

// CommandRequest represents a CLI command execution request.
type CommandRequest struct {
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	WorkingDir string            `json:"workingDir,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	Timeout    time.Duration     `json:"timeout,omitempty"`
	ExpectJSON bool              `json:"expectJSON,omitempty"`
	InputFiles []string          `json:"inputFiles,omitempty"`
}

// CommandResponse represents the result of CLI command execution.
type CommandResponse struct {
	ExitCode int           `json:"exitCode"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
	Files    []string      `json:"files,omitempty"`
}
