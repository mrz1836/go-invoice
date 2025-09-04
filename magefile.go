//go:build mage
// +build mage

// Package main provides custom build targets for go-invoice project using mage.
// This includes targets for building both the main application and MCP server.
package main

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target when no target is specified
var Default = BuildAll //nolint:gochecknoglobals // required by mage framework

const (
	mainBinary = "go-invoice"
	mcpBinary  = "go-invoice-mcp"
	binDir     = "bin"
)

var errSourceBinaryNotExist = errors.New("source binary does not exist")

// BuildMain builds the main go-invoice application.
func BuildMain() error {
	mg.Deps(ensureBinDir)
	log.Println("Building go-invoice...")

	ldflags := buildLDFlags("main")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(binDir, mainBinary), "./cmd/go-invoice"}

	return sh.Run("go", args...)
}

// BuildMCP builds the go-invoice-mcp server.
func BuildMCP() error {
	mg.Deps(ensureBinDir)
	log.Println("Building go-invoice-mcp...")

	ldflags := buildLDFlags("main")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(binDir, mcpBinary), "./cmd/go-invoice-mcp"}

	return sh.Run("go", args...)
}

// BuildAll builds both the main application and MCP server.
func BuildAll() error {
	mg.Deps(BuildMain, BuildMCP)
	log.Println("✅ All builds completed successfully")
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
	log.Println("============================================================")
	log.Println("=== Building Development Version ===")
	log.Println("============================================================")
	log.Println()

	ldflags := buildLDFlagsWithVersion("main", "dev")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(binDir, mainBinary), "./cmd/go-invoice"}

	if err := sh.Run("go", args...); err != nil {
		return err
	}

	if err := installBinary(mainBinary); err != nil {
		return err
	}

	log.Printf("✅ Installed development build of %s to %s\n", mainBinary, getGOPATHBin())
	return nil
}

// DevBuildAll builds development versions of both binaries.
func DevBuildAll() error {
	mg.Deps(ensureBinDir)
	log.Println("============================================================")
	log.Println("=== Building Development Versions (Both) ===")
	log.Println("============================================================")
	log.Println()

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

	log.Printf("✅ Installed development builds to %s\n", getGOPATHBin())
	return nil
}

// Clean removes build artifacts.
func Clean() error {
	log.Println("Cleaning build artifacts...")

	// Remove bin directory
	if err := sh.Rm(binDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	log.Println("✅ Build artifacts cleaned")
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
			return err
		}
	}

	log.Println("✅ All artifacts and installed binaries cleaned")
	return nil
}

// Helper functions

// ensureBinDir creates the bin directory if it doesn't exist.
func ensureBinDir() error {
	return os.MkdirAll(binDir, 0o750)
}

// buildLDFlags constructs ldflags for the build with version information.
func buildLDFlags(pkg string) string {
	version := getVersion()
	commit := getCommit()
	buildDate := time.Now().UTC().Format(time.RFC3339)

	return "-s -w -X " + pkg + ".version=" + version + " -X " + pkg + ".commit=" + commit + " -X " + pkg + ".buildDate=" + buildDate
}

// buildLDFlagsWithVersion constructs ldflags with a specific version.
func buildLDFlagsWithVersion(pkg, version string) string {
	commit := getCommit()
	buildDate := time.Now().UTC().Format(time.RFC3339)

	return "-s -w -X " + pkg + ".version=" + version + " -X " + pkg + ".commit=" + commit + " -X " + pkg + ".buildDate=" + buildDate
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
		return errSourceBinaryNotExist
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(getGOPATHBin(), 0o750); err != nil {
		return err
	}

	// Copy file
	return copyFile(src, dest)
}

// copyFile copies a file from src to dst with executable permissions.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src) //nolint:gosec // src is controlled by build system
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			log.Printf("Error closing source file: %v", closeErr)
		}
	}()

	destFile, err := os.Create(dst) //nolint:gosec // dst is controlled by build system
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			log.Printf("Error closing destination file: %v", closeErr)
		}
	}()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Make executable
	var mode os.FileMode = 0o755
	if runtime.GOOS == "windows" {
		mode = 0o644
	}

	return os.Chmod(dst, mode)
}
