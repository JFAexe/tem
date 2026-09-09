package functions

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"html"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"text/template"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

var (
	ErrAssertion           = errors.New("assertion failed")
	ErrValueRequired       = errors.New("value is required")
	ErrTooManyArguments    = errors.New("too many arguments")
	ErrEmptyRegex          = errors.New("empty regexp")
	ErrEmptyPath           = errors.New("can't walk empty path")
	ErrUpperNegativeOrZero = errors.New("upper boundary must be greater than 0")
	ErrLowerGreaterEqual   = errors.New("lower boundary must be less than upper boundary")
	ErrRangeTooLarge       = errors.New("range is too large")
)

const maxIncludeDepth = 16

type (
	UnaryErrorPredicate    = func(any) (bool, error)
	UnaryPredicate         = func(any) bool
	BinaryErrorPredicate   = func(any, any) (bool, error)
	BinaryPredicate        = func(any, any) bool
	TernaryErrorPredicate  = func(any, any, any) (bool, error)
	TernaryPredicate       = func(any, any, any) bool
	VariadicErrorPredicate = func(...any) (bool, error)
	VariadicPredicate      = func(...any) bool
)

func Namespace(n any) func() any {
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
	return template.FuncMap{
		"hostname":   os.Hostname,
		"assert":     Assert,
		"ternary":    Ternary,
		"default":    Default,
		"coalesce":   Coalesce,
		"zero":       Zero,
		"isZero":     IsZero,
		"isEmpty":    IsEmpty,
		"index":      Index,
		"indexOr":    IndexOr,
		"indexOrSet": IndexOrSet,
		"set":        Set,
		"unset":      Unset,
		"isSet":      IsSet,
		"in":         In,
		"inline":     Inline(t),
		"include":    Include(t),
		"file":       File,
		"html":       HTML,
		"unhtml":     UnHTML,
		"to":         Namespace(new(Convert)),
		"data":       Namespace(new(Data)),
		"env":        NamespaceVararg(new(Env), EnvVarargInit),
		"filepath":   Namespace(new(Filepath)),
		"ip":         Namespace(new(IP)),
		"list":       NamespaceVararg(new(List), ListVarargInit),
		"map":        NamespaceVararg(new(Map), MapVarargInit),
		"math":       Namespace(new(Math)),
		"net":        Namespace(new(Net)),
		"path":       Namespace(new(Path)),
		"is":         Namespace(new(Predicate)),
		"random":     Namespace(new(Random)),
		"regex":      Namespace(new(Regex)),
		"rune":       Namespace(new(Rune)),
		"string":     Namespace(new(String)),
		"time":       Namespace(new(Time)),
		"type":       Namespace(new(Type)),
		"uuid":       Namespace(new(UUID)),
	}
}

func Assert(args ...any) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("%w: missing condition", ErrAssertion)
	}

	condition, rest := popOne(args)

	if convert.ToBool(condition) {
		return "", nil
	}

	if len(rest) > 0 {
		if msg := convert.ToString(rest[0]); msg != "" {
			return "", fmt.Errorf("%w: %s", ErrAssertion, msg)
		}
	}

	return "", fmt.Errorf("%w (condition: %v)", ErrAssertion, condition)
}

func Ternary(truthy, falsy, condition any) any {
	if convert.ToBool(condition) {
		return truthy
	}

	return falsy
}

func Default(def, value any) any {
	if reflection.IsEmpty(value) {
		return def
	}

	return value
}

func Coalesce(args ...any) any {
	for _, v := range args {
		if !reflection.IsEmpty(v) {
			return v
		}
	}

	return nil
}

func Zero(value any) any {
	return reflection.Zero(value)
}

func IsZero(value any) bool {
	return reflection.IsZero(value)
}

func IsEmpty(value any) bool {
	return reflection.IsEmpty(value)
}

func Index(args ...any) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("index: missing item")
	}

	item, path := popOne(args)

	return index(item, path)
}

func index(item any, args []any) (any, error) {
	path := listFlatten(args)
	if len(path) == 0 {
		return nil, fmt.Errorf("index requires at least one key/index")
	}

	v := reflect.ValueOf(item)

	for i, k := range path {
		next, err := reflection.Lookup(v, k)
		if err != nil {
			return nil, fmt.Errorf("key %#q (%d): %w", convert.ToString(k), i, err)
		}

		v = next
	}

	return v.Interface(), nil
}

func IndexOr(args ...any) any {
	if len(args) < 2 {
		return nil
	}

	item, value, keys := popTwo(args)

	return indexOr(item, value, keys)
}

func indexOr(item, value any, args []any) any {
	if len(args) == 0 {
		return Default(item, value)
	}

	if v, err := index(item, args); err == nil {
		return v
	}

	return value
}

