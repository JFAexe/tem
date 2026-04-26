package functions

import "path"

type Path struct{}

func (Path) Clean(s string) string {
	return path.Clean(s)
}

func (Path) IsAbs(s string) bool {
	return path.IsAbs(s)
}
func (Path) Base(s string) string {
	return path.Base(s)
}

func (Path) Dir(s string) string {
	return path.Dir(s)
}

func (Path) Ext(s string) string {
	return path.Ext(s)
}

func (Path) Join(elems ...string) string {
	return path.Join(elems...)
}

func (Path) Split(s string) []string {
	dir, file := path.Split(s)

	return []string{dir, file}
}

func (Path) Match(pattern, name string) (bool, error) {
	return path.Match(pattern, name)
}
