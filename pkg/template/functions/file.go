package functions

import (
	"os"
	"path/filepath"

	"github.com/JFAexe/tem/pkg/convert"
)

type File struct{}

func FileVarargInit(n *File, args []any) (any, error) {
	return n.Content(convert.ToStringList(args)[0])
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
	stat, err := os.Stat(filepath.Clean(path))

	return err == nil && stat.Mode().IsRegular()
}
