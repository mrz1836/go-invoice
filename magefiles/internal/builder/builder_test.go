package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrz1836/go-invoice/magefiles/internal/builder/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_EnsureBinDir_Success(t *testing.T) {
	config := &Config{BinDir: "bin"}
	mockFS := mocks.NewMockFileSystem()
	builder := NewBuilder(config, nil, mockFS, nil, nil)

	err := builder.EnsureBinDir()

	require.NoError(t, err)
	assert.Contains(t, mockFS.Files, "bin")
	assert.Equal(t, os.FileMode(0o750), mockFS.Permissions["bin"])
}

func TestBuilder_EnsureBinDir_Error(t *testing.T) {
	config := &Config{BinDir: "bin"}
	mockFS := mocks.NewMockFileSystem()
	mockFS.MkdirError = assert.AnError
	builder := NewBuilder(config, nil, mockFS, nil, nil)

	err := builder.EnsureBinDir()

	assert.Error(t, err)
}

func TestBuilder_BuildMain_Success(t *testing.T) {
	config := &Config{MainBinary: "go-invoice", BinDir: "bin"}
	mockShell := &mocks.MockShellRunner{OutputValue: "v1.0.0"}
	mockFS := mocks.NewMockFileSystem()
	mockLogger := &mocks.MockLogger{}
	builder := NewBuilder(config, mockShell, mockFS, mockLogger, nil)

	err := builder.BuildMain()

	require.NoError(t, err)
	assert.Len(t, mockShell.RunCalls, 1)

	call := mockShell.RunCalls[0]
	assert.Equal(t, "go", call[0])
	assert.Equal(t, "build", call[1])
	assert.Equal(t, "-trimpath", call[2])
	assert.Equal(t, "-ldflags", call[3])
	assert.Contains(t, call[4], "-s -w")
	assert.Equal(t, "-o", call[5])
	assert.Equal(t, filepath.Join("bin", "go-invoice"), call[6])
	assert.Equal(t, "./cmd/go-invoice", call[7])

	assert.Contains(t, mockLogger.Messages, "Building go-invoice...")
}

func TestBuilder_BuildMCP_Success(t *testing.T) {
	config := &Config{MCPBinary: "go-invoice-mcp", BinDir: "bin"}
	mockShell := &mocks.MockShellRunner{OutputValue: "v1.0.0"}
	mockFS := mocks.NewMockFileSystem()
	mockLogger := &mocks.MockLogger{}
	builder := NewBuilder(config, mockShell, mockFS, mockLogger, nil)

	err := builder.BuildMCP()

	require.NoError(t, err)
	assert.Len(t, mockShell.RunCalls, 1)

	call := mockShell.RunCalls[0]
	assert.Equal(t, "go", call[0])
	assert.Equal(t, filepath.Join("bin", "go-invoice-mcp"), call[6])
	assert.Equal(t, "./cmd/go-invoice-mcp", call[7])

	assert.Contains(t, mockLogger.Messages, "Building go-invoice-mcp...")
}

func TestBuilder_BuildAll_Success(t *testing.T) {
	config := &Config{
		MainBinary: "go-invoice",
		MCPBinary:  "go-invoice-mcp",
		BinDir:     "bin",
	}
	mockShell := &mocks.MockShellRunner{OutputValue: "v1.0.0"}
	mockFS := mocks.NewMockFileSystem()
	mockLogger := &mocks.MockLogger{}
	builder := NewBuilder(config, mockShell, mockFS, mockLogger, nil)

	err := builder.BuildAll()

	require.NoError(t, err)
	assert.Len(t, mockShell.RunCalls, 2)
	assert.Contains(t, mockLogger.Messages, "✅ All builds completed successfully")
}

func TestBuilder_InstallBinary_Success(t *testing.T) {
	config := &Config{BinDir: "bin", GOPATHBin: "/home/test/go/bin"}
	mockFS := mocks.NewMockFileSystem()
	mockFS.Files["bin/go-invoice"] = []byte("binary content")
	mockCopier := &mocks.MockFileCopier{BytesCopied: 1024}
	builder := NewBuilder(config, nil, mockFS, nil, mockCopier)

	err := builder.InstallBinary("go-invoice")

	require.NoError(t, err)
	assert.Contains(t, mockFS.Files, "/home/test/go/bin")
	assert.Equal(t, os.FileMode(0o750), mockFS.Permissions["/home/test/go/bin"])
}

func TestBuilder_InstallBinary_SourceNotExist(t *testing.T) {
	config := &Config{BinDir: "bin", GOPATHBin: "/home/test/go/bin"}
	mockFS := mocks.NewMockFileSystem()
	builder := NewBuilder(config, nil, mockFS, nil, nil)

	err := builder.InstallBinary("go-invoice")

	assert.ErrorIs(t, err, ErrSourceBinaryNotExist)
}

func TestBuilder_DevBuild_Success(t *testing.T) {
	config := &Config{
		MainBinary: "go-invoice",
		BinDir:     "bin",
		GOPATHBin:  "/home/test/go/bin",
	}
	mockShell := &mocks.MockShellRunner{OutputValue: "abc123"}
	mockFS := mocks.NewMockFileSystem()
	mockFS.Files["bin/go-invoice"] = []byte("content")
	mockLogger := &mocks.MockLogger{}
	mockCopier := &mocks.MockFileCopier{BytesCopied: 1024}
	builder := NewBuilder(config, mockShell, mockFS, mockLogger, mockCopier)

	err := builder.DevBuild()

	require.NoError(t, err)
	assert.Len(t, mockShell.RunCalls, 1)

	call := mockShell.RunCalls[0]
	assert.Contains(t, call[4], "-X main.version=dev")

	assert.Contains(t, mockLogger.Messages, "=== Building Development Version ===")
	assert.Contains(t, mockLogger.Messages, "✅ Installed development build of go-invoice to /home/test/go/bin")
}

func TestBuilder_Clean_Success(t *testing.T) {
	config := &Config{BinDir: "bin"}
	mockFS := mocks.NewMockFileSystem()
	mockFS.Files["bin"] = nil
	mockLogger := &mocks.MockLogger{}
	builder := NewBuilder(config, nil, mockFS, mockLogger, nil)

	err := builder.Clean()

	require.NoError(t, err)
	assert.NotContains(t, mockFS.Files, "bin")
	assert.Contains(t, mockLogger.Messages, "Cleaning build artifacts...")
	assert.Contains(t, mockLogger.Messages, "✅ Build artifacts cleaned")
}
