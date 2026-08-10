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

func NamespaceVararg[T any](n T, fn func(T, []any) (any, error)) func(...any) (any, error) {
	return func(args ...any) (any, error) {
		if len(args) > 0 {
			return fn(n, args)
		}

		return n, nil
	}
}

func FuncMap(t *template.Template) template.FuncMap {
	var (
		regexFuncs    = NewRegexFuncs()
		runeFuncs     = NewRuneFuncs()
		randomFuncs   = NewRandomFuncs(runeFuncs)
		envFuncs      = new(Env)
		fileFuncs     = new(File)
		filepathFuncs = new(Filepath)
		pathFuncs     = new(Path)
		stringFuncs   = new(String)
		mathFuncs     = new(Math)
		timeFuncs     = new(Time)
		dataFuncs     = new(Data)
		mapFuncs      = new(Map)
		listFuncs     = new(List)
		convertFuncs  = new(Convert)
	)

	return template.FuncMap{
		"inline":   Inline(t),
		"render":   Render(t),
		"ternary":  Ternary,
		"pwd":      os.Getwd,
		"hostname": os.Hostname,
		"env":      NamespaceVararg(envFuncs, EnvVarargInit),
		"file":     NamespaceVararg(fileFuncs, FileVarargInit),
		"filepath": Namespace(filepathFuncs),
		"path":     Namespace(pathFuncs),
		"string":   Namespace(stringFuncs),
		"regex":    Namespace(regexFuncs),
		"math":     Namespace(mathFuncs),
		"time":     Namespace(timeFuncs),
		"data":     Namespace(dataFuncs),
		"rune":     Namespace(runeFuncs),
		"random":   Namespace(randomFuncs),
		"map":      NamespaceVararg(mapFuncs, MapVarargInit),
		"list":     NamespaceVararg(listFuncs, ListVarargInit),
		"to":       Namespace(convertFuncs),
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
