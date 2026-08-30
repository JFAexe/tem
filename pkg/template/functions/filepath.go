package functions

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/JFAexe/tem/pkg/convert"
)

type WalkInfo struct {
	Name    string `json:"name"     yaml:"name"     toml:"name"`
	Path    string `json:"path"     yaml:"path"     toml:"path"`
	RelPath string `json:"rel_path" yaml:"rel_path" toml:"rel_path"`
	AbsPath string `json:"abs_path" yaml:"abs_path" toml:"abs_path"`
	IsFile  bool   `json:"is_file"  yaml:"is_file"  toml:"is_file"`
	IsDir   bool   `json:"is_dir"   yaml:"is_dir"   toml:"is_dir"`
}

type Filepath struct{}

func (*Filepath) Separator() string {
	return string(filepath.Separator)
}

func (*Filepath) Clean(value any) string {
	return filepath.Clean(convert.ToString(value))
}

func (*Filepath) Abs(value any) (string, error) {
	return filepath.Abs(convert.ToString(value))
}

func (*Filepath) IsAbs(value any) bool {
	return filepath.IsAbs(convert.ToString(value))
}

func (*Filepath) Root(value any) string {
	str := convert.ToString(value)
	if str == "" {
		return ""
	}

	if vol := filepath.VolumeName(str); vol != "" {
		return vol
	}

	for {
		dir, _ := filepath.Split(str)

		if dir == "" || dir == str {
			break
		}

		str = filepath.Clean(dir)
	}

	if s := strings.TrimRight(str, `/\`); s != "" {
		return s
	}

	return str
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

func (*Filepath) TrimExt(value any) string {
	str := convert.ToString(value)
	if str == "" {
		return ""
	}

	var (
		name = filepath.Base(str)
		ext  = filepath.Ext(name)
	)

	if ext == "" || ext == name {
		return str
	}

	return strings.TrimSuffix(str, ext)
}

func (*Filepath) Join(values ...any) string {
	return filepath.Join(convert.ToStringSlice(values)...)
}

func (*Filepath) Split(value any) DirFile {
	dir, file := filepath.Split(convert.ToString(value))

	return DirFile{
		Dir:  dir,
		File: file,
	}
}

func (*Filepath) Rel(root, value any) (string, error) {
	return filepath.Rel(convert.ToString(root), convert.ToString(value))
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

func (*Filepath) Match(pattern, name any) (bool, error) {
	return doublestar.PathMatch(convert.ToString(pattern), convert.ToString(name))
}

func (*Filepath) Glob(value any) ([]string, error) {
	return doublestar.FilepathGlob(convert.ToString(value))
}

func (*Filepath) Walk(args ...any) ([]WalkInfo, error) {
	var (
		target string
		skip   bool
	)

	switch len(args) {
	case 0:
		return nil, ErrValueRequired
	case 1:
		target = convert.ToString(args[0])
	case 2:
		target = convert.ToString(args[1])
		skip = convert.ToBool(args[0])
	default:
		return nil, fmt.Errorf("%w: max is 2", ErrTooManyArguments)
	}

	if target = strings.TrimSpace(target); target == "" {
		return nil, ErrEmptyPath
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	var (
		entries       = make([]WalkInfo, 0)
		root, pattern = doublestar.SplitPattern(target)
	)

	if !strings.ContainsAny(pattern, "*^!?[]{}") {
		root = filepath.Join(root, pattern)
		pattern = "**"
	}

	if err := doublestar.GlobWalk(os.DirFS(filepath.Clean(root)), pattern, func(value string, d fs.DirEntry) (e error) {
		if skip && d.IsDir() {
			return nil
		}

		entry := WalkInfo{
			Name:    d.Name(),
			Path:    filepath.Join(root, value),
			RelPath: value,
			IsFile:  d.Type().IsRegular(),
			IsDir:   d.IsDir(),
		}

		if filepath.IsAbs(entry.Path) {
			entry.AbsPath = filepath.Clean(entry.Path)
		} else {
			entry.AbsPath = filepath.Join(wd, entry.Path)
		}

		entries = append(entries, entry)

		return nil
	}); err != nil {
		return nil, err
	}

	return entries, nil
}

func (*Filepath) Exists(value any) bool {
	_, ok := statPath(value)

	return ok
}

func (*Filepath) IsDir(value any) bool {
	info, ok := statPath(value)

	return ok && info.IsDir()
}

func (*Filepath) IsFile(value any) bool {
	info, ok := statPath(value)

	return ok && info.Mode().IsRegular()
}

func (*Filepath) IsSymlink(value any) bool {
	str := convert.ToString(value)
	if strings.TrimSpace(str) == "" {
		return false
	}

	info, err := os.Lstat(filepath.Clean(str))

	return err == nil && info.Mode()&fs.ModeSymlink != 0
}

func statPath(value any) (fs.FileInfo, bool) {
	str := convert.ToString(value)
	if strings.TrimSpace(str) == "" {
		return nil, false
	}

	info, err := os.Stat(filepath.Clean(str))

	return info, err == nil
}
