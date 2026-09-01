package functions

import (
	"path"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/JFAexe/tem/pkg/convert"
)

type DirFile struct {
	Dir  string `json:"dir"  yaml:"dir"  toml:"dir"`
	File string `json:"file" yaml:"file" toml:"file"`
}

func (p DirFile) String() string {
	return p.Dir + p.File
}

type Path struct{}

func (*Path) Separator() string {
	return "/"
}

func (*Path) Clean(value any) string {
	return path.Clean(convert.ToString(value))
}

func (*Path) IsAbs(value any) bool {
	return pathIsAbs(value)
}

func (*Path) Root(value any) string {
	str := convert.ToString(value)
	if str == "" {
		return ""
	}

	for len(str) > 1 {
		dir, _ := path.Split(str)

		if dir == "" {
			break
		}

		str = path.Clean(dir)
	}

	return str
}

func (*Path) Base(value any) string {
	return path.Base(convert.ToString(value))
}

func (*Path) Dir(value any) string {
	return path.Dir(convert.ToString(value))
}

func (*Path) Ext(value any) string {
	return path.Ext(convert.ToString(value))
}

func (*Path) TrimExt(value any) string {
	str := convert.ToString(value)
	if str == "" {
		return ""
	}

	var (
		name = path.Base(str)
		ext  = path.Ext(name)
	)

	if ext == "" || ext == name {
		return str
	}

	return strings.TrimSuffix(str, ext)
}

func (*Path) Join(values ...any) string {
	return path.Join(convert.ToStringSlice(values)...)
}

func (*Path) Split(value any) DirFile {
	dir, file := path.Split(convert.ToString(value))

	return DirFile{
		Dir:  dir,
		File: file,
	}
}

func (*Path) Match(pattern, name any) (bool, error) {
	return pathMatch(pattern, name)
}

func pathMatch(pattern, name any) (bool, error) {
	return doublestar.Match(convert.ToString(pattern), convert.ToString(name))
}

func pathIsAbs(value any) bool {
	return path.IsAbs(convert.ToString(value))
}
