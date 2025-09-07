package builder

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
)

type Builder struct {
	config  *Config
	shell   ShellRunner
	fs      FileSystem
	logger  Logger
	copier  FileCopier
	ldflags *LDFlagsBuilder
}

func NewBuilder(config *Config, shell ShellRunner, fs FileSystem, logger Logger, copier FileCopier) *Builder {
	return &Builder{
		config:  config,
		shell:   shell,
		fs:      fs,
		logger:  logger,
		copier:  copier,
		ldflags: NewLDFlagsBuilder(shell),
	}
}

func (b *Builder) EnsureBinDir() error {
	return b.fs.MkdirAll(b.config.BinDir, 0o750)
}

func (b *Builder) BuildMain() error {
	if err := b.EnsureBinDir(); err != nil {
		return err
	}

	b.logger.Println("Building go-invoice...")

	ldflags := b.ldflags.Build("main")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(b.config.BinDir, b.config.MainBinary), "./cmd/go-invoice"}

	return b.shell.Run("go", args...)
}

func (b *Builder) BuildMCP() error {
	if err := b.EnsureBinDir(); err != nil {
		return err
	}

	b.logger.Println("Building go-invoice-mcp...")

	ldflags := b.ldflags.Build("main")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(b.config.BinDir, b.config.MCPBinary), "./cmd/go-invoice-mcp"}

	return b.shell.Run("go", args...)
}

func (b *Builder) BuildAll() error {
	if err := b.BuildMain(); err != nil {
		return err
	}
	if err := b.BuildMCP(); err != nil {
		return err
	}

	b.logger.Println("✅ All builds completed successfully")
	return nil
}

func (b *Builder) InstallBinary(binaryName string) error {
	src := filepath.Join(b.config.BinDir, binaryName)
	dest := filepath.Join(b.config.GOPATHBin, binaryName)

	if _, err := b.fs.Stat(src); b.fs.IsNotExist(err) {
		return ErrSourceBinaryNotExist
	}

	if err := b.fs.MkdirAll(b.config.GOPATHBin, 0o750); err != nil {
		return err
	}

	return b.copyFile(src, dest)
}

func (b *Builder) InstallMain() error {
	if err := b.BuildMain(); err != nil {
		return err
	}
	return b.InstallBinary(b.config.MainBinary)
}

func (b *Builder) InstallMCP() error {
	if err := b.BuildMCP(); err != nil {
		return err
	}
	return b.InstallBinary(b.config.MCPBinary)
}

func (b *Builder) InstallAll() error {
	if err := b.BuildAll(); err != nil {
		return err
	}
	if err := b.InstallBinary(b.config.MainBinary); err != nil {
		return err
	}
	return b.InstallBinary(b.config.MCPBinary)
}

func (b *Builder) DevBuild() error {
	if err := b.EnsureBinDir(); err != nil {
		return err
	}

	b.logger.Println("============================================================")
	b.logger.Println("=== Building Development Version ===")
	b.logger.Println("============================================================")
	b.logger.Println()

	ldflags := b.ldflags.BuildWithVersion("main", "dev")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(b.config.BinDir, b.config.MainBinary), "./cmd/go-invoice"}

	if err := b.shell.Run("go", args...); err != nil {
		return err
	}

	if err := b.InstallBinary(b.config.MainBinary); err != nil {
		return err
	}

	b.logger.Printf("✅ Installed development build of %s to %s\n", b.config.MainBinary, b.config.GOPATHBin)
	return nil
}

func (b *Builder) DevBuildAll() error {
	if err := b.EnsureBinDir(); err != nil {
		return err
	}

	b.logger.Println("============================================================")
	b.logger.Println("=== Building Development Versions (Both) ===")
	b.logger.Println("============================================================")
	b.logger.Println()

	ldflags := b.ldflags.BuildWithVersion("main", "dev")

	// Build main with dev version
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(b.config.BinDir, b.config.MainBinary), "./cmd/go-invoice"}
	if err := b.shell.Run("go", args...); err != nil {
		return err
	}

	// Build MCP with dev version
	args = []string{"build", "-trimpath", "-ldflags", ldflags, "-o", filepath.Join(b.config.BinDir, b.config.MCPBinary), "./cmd/go-invoice-mcp"}
	if err := b.shell.Run("go", args...); err != nil {
		return err
	}

	// Install both
	if err := b.InstallBinary(b.config.MainBinary); err != nil {
		return err
	}
	if err := b.InstallBinary(b.config.MCPBinary); err != nil {
		return err
	}

	b.logger.Printf("✅ Installed development builds to %s\n", b.config.GOPATHBin)
	return nil
}

func (b *Builder) Clean() error {
	b.logger.Println("Cleaning build artifacts...")

	if err := b.fs.Remove(b.config.BinDir); err != nil && !b.fs.IsNotExist(err) {
		return err
	}

	b.logger.Println("✅ Build artifacts cleaned")
	return nil
}

func (b *Builder) CleanAll() error {
	if err := b.Clean(); err != nil {
		return err
	}

	mainPath := filepath.Join(b.config.GOPATHBin, b.config.MainBinary)
	mcpPath := filepath.Join(b.config.GOPATHBin, b.config.MCPBinary)

	for _, path := range []string{mainPath, mcpPath} {
		if err := b.fs.Remove(path); err != nil && !b.fs.IsNotExist(err) {
			return err
		}
	}

	b.logger.Println("✅ All artifacts and installed binaries cleaned")
	return nil
}

func (b *Builder) copyFile(src, dst string) error {
	sourceFile, err := b.fs.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil && b.logger != nil {
			b.logger.Printf("Error closing source file: %v", closeErr)
		}
	}()

	destFile, err := b.fs.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil && b.logger != nil {
			b.logger.Printf("Error closing destination file: %v", closeErr)
		}
	}()

	if b.copier != nil {
		if _, err = b.copier.Copy(destFile, sourceFile); err != nil {
			return err
		}
	} else {
		if _, err = io.Copy(destFile, sourceFile); err != nil {
			return err
		}
	}

	var mode os.FileMode = 0o755
	if runtime.GOOS == "windows" {
		mode = 0o644
	}

	return b.fs.Chmod(dst, mode)
}
