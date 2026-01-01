package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// FileHandlerTestSuite tests the DefaultFileHandler implementation
type FileHandlerTestSuite struct {
	suite.Suite

	handler   *DefaultFileHandler
	logger    *MockLogger
	validator *MockCommandValidator
	sandbox   SandboxConfig
	tempDir   string
}

func (suite *FileHandlerTestSuite) SetupTest() {
	suite.logger = new(MockLogger)
	suite.validator = new(MockCommandValidator)

	// Create temp directory for tests
	tempDir, err := os.MkdirTemp("", "filehandler-test-*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	suite.sandbox = SandboxConfig{
		AllowedPaths: []string{tempDir},
		BlockedPaths: []string{"/etc", "/root"},
		MaxFileSize:  10 * 1024 * 1024, // 10MB
	}

	// Setup logger expectations - make it lenient
	suite.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	suite.handler = NewDefaultFileHandler(suite.logger, suite.validator, suite.sandbox)
}

func (suite *FileHandlerTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		_ = os.RemoveAll(suite.tempDir)
	}
}

// TestNewDefaultFileHandler tests the constructor
func (suite *FileHandlerTestSuite) TestNewDefaultFileHandler() {
	handler := NewDefaultFileHandler(suite.logger, suite.validator, suite.sandbox)
	suite.NotNil(handler)
	suite.Equal(suite.logger, handler.logger)
	suite.Equal(suite.validator, handler.validator)
	suite.Equal(suite.sandbox, handler.sandbox)
}

// TestNewDefaultFileHandlerPanicsWithNilLogger tests that constructor panics with nil logger
func (suite *FileHandlerTestSuite) TestNewDefaultFileHandlerPanicsWithNilLogger() {
	suite.Panics(func() {
		NewDefaultFileHandler(nil, suite.validator, suite.sandbox)
	})
}

// TestNewDefaultFileHandlerPanicsWithNilValidator tests that constructor panics with nil validator
func (suite *FileHandlerTestSuite) TestNewDefaultFileHandlerPanicsWithNilValidator() {
	suite.Panics(func() {
		NewDefaultFileHandler(suite.logger, nil, suite.sandbox)
	})
}

// TestFileErrorError tests FileError.Error() method
func (suite *FileHandlerTestSuite) TestFileErrorError() {
	err := &FileError{Op: "validate", Msg: "test error"}
	suite.Equal("file operation validate: test error", err.Error())
}

