package functions

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/bmatcuk/doublestar/v4"
)

var ErrEmptyPath = errors.New("can't walk empty path")

type WalkInfo struct {
	Name    string
	Path    string
	RelPath string
	AbsPath string
	IsFile  bool
	IsDir   bool
}

type Filepath struct{}

func (*Filepath) Clean(value any) string {
	return filepath.Clean(convert.ToString(value))
}

func (*Filepath) Abs(value any) (string, error) {
	return filepath.Abs(convert.ToString(value))
}

func (*Filepath) IsAbs(value any) bool {
	return filepath.IsAbs(convert.ToString(value))
}

func (*Filepath) Base(value any) string {
	return filepath.Base(convert.ToString(value))
}

func (*Filepath) Dir(value any) string {
	return filepath.Dir(convert.ToString(value))
}

func (*Filepath) Ext(value any) string {
	return filepath.Ext(convert.ToString(value))
}

func (*Filepath) Join(values ...string) string {
	return filepath.Join(values...)
}

func (*Filepath) Split(value any) []string {
	dir, file := filepath.Split(convert.ToString(value))

	return []string{dir, file}
}

func (*Filepath) Match(pattern, name any) (bool, error) {
	return doublestar.Match(convert.ToString(pattern), convert.ToString(name))
}

func (*Filepath) Rel(target, base any) (string, error) {
	return filepath.Rel(convert.ToString(target), convert.ToString(base))
}

func (*Filepath) ToSlash(value any) string {
	return filepath.ToSlash(convert.ToString(value))
}

func (*Filepath) FromSlash(value any) string {
	return filepath.FromSlash(convert.ToString(value))
}

func (*Filepath) Volume(value any) string {
	return filepath.VolumeName(convert.ToString(value))
}

func (*Filepath) Glob(value any) ([]string, error) {
	return doublestar.FilepathGlob(convert.ToString(value))
}

func (*Filepath) Walk(root any, args ...any) ([]WalkInfo, error) {
	rp := convert.ToString(root)

	if rp = strings.TrimSpace(rp); rp == "" {
		return nil, ErrEmptyPath
	}

	var (
		entries []WalkInfo
		pattern string
		skipDir bool
	)

	for _, arg := range args {
		skipDir = convert.ToBool(arg)
	}

	rp, pattern = doublestar.SplitPattern(rp)

	if !strings.ContainsAny(pattern, "*^!?[]{}") {
		rp = filepath.Join(rp, pattern)
		pattern = "**"
	}

	if err := doublestar.GlobWalk(os.DirFS(filepath.Clean(rp)), pattern, func(value string, d fs.DirEntry) (e error) {
		if skipDir && d.IsDir() {
			return nil
		}

		entry := WalkInfo{
			Name:    d.Name(),
			Path:    filepath.Join(rp, value),
			RelPath: value,
			IsFile:  d.Type().IsRegular(),
			IsDir:   d.IsDir(),
		}

		if entry.AbsPath, e = filepath.Abs(entry.Path); e != nil {
			return e
		}

		entries = append(entries, entry)

		return nil
	}); err != nil {
		return nil, err
	}

	return entries, nil
}

func (*Filepath) Exists(value any) bool {
	_, err := os.Stat(filepath.Clean(convert.ToString(value)))

	return err == nil
}

func (*Filepath) IsDir(value any) bool {
	stat, err := os.Stat(filepath.Clean(convert.ToString(value)))

	return err == nil && stat.Mode().IsDir()
}

func (*Filepath) IsFile(value any) bool {
	return isFile(value)
}

func (*Filepath) IsSymlink(value any) bool {
	stat, err := os.Lstat(filepath.Clean(convert.ToString(value)))

	return err == nil && stat.Mode()&fs.ModeSymlink != 0
}

func isFile(value any) bool {
	stat, err := os.Stat(filepath.Clean(convert.ToString(value)))

	return err == nil && stat.Mode().IsRegular()
}
