package functions

import (
	"cmp"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/JFAexe/tem/pkg/convert"
)

type List struct{}

func ListVarargInit(n *List, args []any) (any, error) {
	return n.New(args...), nil
}

func (*List) New(values ...any) []any {
	return values
}

func (*List) First(l any) any {
	if v, ok := l.([]any); ok && len(v) > 0 {
		return v[0]
	}

	rv, err := indirect(reflect.ValueOf(l))
	if err != nil || !rv.IsValid() {
		return nil
	}

	if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Len() > 0 {
		return rv.Index(0).Interface()
	}

	return nil
}

func (*List) Last(l any) any {
	if v, ok := l.([]any); ok && len(v) > 0 {
		return v[len(v)-1]
	}

	rv, err := indirect(reflect.ValueOf(l))
	if err != nil || !rv.IsValid() {
		return nil
	}

	if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Len() > 0 {
		return rv.Index(rv.Len() - 1).Interface()
	}

	return nil
}

func (*List) Concat(values ...any) []any {
	return listConcat(values...)
}

func (*List) Reverse(l any) []any {
	out := convert.ToAnySlice(l)

	slices.Reverse(out)

	return out
}

func (*List) Sort(items any) ([]any, error) {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s, nil
	}

	slices.SortStableFunc(s, compareAny)

	return s, nil
}

func (*List) SortBy(key, items any) ([]any, error) {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s, nil
	}

	slices.SortStableFunc(s, func(a, b any) int { return compareAny(extractKey(a, key), extractKey(b, key)) })

	return s, nil
}

func listConcat(values ...any) []any {
	out := make([][]any, 0, len(values))

	for _, v := range values {
		out = append(out, convert.ToAnySlice(v))
	}

	return slices.Concat(out...)
}

func compareAny(a, b any) int {
	if a == nil && b == nil {
		return 0
	}

	switch {
	case a == nil:
		return -1
	case b == nil:
		return 1
	}

	if ta, ok := a.(time.Time); ok {
		if tb, ok2 := b.(time.Time); ok2 {
			return ta.Compare(tb)
		}
	}

	var (
		ra, _ = indirect(reflect.ValueOf(a))
		rb, _ = indirect(reflect.ValueOf(b))
	)

	if (ra.CanInt() || ra.CanFloat()) && (rb.CanInt() || rb.CanFloat()) {
		return cmp.Compare(convert.ToFloat64(a), convert.ToFloat64(b))
	}

	if ra.Kind() == reflect.Bool && rb.Kind() == reflect.Bool {
		if ra.Bool() == rb.Bool() {
			return 0
		}

		if ra.Bool() {
			return 1
		}

		return -1
	}

	return cmp.Compare(strings.ToLower(convert.ToString(a)), strings.ToLower(convert.ToString(b)))
}
