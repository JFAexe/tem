package functions

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
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
		runeFuncs   = new(Rune)
		randomFuncs = NewRandom(runeFuncs)
	)

	return template.FuncMap{
		"pwd":      os.Getwd,
		"hostname": os.Hostname,
		"ternary":  Ternary,
		"default":  Default,
		"indexOr":  IndexOr,
		"set":      Set,
		"unset":    Unset,
		"isSet":    IsSet,
		"inline":   Inline(t),
		"include":  Include(t),
		"file":     File,
		"to":       Namespace(new(Convert)),
		"data":     Namespace(new(Data)),
		"env":      NamespaceVararg(new(Env), EnvVarargInit),
		"filepath": Namespace(new(Filepath)),
		"list":     NamespaceVararg(new(List), ListVarargInit),
		"map":      NamespaceVararg(new(Map), MapVarargInit),
		"math":     Namespace(new(Math)),
		"path":     Namespace(new(Path)),
		"random":   Namespace(randomFuncs),
		"regex":    Namespace(new(Regex)),
		"rune":     Namespace(runeFuncs),
		"string":   Namespace(new(String)),
		"time":     Namespace(new(Time)),
	}
}

func Ternary(truthy, falsy, condition any) any {
	if convert.ToBool(condition) {
		return truthy
	}

	return falsy
}

func Default(val, def any) any {
	if val == nil {
		return def
	}

	v := reflect.ValueOf(val)

	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return def
		}

		v = v.Elem()
	}

	if v.IsZero() {
		return def
	}

	return val
}

func IndexOr(item, def any, args ...any) any {
	v := reflect.ValueOf(item)

	if len(args) == 0 {
		if !v.IsValid() {
			return def
		}

		return item
	}

	for _, arg := range args {
		next, err := reflection.Lookup(v, arg)
		if err != nil {
			return def
		}

		v = next
	}

	return v.Interface()
}

func Set(item, value any, args ...any) (any, error) {
	if len(args) == 0 {
		return item, fmt.Errorf("set requires at least one index/key")
	}

	v := reflect.ValueOf(item)

	for i := range len(args) - 1 {
		next, err := reflection.Lookup(v, args[i])
		if err != nil {
			if errors.Is(err, reflection.ErrKeyMissing) {
				return item, fmt.Errorf("map key not found at step %d (cannot set nested value in nil map)", i)
			}

			return item, fmt.Errorf("at step %d: %w", i, err)
		}

		v = next
	}

	var err error

	if v, err = reflection.IndirectValue(v); err != nil {
		return item, fmt.Errorf("cannot set on nil pointer")
	}

	if !v.IsValid() {
		return item, fmt.Errorf("invalid final container")
	}

	var (
		key = args[len(args)-1]
		val = reflect.ValueOf(value)
	)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		idx, err := reflection.ToIndex(key)
		if err != nil {
			return item, fmt.Errorf("final index: %w", err)
		}

		if idx < 0 || idx >= int64(v.Len()) {
			return item, fmt.Errorf("final index out of range")
		}

		target := v.Index(int(idx))

		if !target.CanSet() {
			return item, fmt.Errorf("target slice element is not settable")
		}

		if val, err = reflection.ConvertValue(val, target.Type()); err != nil {
			return item, err
		}

		target.Set(val)
	case reflect.Map:
		if v.IsNil() {
			return item, fmt.Errorf("cannot set key in nil map")
		}

		kv, err := reflection.ResolveKey(v, key)
		if err != nil {
			return item, fmt.Errorf("final key: %w", err)
		}

		if val, err = reflection.ConvertValue(val, v.Type().Elem()); err != nil {
			return item, err
		}

		v.SetMapIndex(kv, val)
	case reflect.Struct:
		target, err := reflection.ResolveField(v, key)
		if err != nil {
			return item, fmt.Errorf("final key: %w", err)
		}

		if !target.CanSet() {
			return item, fmt.Errorf("target struct field is not settable (struct is unaddressable)")
		}

		if val, err = reflection.ConvertValue(val, target.Type()); err != nil {
			return item, err
		}

		target.Set(val)
	default:
		return item, fmt.Errorf("cannot set on type %v", v.Kind())
	}

	return item, nil
}

func Unset(item any, args ...any) (any, error) {
	if len(args) == 0 {
		return item, fmt.Errorf("unset requires at least one index/key")
	}

	v := reflect.ValueOf(item)

	for i := range len(args) - 1 {
		next, err := reflection.Lookup(v, args[i])
		if err != nil {
			if errors.Is(err, reflection.ErrKeyMissing) || errors.Is(err, reflection.ErrNilPointer) {
				return item, nil
			}

			return item, fmt.Errorf("at step %d: %w", i, err)
		}

		v = next
	}

	var err error

	if v, err = reflection.IndirectValue(v); err != nil {
		if errors.Is(err, reflection.ErrNilPointer) {
			return item, nil
		}

		return item, err
	}

	if !v.IsValid() {
		return item, fmt.Errorf("invalid final container")
	}

	key := args[len(args)-1]

	switch v.Kind() {
	case reflect.Map:
		if v.IsNil() {
			return item, nil
		}

		kv, err := reflection.ResolveKey(v, key)
		if err != nil {
			return item, fmt.Errorf("final key: %w", err)
		}

		v.SetMapIndex(kv, reflect.Value{})

		return item, nil
	case reflect.Struct:
		target, err := reflection.ResolveField(v, key)
		if err != nil {
			return item, fmt.Errorf("final key: %w", err)
		}

		if !target.CanSet() {
			return item, fmt.Errorf("target struct field is not settable (struct is unaddressable)")
		}

		target.Set(reflect.Zero(target.Type()))

		return item, nil
	case reflect.Slice, reflect.Array:
		return item, fmt.Errorf("cannot unset elements from slice or array, use Set with zero value instead")
	}

	return item, fmt.Errorf("cannot unset on type %v", v.Kind())
}

func IsSet(item any, args ...any) bool {
	v := reflect.ValueOf(item)

	if len(args) == 0 {
		return v.IsValid()
	}

	for _, arg := range args {
		next, err := reflection.Lookup(v, arg)
		if err != nil {
			return false
		}

		v = next
	}

	return true
}

func File(value any) (string, error) {
	abs, err := filepath.Abs(convert.ToString(value))
	if err != nil {
		return "", err
	}

	raw, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func Include(t *template.Template) func(name any, data ...any) (string, error) {
	return func(name any, data ...any) (string, error) {
		clone, err := t.Clone()
		if err != nil {
			return "", err
		}

		return render(clone, convert.ToString(name), data...)
	}
}

func Inline(t *template.Template) func(value any, data ...any) (string, error) {
	return func(value any, data ...any) (string, error) {
		clone, err := t.Clone()
		if err != nil {
			return "", err
		}

		name := fmt.Sprint("inline_", strings.ToLower(rand.Text()))

		if clone, err = clone.New(name).Parse(convert.ToString(value)); err != nil {
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

	if t = t.Lookup(name); t == nil {
		return "", fmt.Errorf("no template with name %#q found", name)
	}

	if err = t.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}