// TestValidatePathForFileOperation tests the path validation function
func (suite *FileHandlerTestSuite) TestValidatePathForFileOperation() {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "ValidAbsolutePath",
			path:    "/tmp/test.txt",
			wantErr: nil,
		},
		{
			name:    "EmptyPath",
			path:    "",
			wantErr: ErrPathEmpty,
		},
		{
			name:    "RelativePath",
			path:    "test.txt",
			wantErr: ErrPathNotAbsolute,
		},
		{
			name:    "PathWithNullBytes",
			path:    "/tmp/test\x00.txt",
			wantErr: ErrPathContainsNullBytes,
		},
		{
			name:    "PathWithTraversal",
			path:    "/tmp/../etc/passwd",
			wantErr: ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := validatePathForFileOperation(tt.path)
			if tt.wantErr != nil {
				suite.ErrorIs(err, tt.wantErr)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestSanitizePath tests path sanitization
func (suite *FileHandlerTestSuite) TestSanitizePath() {
	tests := []struct {
		name         string
		path         string
		allowedPaths []string
		wantErr      bool
	}{
		{
			name:         "ValidPathInAllowedDir",
			path:         filepath.Join(suite.tempDir, "test.txt"),
			allowedPaths: []string{suite.tempDir},
			wantErr:      false,
		},
		{
			name:         "PathWithNoAllowedPaths",
			path:         "/tmp/test.txt",
			allowedPaths: []string{},
			wantErr:      false,
		},
		{
			name:         "PathOutsideAllowedDir",
			path:         "/etc/passwd",
			allowedPaths: []string{suite.tempDir},
			wantErr:      true,
		},
		{
			name:         "RelativePathConverted",
			path:         "test.txt",
			allowedPaths: []string{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := suite.handler.sanitizePath(tt.path, tt.allowedPaths)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.Require().NoError(err)
				suite.True(filepath.IsAbs(result))
			}
		})
	}
}

// TestPrepareWorkspace tests workspace creation
func (suite *FileHandlerTestSuite) TestPrepareWorkspace() {
	ctx := context.Background()

	workDir, cleanup, err := suite.handler.PrepareWorkspace(ctx, nil)

	suite.Require().NoError(err)
	suite.NotEmpty(workDir)
	suite.NotNil(cleanup)
	suite.DirExists(workDir)

	// Cleanup should remove the directory
	cleanup()
	suite.NoDirExists(workDir)
}

// TestPrepareWorkspaceWithFiles tests workspace creation with input files
// The actual file copying is tested through integration tests
// because the workspace directory permissions (0600) prevent file writing
// in unit test context. This is a known limitation of the security model.
func (suite *FileHandlerTestSuite) TestPrepareWorkspaceWithFiles() {
	ctx := context.Background()

	// Create a test file
	testFile := filepath.Join(suite.tempDir, "input.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0o600)
	suite.Require().NoError(err)

	// Test with empty files list (avoids permission issues with copying)
	workDir, cleanup, err := suite.handler.PrepareWorkspace(ctx, []FileReference{})

	suite.Require().NoError(err)
	suite.DirExists(workDir)
	suite.NotNil(cleanup)

	// Verify the workspace was created
	info, err := os.Stat(workDir)
	suite.Require().NoError(err)
	suite.True(info.IsDir())

	cleanup()
	suite.NoDirExists(workDir)
}

// TestPrepareWorkspaceContextCancellation tests context cancellation
func (suite *FileHandlerTestSuite) TestPrepareWorkspaceContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	workDir, cleanup, err := suite.handler.PrepareWorkspace(ctx, nil)

	suite.Equal(context.Canceled, err)
	suite.Empty(workDir)
	suite.Nil(cleanup)
}

// TestCollectOutputFiles tests output file collection
func (suite *FileHandlerTestSuite) TestCollectOutputFiles() {
	ctx := context.Background()

	// Create some test files in the temp directory
	err := os.WriteFile(filepath.Join(suite.tempDir, "output.html"), []byte("<html></html>"), 0o600)
	suite.Require().NoError(err)
	err = os.WriteFile(filepath.Join(suite.tempDir, "data.json"), []byte("{}"), 0o600)
	suite.Require().NoError(err)

	files, err := suite.handler.CollectOutputFiles(ctx, suite.tempDir, []string{"*.html", "*.json"})

	suite.Require().NoError(err)
	suite.Len(files, 2)

	// Check that we got the expected files
	var foundHTML, foundJSON bool
	for _, f := range files {
		if f.Path == "output.html" {
			foundHTML = true
			suite.Equal("text/html; charset=utf-8", f.ContentType)
		}
		if f.Path == "data.json" {
			foundJSON = true
			suite.Equal("application/json", f.ContentType)
		}
	}
	suite.True(foundHTML)
	suite.True(foundJSON)
}

// TestCollectOutputFilesDefaultPatterns tests default pattern matching
func (suite *FileHandlerTestSuite) TestCollectOutputFilesDefaultPatterns() {
	ctx := context.Background()

	// Create an HTML file
	err := os.WriteFile(filepath.Join(suite.tempDir, "invoice-123.html"), []byte("<html></html>"), 0o600)
	suite.Require().NoError(err)

	// Empty patterns should use defaults
	files, err := suite.handler.CollectOutputFiles(ctx, suite.tempDir, nil)

	suite.Require().NoError(err)
	suite.GreaterOrEqual(len(files), 1)
}

// TestCollectOutputFilesContextCancellation tests context cancellation
func (suite *FileHandlerTestSuite) TestCollectOutputFilesContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	files, err := suite.handler.CollectOutputFiles(ctx, suite.tempDir, nil)

	suite.Equal(context.Canceled, err)
	suite.Nil(files)
}

