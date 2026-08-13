package functions

import (
	"path"

	"github.com/JFAexe/tem/pkg/convert"
)

type Path struct{}

func (*Path) Clean(value any) string {
	return path.Clean(convert.ToString(value))
}

func (*Path) IsAbs(value any) bool {
	return path.IsAbs(convert.ToString(value))
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

func (*Path) Join(values ...any) string {
	return path.Join(convert.ToStringSlice(values)...)
}

func (*Path) Split(value any) []string {
	dir, file := path.Split(convert.ToString(value))

	return []string{dir, file}
}

func (*Path) Match(pattern, name any) (bool, error) {
	return path.Match(convert.ToString(pattern), convert.ToString(name))
}
