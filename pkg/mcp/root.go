// Package mcp serves the Theory of Change tools to AI agents over the Model
// Context Protocol. Every tool is a thin adapter over pkg/service and the
// domain packages, so the server and the command line behave identically.
//
// Tools take a theory inline (the "document" argument) or by path. A path is
// confined to the server's root directory: it may not climb out with "..", and
// a symbolic link that leads outside the root is refused.
package mcp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrOutsideRoot is returned for a path that leaves the server's root.
var ErrOutsideRoot = errors.New("path is outside the server root")

// Root confines file access to one directory.
type Root struct {
	dir string
}

// NewRoot returns a root for dir, which must exist. Symbolic links in dir
// itself are resolved, so the confinement test compares real paths.
func NewRoot(dir string) (*Root, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("mcp: root %q: %w", dir, err)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("mcp: root %q: %w", dir, err)
	}
	info, err := os.Stat(real)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("mcp: root %q is not a directory", dir)
	}
	return &Root{dir: real}, nil
}

// Dir returns the root directory.
func (r *Root) Dir() string { return r.dir }

// Resolve turns a path from a tool call into a real path inside the root.
// A relative path is taken from the root. A path with a ".." element is
// refused outright, as is any path that, once symbolic links are followed,
// lies outside the root. The file need not exist, but its directory must,
// so a tool can create a file in place.
func (r *Root) Resolve(path string) (string, error) {
	if path == "" {
		return "", errors.New("mcp: empty path")
	}
	if strings.ContainsRune(path, 0) {
		return "", fmt.Errorf("%w: %q", ErrOutsideRoot, path)
	}
	for _, part := range strings.FieldsFunc(path, func(c rune) bool { return c == '/' || c == filepath.Separator }) {
		if part == ".." {
			return "", fmt.Errorf("%w: %q has a \"..\" element", ErrOutsideRoot, path)
		}
	}
	full := path
	if !filepath.IsAbs(full) {
		full = filepath.Join(r.dir, full)
	}
	full = filepath.Clean(full)

	// Follow links for the longest part that exists, then re-attach the rest.
	existing, rest := full, ""
	for {
		real, err := filepath.EvalSymlinks(existing)
		if err == nil {
			resolved := filepath.Join(real, rest)
			if !r.contains(resolved) {
				return "", fmt.Errorf("%w: %q", ErrOutsideRoot, path)
			}
			return resolved, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("mcp: resolve %q: %w", path, err)
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			return "", fmt.Errorf("%w: %q", ErrOutsideRoot, path)
		}
		rest = filepath.Join(filepath.Base(existing), rest)
		existing = parent
	}
}

func (r *Root) contains(path string) bool {
	rel, err := filepath.Rel(r.dir, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
