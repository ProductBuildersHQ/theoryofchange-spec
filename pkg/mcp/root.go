// Package mcp serves the impact model tools to AI agents over the Model
// Context Protocol. Every tool is a thin adapter over pkg/service and the
// domain packages, so the server and the command line behave identically.
//
// Tools take an impact model inline (the "document" argument) or by path. A path is
// confined to the server's root directory: it may not climb out with "..", and
// a symbolic link that leads outside the root is refused.
package mcp

import (
	"fmt"

	oscfs "github.com/grokify/oscompat/fs"
)

// ErrOutsideRoot is returned for a path that leaves the server's root. It is
// the oscompat error, so errors.Is works with either name.
var ErrOutsideRoot = oscfs.ErrOutsideRoot

// Root confines file access to one directory.
type Root struct {
	dir string
}

// NewRoot returns a root for dir, which must exist. Symbolic links in dir
// itself are resolved, so the confinement test compares real paths.
func NewRoot(dir string) (*Root, error) {
	real, err := oscfs.RealDir(dir)
	if err != nil {
		return nil, fmt.Errorf("mcp: root: %w", err)
	}
	return &Root{dir: real}, nil
}

// Dir returns the root directory.
func (r *Root) Dir() string { return r.dir }

// Resolve turns a path from a tool call into a real path inside the root.
// A relative path is taken from the root. A path with a ".." element, a rooted
// or drive path outside the root, and a path that leads outside once symbolic
// links are followed are all refused, with the same answer on every platform.
// The file need not exist, so a tool can create a file in place.
func (r *Root) Resolve(path string) (string, error) {
	resolved, err := oscfs.ResolveInRoot(r.dir, path)
	if err != nil {
		return "", fmt.Errorf("mcp: %w", err)
	}
	return resolved, nil
}
