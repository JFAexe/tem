package template

import (
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/JFAexe/tem/pkg/env"
)

type Template struct {
	*template.Template
	envs env.Store
}

func New(envs map[string]string) *Template {
	t := &Template{
		Template: template.New("root_template"),
		envs:     make(env.Store),
	}

	maps.Copy(t.envs, envs)

	t.Funcs(CommonFuncs()).Funcs(EnvFuncs(t.envs)).Funcs(TemplateFuncs(t.Template))

	return t
}

func (t *Template) ParsePath(path string) error {
	if path := strings.TrimSpace(path); path == "" {
		return nil
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get abs path for template include files: %w", err)
	}

	if strings.ContainsAny(abs, "*?") || strings.Contains(abs, "**") {
		if _, err := t.ParseGlob(abs); err != nil {
			return fmt.Errorf("failed to parse glob template include files %#q: %w", abs, err)
		}

		return nil
	}

	return filepath.Walk(abs, func(p string, i fs.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if i.Mode().IsRegular() {
			if _, err := t.ParseFiles(p); err != nil {
				return fmt.Errorf("failed to parse template include file %#q: %w", p, err)
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
