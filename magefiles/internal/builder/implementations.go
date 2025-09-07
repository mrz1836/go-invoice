package builder

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/magefile/mage/sh"
)

type RealShellRunner struct{}

func (r *RealShellRunner) Run(cmd string, args ...string) error {
	return sh.Run(cmd, args...)
}

func (r *RealShellRunner) Output(cmd string, args ...string) (string, error) {
	return sh.Output(cmd, args...)
}

type RealFileSystem struct{}

func (r *RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (r *RealFileSystem) Remove(path string) error {
	return os.RemoveAll(path)
}

func (r *RealFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (r *RealFileSystem) Create(path string) (*os.File, error) {
	return os.Create(path) //nolint:gosec // path is controlled by build system
}

func (r *RealFileSystem) Open(path string) (*os.File, error) {
	return os.Open(path) //nolint:gosec // path is controlled by build system
}

func (r *RealFileSystem) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

func (r *RealFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (r *RealFileSystem) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

type RealLogger struct{}

func (r *RealLogger) Println(v ...any) {
	log.Println(v...)
}

func (r *RealLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

type RealFileCopier struct{}

func (r *RealFileCopier) Copy(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

func NewDefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	gopathBin := ""
	if err == nil {
		gopathBin = filepath.Join(homeDir, "go", "bin")
	}
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		gopathBin = filepath.Join(gopath, "bin")
	}

	return &Config{
		MainBinary: "go-invoice",
		MCPBinary:  "go-invoice-mcp",
		BinDir:     "bin",
		GOPATHBin:  gopathBin,
	}
}

func NewDefaultBuilder() *Builder {
	config := NewDefaultConfig()
	shell := &RealShellRunner{}
	fs := &RealFileSystem{}
	logger := &RealLogger{}
	copier := &RealFileCopier{}

	return NewBuilder(config, shell, fs, logger, copier)
}
