package functions

import (
	"os"
	"path/filepath"

	"github.com/JFAexe/tem/pkg/convert"
)

type File struct{}

func FileVarargInit(n *File, args []any) (any, error) {
	return n.Content(convert.ToStringSlice(args)[0])
}

func (*File) Content(value any) (string, error) {
	absPath, err := filepath.Abs(convert.ToString(value))
	if err != nil {
		return "", err
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func (*File) Exists(value any) bool {
	return isFile(value)
}
