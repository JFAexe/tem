package functions

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func Namespace[T any](n T) func() any {
	return func() any {
		return n
	}
}

func VarargNamespace[T any](n T, fn func(T, []any) (any, error)) func(...any) (any, error) {
	return func(args ...any) (any, error) {
		if len(args) > 0 {
			return fn(n, args)
		}

		return n, nil
	}
}

func FuncMap(t *template.Template) template.FuncMap {
	runeFuncs := NewRuneFuncs()

	return template.FuncMap{
		"inline":   Inline(t),
		"render":   Render(t),
		"ternary":  Ternary,
		"pwd":      os.Getwd,
		"hostname": os.Hostname,
		"env":      VarargNamespace(new(Env), EnvVarargInit),
		"file":     VarargNamespace(new(File), FileVarargInit),
		"filepath": Namespace(new(Filepath)),
		"path":     Namespace(new(Path)),
		"string":   Namespace(new(String)),
		"regex":    Namespace(NewRegexFuncs()),
		"math":     Namespace(new(Math)),
		"time":     Namespace(new(Time)),
		"data":     Namespace(new(Data)),
		"rune":     Namespace(runeFuncs),
		"random":   Namespace(NewRandomFuncs(runeFuncs)),
		"map":      VarargNamespace(new(Map), MapVarargInit),
		"list":     VarargNamespace(new(List), ListVarargInit),
		"to":       Namespace(new(Convert)),
	}
}

func Ternary(truthy, falsy any, cond bool) any {
	if cond {
		return truthy
	}

	return falsy
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
