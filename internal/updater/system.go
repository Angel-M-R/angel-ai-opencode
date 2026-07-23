package updater

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// File is the temporary-file seam needed by update application.
type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
	Name() string
}

// FileSystem isolates executable discovery and replacement operations.
type FileSystem interface {
	Executable() (string, error)
	EvalSymlinks(string) (string, error)
	Open(string) (io.ReadCloser, error)
	Stat(string) (fs.FileInfo, error)
	CreateTemp(string, string) (File, error)
	Chmod(string, fs.FileMode) error
	Rename(string, string) error
	Remove(string) error
}

// Process isolates argument/environment capture and process replacement.
type Process interface {
	Args() []string
	Environ() []string
	Exec(string, []string, []string) error
}

type osFileSystem struct{}

func (osFileSystem) Executable() (string, error)              { return os.Executable() }
func (osFileSystem) EvalSymlinks(path string) (string, error) { return filepath.EvalSymlinks(path) }
func (osFileSystem) Open(name string) (io.ReadCloser, error)  { return os.Open(name) }
func (osFileSystem) Stat(name string) (fs.FileInfo, error)    { return os.Stat(name) }
func (osFileSystem) CreateTemp(dir, pattern string) (File, error) {
	return os.CreateTemp(dir, pattern)
}
func (osFileSystem) Chmod(name string, mode fs.FileMode) error { return os.Chmod(name, mode) }
func (osFileSystem) Rename(oldPath, newPath string) error      { return os.Rename(oldPath, newPath) }
func (osFileSystem) Remove(name string) error                  { return os.Remove(name) }

type osProcess struct{}

func (osProcess) Args() []string    { return append([]string(nil), os.Args...) }
func (osProcess) Environ() []string { return append([]string(nil), os.Environ()...) }
func (osProcess) Exec(path string, args, environment []string) error {
	return syscall.Exec(path, args, environment)
}
