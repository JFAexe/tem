package functions

import (
	"os"
	"path/filepath"
)

type File struct{}

func FileNamespace() func(...any) (any, error) {
	n := new(File)

	return func(args ...any) (any, error) {
		if len(args) > 0 {
			return n.Content(ToStringList(args)[0])
		}

		return n, nil
	}
}

func (*File) Content(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func (*File) Exists(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return stat.Mode().IsRegular()
	}

	return false
}
