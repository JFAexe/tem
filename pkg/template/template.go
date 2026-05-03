package template

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/JFAexe/tem/pkg/template/functions"
)

type Option = func(t *Template)

func WithDelims(left, right string) Option {
	return func(t *Template) {
		if left = strings.TrimSpace(left); left == "" {
			left = "{{"
		}

		if right = strings.TrimSpace(right); right == "" {
			left = "}}"
		}

		t.Delims(left, right)
	}
}

type Template struct {
	*template.Template
}

func New(name string, options ...Option) *Template {
	t := &Template{
		Template: template.New(name),
	}

	for _, option := range options {
		option(t)
	}

	t.Funcs(functions.FuncMap(t.Template))

	return t
}

func (t *Template) ParsePaths(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	for _, path := range paths {
		if err := t.ParsePath(path); err != nil {
			return fmt.Errorf("failed to parse paths: %w", err)
		}
	}

	return nil
}

func (t *Template) ParsePath(path string) (err error) {
	if path := strings.TrimSpace(path); path == "" {
		return nil
	}

	var paths []string

	if strings.ContainsAny(path, "*^!?[]{}") {
		if paths, err = doublestar.FilepathGlob(path); err != nil {
			return fmt.Errorf("failed to walk template includes glob %#q: %w", path, err)
		}
	} else {
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get abs path for template includes: %w", err)
		}

		if err = filepath.WalkDir(abs, func(p string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}

			if d.Type().IsRegular() {
				if p, e = filepath.Abs(p); e != nil {
					return fmt.Errorf("failed to get abs path for %#q: %w", d.Name(), err)
				}

				paths = append(paths, p)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to walk template includes path %#q: %w", path, err)
		}
	}

	if _, err := t.ParseFiles(paths...); err != nil {
		return fmt.Errorf("failed to parse template includes: %w", err)
	}

	return nil
}
