//go:build mage
// +build mage

// Package main provides custom build targets for go-invoice project using mage.
// This includes targets for building both the main application and MCP server.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target when no target is specified
var Default = BuildAll

const (
	mainBinary = "go-invoice"
	mcpBinary  = "go-invoice-mcp"
	binDir     = "bin"
)

// BuildMain builds the main go-invoice application.
func BuildMain() error {
	mg.Deps(ensureBinDir)
	fmt.Println("Building go-invoice...")

	ldflags := buildLDFlags("main")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(binDir, mainBinary), "./cmd/go-invoice"}

	return sh.Run("go", args...)
}

// BuildMCP builds the go-invoice-mcp server.
func BuildMCP() error {
	mg.Deps(ensureBinDir)
	fmt.Println("Building go-invoice-mcp...")

	ldflags := buildLDFlags("main")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(binDir, mcpBinary), "./cmd/go-invoice-mcp"}

	return sh.Run("go", args...)
}

// BuildAll builds both the main application and MCP server.
func BuildAll() error {
	mg.Deps(BuildMain, BuildMCP)
	fmt.Println("✅ All builds completed successfully")
	return nil
}

// InstallMain builds and installs go-invoice to $GOPATH/bin.
func InstallMain() error {
	mg.Deps(BuildMain)
	return installBinary(mainBinary)
}

// InstallMCP builds and installs go-invoice-mcp to $GOPATH/bin.
func InstallMCP() error {
	mg.Deps(BuildMCP)
	return installBinary(mcpBinary)
}

// InstallAll builds and installs both binaries to $GOPATH/bin.
func InstallAll() error {
	mg.Deps(BuildAll)
	if err := installBinary(mainBinary); err != nil {
		return err
	}
	return installBinary(mcpBinary)
}

// DevBuild builds development versions with forced "dev" version.
func DevBuild() error {
	mg.Deps(ensureBinDir)
	fmt.Println("============================================================")
	fmt.Println("=== Building Development Version ===")
	fmt.Println("============================================================")
	fmt.Println()

	ldflags := buildLDFlagsWithVersion("main", "dev")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(binDir, mainBinary), "./cmd/go-invoice"}

	if err := sh.Run("go", args...); err != nil {
		return err
	}

	if err := installBinary(mainBinary); err != nil {
		return err
	}

	fmt.Printf("✅ Installed development build of %s to %s\n", mainBinary, getGOPATHBin())
	return nil
}

// DevBuildAll builds development versions of both binaries.
func DevBuildAll() error {
	mg.Deps(ensureBinDir)
	fmt.Println("============================================================")
	fmt.Println("=== Building Development Versions (Both) ===")
	fmt.Println("============================================================")
	fmt.Println()

	// Build main with dev version
	ldflags := buildLDFlagsWithVersion("main", "dev")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(binDir, mainBinary), "./cmd/go-invoice"}
	if err := sh.Run("go", args...); err != nil {
		return err
	}

	// Build MCP with dev version
	args = []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(binDir, mcpBinary), "./cmd/go-invoice-mcp"}
	if err := sh.Run("go", args...); err != nil {
		return err
	}

	// Install both
	if err := installBinary(mainBinary); err != nil {
		return err
	}
	if err := installBinary(mcpBinary); err != nil {
		return err
	}

	fmt.Printf("✅ Installed development builds to %s\n", getGOPATHBin())
	return nil
}

// Clean removes build artifacts.
func Clean() error {
	fmt.Println("Cleaning build artifacts...")

	// Remove bin directory
	if err := sh.Rm(binDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", binDir, err)
	}

	fmt.Println("✅ Build artifacts cleaned")
	return nil
}

// CleanAll removes build artifacts and installed binaries.
func CleanAll() error {
	mg.Deps(Clean)

	gopathBin := getGOPATHBin()
	mainPath := filepath.Join(gopathBin, mainBinary)
	mcpPath := filepath.Join(gopathBin, mcpBinary)

	// Remove installed binaries
	for _, path := range []string{mainPath, mcpPath} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	fmt.Println("✅ All artifacts and installed binaries cleaned")
	return nil
}

// Helper functions

// ensureBinDir creates the bin directory if it doesn't exist.
func ensureBinDir() error {
	return os.MkdirAll(binDir, 0o755)
}

// buildLDFlags constructs ldflags for the build with version information.
func buildLDFlags(pkg string) string {
	version := getVersion()
	commit := getCommit()
	buildDate := time.Now().UTC().Format(time.RFC3339)

	return fmt.Sprintf("-s -w -X %s.version=%s -X %s.commit=%s -X %s.buildDate=%s",
		pkg, version, pkg, commit, pkg, buildDate)
}

// buildLDFlagsWithVersion constructs ldflags with a specific version.
func buildLDFlagsWithVersion(pkg, version string) string {
	commit := getCommit()
	buildDate := time.Now().UTC().Format(time.RFC3339)

	return fmt.Sprintf("-s -w -X %s.version=%s -X %s.commit=%s -X %s.buildDate=%s",
		pkg, version, pkg, commit, pkg, buildDate)
}

// getVersion returns the current version from git tags or "dev".
func getVersion() string {
	version, err := sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return "dev"
	}
	return version
}

// getCommit returns the current git commit hash.
func getCommit() string {
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "unknown"
	}
	return commit
}

// getGOPATHBin returns the GOPATH bin directory.
func getGOPATHBin() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// Default GOPATH
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		gopath = filepath.Join(home, "go")
	}
	return filepath.Join(gopath, "bin")
}

// installBinary copies a binary from bin/ to $GOPATH/bin.
func installBinary(binaryName string) error {
	src := filepath.Join(binDir, binaryName)
	dest := filepath.Join(getGOPATHBin(), binaryName)

	// Check if source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source binary %s does not exist", src)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(getGOPATHBin(), 0o755); err != nil {
		return fmt.Errorf("failed to create GOPATH bin directory: %w", err)
	}

	// Copy file
	return copyFile(src, dest)
}

// copyFile copies a file from src to dst with executable permissions.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Make executable
	var mode os.FileMode = 0o755
	if runtime.GOOS == "windows" {
		mode = 0o644
	}

	return os.Chmod(dst, mode)
}
