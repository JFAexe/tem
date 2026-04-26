package functions

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/JFAexe/tem/pkg/env"
)

func FuncMap(t *template.Template, e env.Store) template.FuncMap {
	var (
		envFuncs      = &Env{envs: e}
		fileFuncs     = new(File)
		filepathFuncs = new(Filepath)
		pathFuncs     = new(Path)
		stringFuncs   = new(String)
		regexFuncs    = &Regex{cache: make(map[string]*regexp.Regexp)}
		timeFuncs     = new(Time)
		dataFuncs     = &Data{envs: e}
		mapFuncs      = new(Map)
		listFuncs     = new(List)
	)

	return template.FuncMap{
		"inline":   Inline(t),
		"render":   Render(t),
		"ternary":  Ternary,
		"pwd":      os.Getwd,
		"hostname": os.Hostname,
		"env": func(args ...any) any {
			if len(args) > 0 {
				return envFuncs.Get(ToStringList(args)[0])
			}

			return envFuncs
		},
		"file": func(args ...any) (any, error) {
			if len(args) > 0 {
				return fileFuncs.Content(ToStringList(args)[0])
			}

			return fileFuncs, nil
		},
		"filepath": func() any { return filepathFuncs },
		"path":     func() any { return pathFuncs },
		"string":   func() any { return stringFuncs },
		"regex":    func() any { return regexFuncs },
		"time":     func() any { return timeFuncs },
		"data":     func() any { return dataFuncs },
		"map": func(args ...any) (any, error) {
			if len(args) > 0 {
				return mapFuncs.New(args...)
			}

			return mapFuncs, nil
		},
		"list": func(args ...any) any {
			if len(args) > 0 {
				return listFuncs.New(args...)
			}

			return listFuncs
		},
		"toAny":        ToAny,
		"toString":     ToString,
		"toList":       ToList,
		"toStringList": ToStringList,
	}
}

func Render(t *template.Template) func(name string, data ...any) (string, error) {
	return func(name string, data ...any) (string, error) {
		clone, err := t.Clone()
		if err != nil {
			return "", err
		}

		return render(clone, name, data...)
	}
}

func Inline(t *template.Template) func(value string, data ...any) (string, error) {
	return func(value string, data ...any) (string, error) {
		clone, err := t.Clone()
		if err != nil {
			return "", err
		}

		name := fmt.Sprint("inline_", strings.ToLower(rand.Text()))

		if clone, err = clone.New(name).Parse(value); err != nil {
			return "", err
		}

		return render(clone, name, data...)
	}
}

func Ternary(truthy, falsy any, cond bool) any {
	if cond {
		return truthy
	}

	return falsy
}

func ToAny(v any) any {
	return v
}

func ToString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.GoStringer:
		return v.GoString()
	case fmt.Stringer:
		return v.String()
	}

	switch rv := reflect.ValueOf(value); rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return ""
		}

		return ToString(rv.Elem().Interface())
	}

	return fmt.Sprint(value)
}

func ToList(value any) []any {
	return toTypedList(value, ToAny)
}

func ToStringList(value any) []string {
	return toTypedList(value, ToString)
}

func toTypedList[T any](value any, fn func(any) T) []T {
	if value == nil {
		return make([]T, 0)
	}

	if v, ok := value.([]T); ok {
		return slices.Clone(v)
	}

	rv := reflect.ValueOf(value)

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return []T{fn(value)}
	}

	out := make([]T, 0, rv.Len())

	for i := range rv.Len() {
		if item := rv.Index(i).Interface(); item != nil {
			out = append(out, fn(item))
		}
	}

	return out
}

func render(t *template.Template, name string, data ...any) (string, error) {
	var (
		ctx any
		err error
	)

	if len(data) == 1 {
		ctx = data[0]
	} else if ctx, err = new(Map).New(data...); err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if err = t.Lookup(name).Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}
