package template

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/JFAexe/tem/pkg/template/functions"
)

var ErrNoMatches = errors.New("pattern matches no files")

const (
	DefaultLeftDelim  = "[["
	DefaultRightDelim = "]]"
)

type Option func(*Template)

func WithDelims(left, right string) Option {
	return func(t *Template) {
		if left = strings.TrimSpace(left); left == "" {
			left = DefaultLeftDelim
		}

		if right = strings.TrimSpace(right); right == "" {
			right = DefaultRightDelim
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

func (t *Template) ParseGlob(pattern string) (*template.Template, error) {
	files, err := expandFilePaths([]string{pattern})
	if err != nil {
		return nil, err
	}

	return t.ParseFiles(files...)
}

func (t *Template) ParseFS(fsys fs.FS, patterns ...string) (*template.Template, error) {
	names, err := expandFSPatterns(fsys, patterns)
	if err != nil {
		return nil, err
	}

	return t.Template.ParseFS(fsys, names...)
}

func (t *Template) ParsePaths(paths ...string) (*template.Template, error) {
	for _, p := range paths {
		if _, err := t.ParsePath(p); err != nil {
			return nil, err
		}
	}

	return t.Template, nil
}

func (t *Template) ParsePath(path string) (*template.Template, error) {
	if path = strings.TrimSpace(path); path == "" {
		return t.Template, nil
	}

	switch info, err := os.Stat(path); {
	case os.IsNotExist(err) && hasGlobMeta(path):
		files, err := expandFilePaths([]string{path})
		if err != nil {
			return nil, err
		}

		if _, e := t.ParseFiles(files...); e != nil {
			return nil, fmt.Errorf("failed to glob path %#q: %w", path, e)
		}

		return t.Template, nil
	case err == nil:
		switch {
		case info.Mode().IsRegular():
			if _, e := t.ParseFiles(path); e != nil {
				return nil, fmt.Errorf("failed to parse file %#q: %w", path, e)
			}

			return t.Template, nil
		case info.IsDir():
			files, e := walkRegularFiles(path)
			if e != nil {
				return nil, e
			}

			if _, e = t.ParseFiles(files...); e != nil {
				return nil, fmt.Errorf("failed to parse dir %#q: %w", path, e)
			}

			return t.Template, nil
		default:
			return nil, fmt.Errorf("failed to parse path %#q: not a regular file or directory", path)
		}
	default:
		return nil, fmt.Errorf("template path %#q: %w", path, err)
	}
}

func expandFilePaths(patterns []string) (out []string, err error) {
	seen := make(map[string]struct{})

	for _, p := range patterns {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}

		if !hasGlobMeta(p) {
			if abs, err := filepath.Abs(p); err == nil {
				p = abs
			}

			if _, ok := seen[p]; !ok {
				seen[p] = struct{}{}

				out = append(out, p)
			}

			continue
		}

		matches, err := doublestar.FilepathGlob(p, doublestar.WithFailOnIOErrors(), doublestar.WithFilesOnly())
		if err != nil {
			return nil, fmt.Errorf("failed to glob path %#q: %w", p, err)
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf("failed to glob path %#q: %w", p, ErrNoMatches)
		}

		for _, m := range matches {
			if abs, err := filepath.Abs(m); err == nil {
				m = abs
			}

			if _, ok := seen[m]; ok {
				continue
			}

			seen[m] = struct{}{}

			out = append(out, m)
		}
	}

	return out, nil
}

func expandFSPatterns(fsys fs.FS, patterns []string) (out []string, err error) {
	seen := make(map[string]struct{})

	for _, p := range patterns {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}

		if p = filepath.ToSlash(p); !hasGlobMeta(p) {
			if _, err := fs.Stat(fsys, p); err != nil {
				return nil, fmt.Errorf("failed to stat path %#q: %w", p, err)
			}

			if _, ok := seen[p]; !ok {
				seen[p] = struct{}{}

				out = append(out, p)
			}

			continue
		}

		matches, err := doublestar.Glob(fsys, p, doublestar.WithFailOnIOErrors(), doublestar.WithFilesOnly())
		if err != nil {
			return nil, fmt.Errorf("failed to glob path %#q: %w", p, err)
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf("failed to glob path %#q: %w", p, ErrNoMatches)
		}

		for _, m := range matches {
			if _, ok := seen[m]; ok {
				continue
			}

			seen[m] = struct{}{}

			out = append(out, m)
		}
	}

	return out, nil
}

func walkRegularFiles(root string) (files []string, err error) {
	if err = filepath.WalkDir(root, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		if d.Type().IsRegular() {
			files = append(files, p)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk %#q: %w", root, err)
	}

	return files, nil
}

func hasGlobMeta(s string) bool {
	return strings.ContainsAny(s, "*?[]{}")
}
