package template

import (
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/JFAexe/tem/pkg/env"
	"github.com/JFAexe/tem/pkg/template/functions"
)

type Option = func(t *Template)

func WithEnvs(envs env.Map) Option {
	return func(t *Template) {
		maps.Copy(t.envs, envs)
	}
}

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
	envs env.Store
}

func New(name string, options ...Option) *Template {
	t := &Template{
		Template: template.New(name),
		envs:     make(env.Store),
	}

	for _, option := range options {
		option(t)
	}

	t.Funcs(functions.FuncMap(t.Template, t.envs))

	return t
}

func (t *Template) ParsePath(path string) error {
	if path := strings.TrimSpace(path); path == "" {
		return nil
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get abs path for template includes: %w", err)
	}

	if strings.ContainsAny(abs, "*?") || strings.Contains(abs, "**") {
		if _, err := t.ParseGlob(abs); err != nil {
			return fmt.Errorf("failed to parse glob template includes %#q: %w", abs, err)
		}

		return nil
	}

	return filepath.Walk(abs, func(p string, i fs.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if i.Mode().IsRegular() {
			if _, err := t.ParseFiles(p); err != nil {
				return fmt.Errorf("failed to parse template include %#q: %w", p, err)
			}
		}

		return nil
	})
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
