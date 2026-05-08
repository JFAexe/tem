package functions

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

var ErrEmptyPath = errors.New("can't walk empty path")

type WalkInfo struct {
	Name     string
	Path     string
	FullPath string
	IsFile   bool
	IsDir    bool
}

type Filepath struct{}

func (*Filepath) Clean(s string) string {
	return filepath.Clean(s)
}

func (*Filepath) Abs(s string) (string, error) {
	return filepath.Abs(s)
}

func (*Filepath) IsAbs(s string) bool {
	return filepath.IsAbs(s)
}

func (*Filepath) Base(s string) string {
	return filepath.Base(s)
}

func (*Filepath) Dir(s string) string {
	return filepath.Dir(s)
}

func (*Filepath) Ext(s string) string {
	return filepath.Ext(s)
}

func (*Filepath) Join(elems ...string) string {
	return filepath.Join(elems...)
}

func (*Filepath) Split(s string) []string {
	dir, file := filepath.Split(s)

	return []string{dir, file}
}

func (*Filepath) Match(pattern, name string) (bool, error) {
	return doublestar.Match(pattern, name)
}

func (*Filepath) Rel(target, base string) (string, error) {
	return filepath.Rel(target, base)
}

func (*Filepath) ToSlash(s string) string {
	return filepath.ToSlash(s)
}

func (*Filepath) FromSlash(s string) string {
	return filepath.FromSlash(s)
}

func (*Filepath) Volume(s string) string {
	return filepath.VolumeName(s)
}

func (*Filepath) Glob(s string) ([]string, error) {
	return doublestar.FilepathGlob(s)
}

func (*Filepath) Walk(root string, args ...bool) ([]WalkInfo, error) {
	if root = strings.TrimSpace(root); root == "" {
		return nil, ErrEmptyPath
	}

	var (
		entries []WalkInfo
		pattern string
		skipDir bool
	)

	for _, arg := range args {
		skipDir = arg
	}

	root, pattern = doublestar.SplitPattern(root)

	if !strings.ContainsAny(pattern, "*^!?[]{}") {
		root = filepath.Join(root, pattern)
		pattern = "**"
	}

	if err := doublestar.GlobWalk(os.DirFS(filepath.Clean(root)), pattern, func(p string, d fs.DirEntry) (e error) {
		if skipDir && d.IsDir() {
			return nil
		}

		info := WalkInfo{
			Name:   d.Name(),
			Path:   p,
			IsFile: d.Type().IsRegular(),
			IsDir:  d.IsDir(),
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

func (*Filepath) Exists(path string) bool {
	_, err := os.Stat(filepath.Clean(path))

	return err == nil
}

func (*Filepath) IsDir(path string) bool {
	stat, err := os.Stat(filepath.Clean(path))

	return err == nil && stat.Mode().IsDir()
}

func (*Filepath) IsFile(path string) bool {
	stat, err := os.Stat(filepath.Clean(path))

	return err == nil && stat.Mode().IsRegular()
}

func (*Filepath) IsSymlink(path string) bool {
	stat, err := os.Lstat(filepath.Clean(path))

	return err == nil && stat.Mode()&fs.ModeSymlink != 0
}
