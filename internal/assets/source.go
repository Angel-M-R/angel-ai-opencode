// Package assets defines the runtime source for installable assets.
package assets

import (
	"fmt"
	"io/fs"
	"os"
)

// Source provides read-only access to an asset tree. Destination inspection
// and mutation remain the responsibility of callers using the OS filesystem.
type Source struct {
	fsys     fs.FS
	name     string
	embedded bool
}

// New constructs a source backed by fsys.
func New(fsys fs.FS, name string) Source {
	return Source{fsys: fsys, name: name}
}

// Embedded constructs a source backed by assets compiled into the executable.
func Embedded(fsys fs.FS) Source {
	return Source{fsys: fsys, name: "embedded assets", embedded: true}
}

// Directory constructs a source rooted at directory.
func Directory(directory string) Source {
	return Source{fsys: os.DirFS(directory), name: directory}
}

// FS exposes the read-only filesystem for generic fs operations.
func (source Source) FS() fs.FS {
	return source.fsys
}

// Name describes the source in diagnostics.
func (source Source) Name() string {
	if source.name == "" {
		return "assets"
	}
	return source.name
}

// ReadFile reads one asset relative to the source root.
func (source Source) ReadFile(name string) ([]byte, error) {
	if source.fsys == nil {
		return nil, fmt.Errorf("asset source is not configured")
	}
	return fs.ReadFile(source.fsys, name)
}

// ReadDir reads one asset directory relative to the source root.
func (source Source) ReadDir(name string) ([]fs.DirEntry, error) {
	if source.fsys == nil {
		return nil, fmt.Errorf("asset source is not configured")
	}
	return fs.ReadDir(source.fsys, name)
}

// Stat returns metadata for one asset relative to the source root.
func (source Source) Stat(name string) (fs.FileInfo, error) {
	if source.fsys == nil {
		return nil, fmt.Errorf("asset source is not configured")
	}
	return fs.Stat(source.fsys, name)
}

// WalkDir walks an asset subtree relative to the source root.
func (source Source) WalkDir(root string, fn fs.WalkDirFunc) error {
	if source.fsys == nil {
		return fmt.Errorf("asset source is not configured")
	}
	return fs.WalkDir(source.fsys, root, fn)
}

// FileMode returns the destination permissions for an asset file. go:embed
// does not retain source permissions, so embedded regular files use 0644;
// directory-backed overrides retain their existing permissions.
func (source Source) FileMode(name string) (fs.FileMode, error) {
	info, err := source.Stat(name)
	if err != nil {
		return 0, err
	}
	if source.embedded && info.Mode().IsRegular() {
		return 0o644, nil
	}
	return info.Mode().Perm(), nil
}