func IndexOrSet(args ...any) (any, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("indexOrSet: expected ...path, value, item")
	}

	item, value, keys := popTwo(args)

	if v, err := index(item, keys); err == nil {
		return v, nil
	}

	if _, err := set(item, value, keys); err != nil {
		return nil, err
	}

	return index(item, keys)
}

func Set(args ...any) (any, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("set: expected ...path, value, item")
	}

	item, value, keys := popTwo(args)

	return set(item, value, keys)
}

func set(item, value any, args []any) (any, error) {
	path := listFlatten(args)
	if len(path) == 0 {
		return item, fmt.Errorf("set requires at least one index/key")
	}

	var (
		v         = reflect.ValueOf(item)
		key, rest = popOne(path)
	)

	for i, k := range rest {
		next, err := reflection.Lookup(v, k)
		if err != nil {
			return item, fmt.Errorf("key %#q (%d): %w", convert.ToString(k), i, err)
		}

		v = next
	}

	if err := reflection.Set(v, key, value); err != nil {
		return item, fmt.Errorf("key %#q: %w", convert.ToString(key), err)
	}

	return item, nil
}

func Unset(args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("unset: expected ...path, item")
	}

	item, path := popOne(args)

	return unset(item, path)
}

func unset(item any, args []any) (any, error) {
	path := listFlatten(args)
	if len(path) == 0 {
		return item, fmt.Errorf("unset requires at least one index/key")
	}

	var (
		v         = reflect.ValueOf(item)
		key, rest = popOne(path)
	)

	for i, k := range rest {
		next, err := reflection.Lookup(v, k)
		if err != nil {
			if errors.Is(err, reflection.ErrKeyMissing) || errors.Is(err, reflection.ErrNilPointer) {
				return item, nil
			}

			return item, fmt.Errorf("key %#q (%d): %w", convert.ToString(k), i, err)
		}

		v = next
	}

	if err := reflection.Unset(v, key); err != nil {
		return item, fmt.Errorf("key %#q: %w", convert.ToString(key), err)
	}

	return item, nil
}

func IsSet(args ...any) bool {
	if len(args) == 0 {
		return false
	}

	item, path := popOne(args)

	return isSet(item, path)
}

func isSet(item any, args []any) bool {
	if len(args) == 0 {
		return reflect.ValueOf(item).IsValid()
	}

	_, err := index(item, args)

	return err == nil
}

func In(args ...any) bool {
	if len(args) < 2 {
		return false
	}

	item, values := popOne(args)

	for _, v := range values {
		if in(item, v) {
			return true
		}
	}

	return false
}

func in(item, value any) bool {
	if value == nil || item == nil {
		return false
	}

	switch reflection.IndirectValue(item).Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct:
		for v, err := range reflection.Values(item) {
			if err != nil {
				return false
			}

			if equalAny(v, value) {
				return true
			}
		}
	}

	return equalAny(item, value)
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

func Include(t *template.Template) func(args ...any) (string, error) {
	var depth atomic.Int32

	return func(args ...any) (string, error) {
		if depth.Load() >= maxIncludeDepth {
			return "", fmt.Errorf("include depth limit exceeded (%d)", maxIncludeDepth)
		}

		var (
			name string
			ctx  any
		)

		switch len(args) {
		case 0:
			return "", fmt.Errorf("%w: expected name, or context, name", ErrValueRequired)
		case 1:
			name = convert.ToString(args[0])
			ctx = make(map[string]any)
		case 2:
			name = convert.ToString(args[1])
			ctx = args[0]
		default:
			return "", fmt.Errorf("%w: expected name, or context, name", ErrTooManyArguments)
		}

		depth.Add(1)
		defer depth.Add(-1)

		return render(t, name, ctx)
	}
}

func Inline(t *template.Template) func(args ...any) (string, error) {
	return func(args ...any) (string, error) {
		var (
			tpl, ctx any
			opts     []any
		)

		switch len(args) {
		case 0:
			return "", fmt.Errorf("%w: expected template, or context, template, or ...options, context, template", ErrValueRequired)
		case 1:
			tpl = args[0]
			ctx = make(map[string]any)
		case 2:
			tpl = args[1]
			ctx = args[0]
		default:
			tpl, ctx, opts = popTwo(args)
		}

		var left, right string

		for _, o := range opts {
			opt, arg, ok := strings.Cut(convert.ToString(o), "=")
			if !ok {
				return "", fmt.Errorf("inline: invalid option %#q (expected `key=value`)", o)
			}

			switch normalizeString(opt) {
			case "delims":
				delims := strings.Split(strings.TrimSpace(arg), " ")

				if len(delims) != 2 {
					return "", fmt.Errorf("inline: option delims expects `<left> <right>`, got %#q", arg)
				}

				left, right = strings.TrimSpace(delims[0]), strings.TrimSpace(delims[1])
			default:
				return "", fmt.Errorf("inline: unknown option %#q", opt)
			}
		}

		clone, err := t.Clone()
		if err != nil {
			return "", fmt.Errorf("inline: %w", err)
		}

		clone.Delims(left, right)

		clone.Funcs(template.FuncMap{
			"inline": func(...any) (string, error) {
				return "", errors.New("inline templates are restricted inside of inline templates")
			},
		})

		name := fmt.Sprint("inline_", strings.ToLower(rand.Text()))

		if clone, err = clone.New(name).Parse(convert.ToString(tpl)); err != nil {
			return "", fmt.Errorf("inline: %w", err)
		}

		return render(clone, name, ctx)
	}
}

