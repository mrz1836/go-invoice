//go:build mage

// Package main provides custom build targets for go-invoice project using mage.
// This includes targets for building both the main application and MCP server.
package magefiles

import (
	"github.com/mrz/go-invoice/magefiles/internal/builder"
)

// Default target when no target is specified
var Default = BuildAll //nolint:gochecknoglobals // required by mage framework

var build = builder.NewDefaultBuilder() //nolint:gochecknoglobals // shared builder instance

// BuildMain builds the main go-invoice application.
func BuildMain() error {
	return build.BuildMain()
}

// BuildMCP builds the go-invoice-mcp server.
func BuildMCP() error {
	return build.BuildMCP()
}

// BuildAll builds both the main application and MCP server.
func BuildAll() error {
	return build.BuildAll()
}

// InstallMain builds and installs go-invoice to $GOPATH/bin.
func InstallMain() error {
	return build.InstallMain()
}

// InstallMCP builds and installs go-invoice-mcp to $GOPATH/bin.
func InstallMCP() error {
	return build.InstallMCP()
}

// InstallAll builds and installs both binaries to $GOPATH/bin.
func InstallAll() error {
	return build.InstallAll()
}

// DevBuild builds development versions with forced "dev" version.
func DevBuild() error {
	return build.DevBuild()
}

// DevBuildAll builds development versions of both binaries.
func DevBuildAll() error {
	return build.DevBuildAll()
}

// Clean removes build artifacts.
func Clean() error {
	return build.Clean()
}

// CleanAll removes build artifacts and installed binaries.
func CleanAll() error {
	return build.CleanAll()
}
