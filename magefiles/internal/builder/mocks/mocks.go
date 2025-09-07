package mocks

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type MockShellRunner struct {
	RunCalls    [][]string
	OutputCalls [][]string
	RunError    error
	OutputValue string
	OutputError error
}

func (m *MockShellRunner) Run(cmd string, args ...string) error {
	call := append([]string{cmd}, args...)
	m.RunCalls = append(m.RunCalls, call)
	return m.RunError
}

func (m *MockShellRunner) Output(cmd string, args ...string) (string, error) {
	call := append([]string{cmd}, args...)
	m.OutputCalls = append(m.OutputCalls, call)
	return m.OutputValue, m.OutputError
}

func (m *MockShellRunner) Reset() {
	m.RunCalls = nil
	m.OutputCalls = nil
	m.RunError = nil
	m.OutputValue = ""
	m.OutputError = nil
}

type MockFileSystem struct {
	Files        map[string][]byte
	Permissions  map[string]os.FileMode
	StatError    error
	CreateError  error
	OpenError    error
	ChmodError   error
	MkdirError   error
	RemoveError  error
	HomeDirPath  string
	HomeDirError error
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:       make(map[string][]byte),
		Permissions: make(map[string]os.FileMode),
		HomeDirPath: "/home/test",
	}
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if m.MkdirError != nil {
		return m.MkdirError
	}
	m.Files[path] = nil // Directory marker
	m.Permissions[path] = perm
	return nil
}

func (m *MockFileSystem) Remove(path string) error {
	if m.RemoveError != nil {
		return m.RemoveError
	}
	delete(m.Files, path)
	delete(m.Permissions, path)
	return nil
}

func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	if m.StatError != nil {
		return nil, m.StatError
	}
	if _, exists := m.Files[path]; exists {
		return &MockFileInfo{name: path}, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Create(path string) (*os.File, error) {
	if m.CreateError != nil {
		return nil, m.CreateError
	}
	m.Files[path] = []byte{}
	return &os.File{}, nil
}

func (m *MockFileSystem) Open(path string) (*os.File, error) {
	if m.OpenError != nil {
		return nil, m.OpenError
	}
	if _, exists := m.Files[path]; !exists {
		return nil, os.ErrNotExist
	}
	return &os.File{}, nil
}

func (m *MockFileSystem) Chmod(path string, mode os.FileMode) error {
	if m.ChmodError != nil {
		return m.ChmodError
	}
	m.Permissions[path] = mode
	return nil
}

func (m *MockFileSystem) UserHomeDir() (string, error) {
	return m.HomeDirPath, m.HomeDirError
}

func (m *MockFileSystem) IsNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

type MockFileInfo struct {
	name string
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return 0 }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *MockFileInfo) IsDir() bool        { return false }
func (m *MockFileInfo) Sys() any           { return nil }

type MockLogger struct {
	Messages []string
}

func (m *MockLogger) Println(v ...any) {
	m.Messages = append(m.Messages, strings.TrimSpace(fmt.Sprint(v...)))
}

func (m *MockLogger) Printf(format string, v ...any) {
	m.Messages = append(m.Messages, strings.TrimSpace(fmt.Sprintf(format, v...)))
}

func (m *MockLogger) Reset() {
	m.Messages = nil
}

type MockFileCopier struct {
	CopyError   error
	BytesCopied int64
}

func (m *MockFileCopier) Copy(dst io.Writer, src io.Reader) (int64, error) {
	if m.CopyError != nil {
		return 0, m.CopyError
	}
	return m.BytesCopied, nil
}