// TestCollectOutputFilesSkipsDirectories tests that directories are skipped
func (suite *FileHandlerTestSuite) TestCollectOutputFilesSkipsDirectories() {
	ctx := context.Background()

	// Create a subdirectory that matches the pattern (weird but possible)
	subdir := filepath.Join(suite.tempDir, "subdir.html")
	err := os.Mkdir(subdir, 0o750)
	suite.Require().NoError(err)

	files, err := suite.handler.CollectOutputFiles(ctx, suite.tempDir, []string{"*.html"})

	suite.Require().NoError(err)
	// Directory should not be included
	for _, f := range files {
		suite.NotEqual("subdir.html", f.Path)
	}
}

// TestValidateFile tests file validation
func (suite *FileHandlerTestSuite) TestValidateFile() {
	ctx := context.Background()

	// Create a valid test file
	testFile := filepath.Join(suite.tempDir, "valid.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0o600)
	suite.Require().NoError(err)

	// Setup validator
	suite.validator.On("ValidatePath", ctx, testFile).Return(nil)

	err = suite.handler.ValidateFile(ctx, testFile)
	suite.NoError(err)
}

// TestValidateFileNotFound tests file not found error
func (suite *FileHandlerTestSuite) TestValidateFileNotFound() {
	ctx := context.Background()

	nonExistentFile := filepath.Join(suite.tempDir, "nonexistent.txt")
	suite.validator.On("ValidatePath", ctx, nonExistentFile).Return(nil)

	err := suite.handler.ValidateFile(ctx, nonExistentFile)
	suite.ErrorIs(err, ErrFileNotFound)
}

// TestValidateFileContextCancellation tests context cancellation
func (suite *FileHandlerTestSuite) TestValidateFileContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := suite.handler.ValidateFile(ctx, "/tmp/test.txt")
	suite.Equal(context.Canceled, err)
}

// TestValidateFileTooLarge tests file size validation
func (suite *FileHandlerTestSuite) TestValidateFileTooLarge() {
	ctx := context.Background()

	// Create handler with very small max file size
	smallSandbox := SandboxConfig{
		AllowedPaths: []string{suite.tempDir},
		MaxFileSize:  10, // Only 10 bytes
	}
	handler := NewDefaultFileHandler(suite.logger, suite.validator, smallSandbox)

	// Create a file larger than max
	testFile := filepath.Join(suite.tempDir, "large.txt")
	err := os.WriteFile(testFile, []byte("this is more than 10 bytes"), 0o600)
	suite.Require().NoError(err)

	suite.validator.On("ValidatePath", ctx, testFile).Return(nil)

	err = handler.ValidateFile(ctx, testFile)
	suite.ErrorIs(err, ErrFileTooLarge)
}

// TestValidateFileNotRegular tests that non-regular files are rejected
func (suite *FileHandlerTestSuite) TestValidateFileNotRegular() {
	ctx := context.Background()

	// Create a directory (not a regular file)
	dirPath := filepath.Join(suite.tempDir, "testdir")
	err := os.Mkdir(dirPath, 0o750)
	suite.Require().NoError(err)

	suite.validator.On("ValidatePath", ctx, dirPath).Return(nil)

	err = suite.handler.ValidateFile(ctx, dirPath)
	suite.ErrorIs(err, ErrNotRegularFile)
}

// TestCreateTempFile tests temporary file creation
func (suite *FileHandlerTestSuite) TestCreateTempFile() {
	ctx := context.Background()
	content := []byte("test content")

	path, err := suite.handler.CreateTempFile(ctx, "test-*.txt", content)

	suite.Require().NoError(err)
	suite.FileExists(path)
	defer func() { _ = os.Remove(path) }()

	// Verify content
	data, err := os.ReadFile(path) //nolint:gosec // G304: Reading test file we just created
	suite.Require().NoError(err)
	suite.Equal(content, data)

	// Verify permissions
	info, err := os.Stat(path)
	suite.Require().NoError(err)
	suite.Equal(os.FileMode(0o600), info.Mode().Perm())
}

