package functions

import (
	"fmt"
	"os"
	"path/filepath"
)

type WalkInfo struct {
	os.FileInfo
	Path     string
	FullPath string
}

type Filepath struct{}

func (Filepath) Clean(s string) string {
	return filepath.Clean(s)
}

func (Filepath) Abs(s string) (string, error) {
	return filepath.Abs(s)
}

func (Filepath) IsAbs(s string) bool {
	return filepath.IsAbs(s)
}

func (Filepath) Base(s string) string {
	return filepath.Base(s)
}

func (Filepath) Dir(s string) string {
	return filepath.Dir(s)
}

func (Filepath) Ext(s string) string {
	return filepath.Ext(s)
}

func (Filepath) Join(elems ...string) string {
	return filepath.Join(elems...)
}

func (Filepath) Split(s string) []string {
	dir, file := filepath.Split(s)

	return []string{dir, file}
}

func (Filepath) Match(pattern, name string) (bool, error) {
	return filepath.Match(pattern, name)
}

func (Filepath) Rel(target, base string) (string, error) {
	return filepath.Rel(target, base)
}

func (Filepath) ToSlash(s string) string {
	return filepath.ToSlash(s)
}

func (Filepath) FromSlash(s string) string {
	return filepath.FromSlash(s)
}

func (Filepath) VolumeName(s string) string {
	return filepath.VolumeName(s)
}

func (Filepath) Glob(s string) ([]string, error) {
	return filepath.Glob(s)
}

func (Filepath) Walk(root string, args ...any) ([]WalkInfo, error) {
	if root == "" {
		return nil, fmt.Errorf("can't walk empty path")
	}

	var (
		entries []WalkInfo
		skipDir bool
	)

	for _, arg := range args {
		if v, ok := arg.(bool); ok {
			skipDir = v
		}
	}

	if err := filepath.Walk(root, func(p string, i os.FileInfo, e error) error {
		if e != nil {
			return fmt.Errorf("failed to walk dir: %w", e)
		}

		if skipDir && i.IsDir() {
			return nil
		}

		info := WalkInfo{
			Path:     p,
			FileInfo: i,
		}

		if info.FullPath, e = filepath.Abs(info.Path); e != nil {
			return e
		}

		entries = append(entries, info)

		return nil
	}); err != nil {
		return nil, err
	}

	return entries, nil
}
