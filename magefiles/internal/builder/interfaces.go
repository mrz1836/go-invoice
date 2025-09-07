package builder

import (
	"io"
	"os"
)

type ShellRunner interface {
	Run(cmd string, args ...string) error
	Output(cmd string, args ...string) (string, error)
}

type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	Stat(path string) (os.FileInfo, error)
	Create(path string) (*os.File, error)
	Open(path string) (*os.File, error)
	Chmod(path string, mode os.FileMode) error
	UserHomeDir() (string, error)
	IsNotExist(err error) bool
}

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

type FileCopier interface {
	Copy(dst io.Writer, src io.Reader) (int64, error)
}
