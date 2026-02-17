package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrz1836/go-invoice/magefiles/internal/builder/mocks"
)

func TestLDFlagsBuilder_Build_Success(t *testing.T) {
	mockShell := &mocks.MockShellRunner{OutputValue: "v1.0.0"}
	builder := NewLDFlagsBuilder(mockShell)

	result := builder.Build("main")

	assert.Contains(t, result, "-s -w")
	assert.Contains(t, result, "-X main.version=v1.0.0")
	assert.Contains(t, result, "-X main.commit=v1.0.0")
	assert.Contains(t, result, "-X main.buildDate=")

	// Verify git commands were called
	assert.Len(t, mockShell.OutputCalls, 2)
	assert.Equal(t, []string{"git", "describe", "--tags", "--always", "--dirty"}, mockShell.OutputCalls[0])
	assert.Equal(t, []string{"git", "rev-parse", "--short", "HEAD"}, mockShell.OutputCalls[1])
}

func TestLDFlagsBuilder_Build_GitVersionError(t *testing.T) {
	mockShell := &mocks.MockShellRunner{OutputError: assert.AnError}
	builder := NewLDFlagsBuilder(mockShell)

	result := builder.Build("main")

	assert.Contains(t, result, "-X main.version=dev")
	assert.Contains(t, result, "-X main.commit=unknown")
}

func TestLDFlagsBuilder_BuildWithVersion(t *testing.T) {
	mockShell := &mocks.MockShellRunner{OutputValue: "abc123"}
	builder := NewLDFlagsBuilder(mockShell)

	result := builder.BuildWithVersion("main", "custom-version")

	assert.Contains(t, result, "-X main.version=custom-version")
	assert.Contains(t, result, "-X main.commit=abc123")

	// Should only call git for commit, not version
	assert.Len(t, mockShell.OutputCalls, 1)
	assert.Equal(t, []string{"git", "rev-parse", "--short", "HEAD"}, mockShell.OutputCalls[0])
}

func TestLDFlagsBuilder_GetVersion_Success(t *testing.T) {
	mockShell := &mocks.MockShellRunner{OutputValue: "v2.1.0-beta"}
	builder := NewLDFlagsBuilder(mockShell)

	version := builder.getVersion()

	assert.Equal(t, "v2.1.0-beta", version)
}

func TestLDFlagsBuilder_GetVersion_Error(t *testing.T) {
	mockShell := &mocks.MockShellRunner{OutputError: assert.AnError}
	builder := NewLDFlagsBuilder(mockShell)

	version := builder.getVersion()

	assert.Equal(t, "dev", version)
}

func TestLDFlagsBuilder_GetCommit_Success(t *testing.T) {
	mockShell := &mocks.MockShellRunner{OutputValue: "deadbeef"}
	builder := NewLDFlagsBuilder(mockShell)

	commit := builder.getCommit()

	assert.Equal(t, "deadbeef", commit)
}

func TestLDFlagsBuilder_GetCommit_Error(t *testing.T) {
	mockShell := &mocks.MockShellRunner{OutputError: assert.AnError}
	builder := NewLDFlagsBuilder(mockShell)

	commit := builder.getCommit()

	assert.Equal(t, "unknown", commit)
}