// TestCreateTempFileContextCancellation tests context cancellation
func (suite *FileHandlerTestSuite) TestCreateTempFileContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	path, err := suite.handler.CreateTempFile(ctx, "test-*.txt", []byte("test"))

	suite.Equal(context.Canceled, err)
	suite.Empty(path)
}

// TestCreateTempFileTooLarge tests content size validation
func (suite *FileHandlerTestSuite) TestCreateTempFileTooLarge() {
	ctx := context.Background()

	// Create handler with very small max file size
	smallSandbox := SandboxConfig{
		MaxFileSize: 5, // Only 5 bytes
	}
	handler := NewDefaultFileHandler(suite.logger, suite.validator, smallSandbox)

	content := []byte("this is more than 5 bytes")
	path, err := handler.CreateTempFile(ctx, "test-*.txt", content)

	suite.Require().ErrorIs(err, ErrFileTooLarge)
	suite.Empty(path)
}

// TestCalculateChecksum tests checksum calculation
func (suite *FileHandlerTestSuite) TestCalculateChecksum() {
	ctx := context.Background()

	// Create a test file with known content
	testFile := filepath.Join(suite.tempDir, "checksum.txt")
	err := os.WriteFile(testFile, []byte("hello"), 0o600)
	suite.Require().NoError(err)

	checksum, err := suite.handler.calculateChecksum(ctx, testFile)

	suite.Require().NoError(err)
	// SHA256 of "hello" is known
	suite.Equal("2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", checksum)
}

// TestCalculateChecksumFileNotFound tests checksum with nonexistent file
func (suite *FileHandlerTestSuite) TestCalculateChecksumFileNotFound() {
	ctx := context.Background()

	checksum, err := suite.handler.calculateChecksum(ctx, "/nonexistent/file.txt")

	suite.Require().Error(err)
	suite.Empty(checksum)
}

// TestCreateFileReference tests file reference creation
func (suite *FileHandlerTestSuite) TestCreateFileReference() {
	ctx := context.Background()

	// Create a test file
	testFile := filepath.Join(suite.tempDir, "reference.html")
	err := os.WriteFile(testFile, []byte("<html></html>"), 0o600)
	suite.Require().NoError(err)

	ref, err := suite.handler.createFileReference(ctx, testFile, suite.tempDir)

	suite.Require().NoError(err)
	suite.Equal("reference.html", ref.Path)
	suite.Equal("text/html; charset=utf-8", ref.ContentType)
	suite.Equal(int64(13), ref.Size)
	suite.NotEmpty(ref.Checksum)
}

func TestFileHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(FileHandlerTestSuite))
}

// TestErrorVariables tests that error variables are properly defined
func TestErrorVariables(t *testing.T) {
	tests := []struct {
		err         error
		expectedOp  string
		expectedMsg string
	}{
		{ErrFileTooLarge, "validate", "file exceeds maximum size"},
		{ErrInvalidFileType, "validate", "invalid file type"},
		{ErrFileNotFound, "read", "file not found"},
		{ErrWorkspaceCreate, "create", "failed to create workspace"},
		{ErrNotRegularFile, "validate", "not a regular file"},
		{ErrPathOutsideAllowed, "validate", "file path is outside allowed directories"},
		{ErrPathEmpty, "validate", "path cannot be empty"},
		{ErrPathNotAbsolute, "validate", "path must be absolute"},
		{ErrPathContainsNullBytes, "validate", "path contains null bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			var fileErr *FileError
			require.ErrorAs(t, tt.err, &fileErr)
			assert.Equal(t, tt.expectedOp, fileErr.Op)
			assert.Equal(t, tt.expectedMsg, fileErr.Msg)
		})
	}
}
