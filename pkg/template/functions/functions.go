package functions

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/JFAexe/tem/pkg/convert"
)

var (
	ErrNilPointer   = errors.New("nil pointer")
	ErrInvalidIndex = errors.New("invalid index")
	ErrOutOfRange   = errors.New("index out of range")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrKeyMissing   = errors.New("map key missing")
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
		regexFuncs    = NewRegex()
		runeFuncs     = NewRune()
		randomFuncs   = NewRandom(runeFuncs)
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
		"include":  Include(t),
		"set":      Set,
		"unset":    Unset,
		"indexOr":  IndexOr,
		"default":  Default,
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

	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && !v.IsNil() {
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
		next, err := traverse(v, arg)
		if err != nil {
			return def
		}

		v = next
	}

	if v.IsValid() {
		return v.Interface()
	}

	return def
}

func Set(item, value any, args ...any) (any, error) {
	if len(args) == 0 {
		return item, fmt.Errorf("set requires at least one index/key")
	}

	v := reflect.ValueOf(item)

	for i := 0; i < len(args)-1; i++ {
		next, err := traverse(v, args[i])
		if err != nil {
			if errors.Is(err, ErrKeyMissing) {
				return item, fmt.Errorf("map key not found at step %d (cannot set nested value in nil map)", i)
			}

			return item, fmt.Errorf("at step %d: %w", i, err)
		}

		v = next
	}

	var err error

	if v, err = indirect(v); err != nil {
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
		idx, err := toIndex(key)
		if err != nil {
			return item, fmt.Errorf("final index: %w", err)
		}

		if idx < 0 || int(idx) >= v.Len() {
			return item, fmt.Errorf("final index out of range")
		}

		target := v.Index(int(idx))

		if !target.CanSet() {
			return item, fmt.Errorf("target slice element is not settable")
		}

		if !val.IsValid() {
			val = reflect.Zero(target.Type())
		} else if val.Type() != target.Type() {
			if !val.Type().ConvertibleTo(target.Type()) {
				return item, fmt.Errorf("value type %v not convertible to %v", val.Type(), target.Type())
			}
			val = val.Convert(target.Type())
		}

		target.Set(val)
	case reflect.Map:
		if v.IsNil() {
			return item, fmt.Errorf("cannot set key in nil map")
		}

		kv, err := resolveKey(v, key)
		if err != nil {
			return item, fmt.Errorf("final key: %w", err)
		}

		if !val.IsValid() {
			val = reflect.Zero(v.Type().Elem())
		} else if val.Type() != v.Type().Elem() {
			if !val.Type().ConvertibleTo(v.Type().Elem()) {
				return item, fmt.Errorf("value type %v not convertible to map elem type %v", val.Type(), v.Type().Elem())
			}

			val = val.Convert(v.Type().Elem())
		}

		v.SetMapIndex(kv, val)
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

	for i := 0; i < len(args)-1; i++ {
		next, err := traverse(v, args[i])
		if err != nil {
			if errors.Is(err, ErrKeyMissing) || errors.Is(err, ErrNilPointer) {
				return item, nil
			}

			return item, fmt.Errorf("at step %d: %w", i, err)
		}

		v = next
	}

	var err error

	if v, err = indirect(v); err != nil {
		if errors.Is(err, ErrNilPointer) {
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

		kv, err := resolveKey(v, key)
		if err != nil {
			return item, fmt.Errorf("final key: %w", err)
		}

		v.SetMapIndex(kv, reflect.Value{})

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
		next, err := traverse(v, arg)
		if err != nil {
			return false
		}
		v = next
	}

	return true
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

func indirect(v reflect.Value) (reflect.Value, error) {
	for {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return reflect.Value{}, ErrNilPointer
			}

			v = v.Elem()
		default:
			return v, nil
		}
	}
}

func toIndex(i any) (int64, error) {
	iv := reflect.ValueOf(i)

	if !iv.IsValid() {
		return 0, ErrInvalidIndex
	}

	switch iv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return iv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(iv.Uint()), nil
	}

	return 0, ErrInvalidIndex
}

func resolveKey(m reflect.Value, key any) (reflect.Value, error) {
	kv := reflect.ValueOf(key)

	if !kv.IsValid() {
		kv = reflect.Zero(m.Type().Key())
	}

	if !kv.Type().ConvertibleTo(m.Type().Key()) {
		return reflect.Value{}, ErrTypeMismatch
	}

	return kv.Convert(m.Type().Key()), nil
}

func extractKey(item any, key any) any {
	rv, err := indirect(reflect.ValueOf(item))
	if err != nil || !rv.IsValid() {
		return nil
	}

	switch rv.Kind() {
	case reflect.Map:
		kv, err := resolveKey(rv, key)
		if err != nil {
			return nil
		}

		if v := rv.MapIndex(kv); v.IsValid() {
			return v.Interface()
		}
	case reflect.Struct:
		if ks, ok := key.(string); ok {
			if field := rv.FieldByName(ks); field.IsValid() {
				return field.Interface()
			}
		}
	}

	return nil
}

func traverse(v reflect.Value, key any) (reflect.Value, error) {
	v, err := indirect(v)
	if err != nil {
		return reflect.Value{}, err
	}

	if !v.IsValid() {
		return reflect.Value{}, ErrInvalidIndex
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		idx, err := toIndex(key)
		if err != nil {
			return reflect.Value{}, err
		}

		if idx < 0 || int(idx) >= v.Len() {
			return reflect.Value{}, ErrOutOfRange
		}

		return v.Index(int(idx)), nil
	case reflect.Map:
		kv, err := resolveKey(v, key)
		if err != nil {
			return reflect.Value{}, err
		}

		val := v.MapIndex(kv)

		if !val.IsValid() {
			return reflect.Value{}, ErrKeyMissing
		}

		return val, nil
	}

	return reflect.Value{}, fmt.Errorf("cannot index into type %v", v.Kind())
}