func HTML(value any) string {
	return html.EscapeString(convert.ToString(value))
}

func UnHTML(value any) string {
	return html.UnescapeString(convert.ToString(value))
}

func render(t *template.Template, name string, ctx any) (string, error) {
	var buf bytes.Buffer

	if t = t.Lookup(name); t == nil {
		return "", fmt.Errorf("no template with name %#q found", name)
	}

	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func normalizeString(value any) string {
	return strings.ToLower(strings.TrimSpace(convert.ToString(value)))
}

func joinKeys(value any) string {
	return strings.Join(slices.Sorted(maps.Keys(convert.ToStringAnyMap(value))), ", ")
}

func popOne(args []any) (first any, rest []any) {
	return args[len(args)-1], args[:len(args)-1]
}

func popTwo(args []any) (first, second any, rest []any) {
	return args[len(args)-1], args[len(args)-2], args[:len(args)-2]
}

func popPredicate(args []any) ([]any, UnaryErrorPredicate, error) {
	if len(args) == 0 {
		return nil, nil, fmt.Errorf("%w: expected predicate, collection", ErrValueRequired)
	}

	var (
		split = len(args) - 1
		arity = -1
	)

	switch args[0].(type) {
	case UnaryErrorPredicate, UnaryPredicate:
		arity = 0
	case BinaryErrorPredicate, BinaryPredicate:
		arity = 1
	case TernaryErrorPredicate, TernaryPredicate:
		arity = 2
	}

	if arity >= 0 && len(args) >= arity+1 {
		split = arity + 1
	}

	fn, err := bindPredicate(args[:split])
	if err != nil {
		return nil, nil, err
	}

	return listArgSlice(args[split:]), fn, nil
}

func bindPredicate(args []any) (UnaryErrorPredicate, error) {
	if len(args) == 0 {
		return func(v any) (bool, error) { return convert.ToBool(v), nil }, nil
	}

	var (
		pred  = args[0]
		fixed = args[1:]
	)

	switch p := pred.(type) {
	case VariadicErrorPredicate:
		return func(v any) (bool, error) { return p(append(slices.Clone(fixed), v)...) }, nil
	case VariadicPredicate:
		return func(v any) (bool, error) { return p(append(slices.Clone(fixed), v)...), nil }, nil
	}

	switch len(fixed) {
	case 0:
		if p, ok := unaryPredicate(pred); ok {
			return p, nil
		}

		return func(v any) (bool, error) { return equalAny(v, pred), nil }, nil
	case 1:
		switch p := pred.(type) {
		case BinaryErrorPredicate:
			return func(v any) (bool, error) { return p(fixed[0], v) }, nil
		case BinaryPredicate:
			return func(v any) (bool, error) { return p(fixed[0], v), nil }, nil
		}
	case 2:
		switch p := pred.(type) {
		case TernaryErrorPredicate:
			return func(v any) (bool, error) { return p(fixed[0], fixed[1], v) }, nil
		case TernaryPredicate:
			return func(v any) (bool, error) { return p(fixed[0], fixed[1], v), nil }, nil
		}
	}

	return nil, fmt.Errorf("unsupported predicate type %T", pred)
}

func unaryPredicate(pred any) (UnaryErrorPredicate, bool) {
	switch p := pred.(type) {
	case UnaryErrorPredicate:
		return p, true
	case UnaryPredicate:
		return func(v any) (bool, error) { return p(v), nil }, true
	}

	return nil, false
}

func unaryPredicates(args []any) ([]UnaryErrorPredicate, error) {
	preds := make([]UnaryErrorPredicate, 0, len(args))

	for _, pred := range args {
		p, ok := unaryPredicate(pred)
		if !ok {
			return nil, fmt.Errorf("unsupported predicate type %T", pred)
		}

		preds = append(preds, p)
	}

	return preds, nil
}
