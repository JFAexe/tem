package functions

import (
	"cmp"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

type List struct{}

func ListVarargInit(n *List, args []any) (any, error) {
	return n.New(args...), nil
}

func (*List) New(values ...any) []any {
	return values
}

func (*List) First(items any) any {
	if v, ok := items.([]any); ok && len(v) > 0 {
		return v[0]
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(items))
	if err != nil || !rv.IsValid() {
		return nil
	}

	if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Len() > 0 {
		return rv.Index(0).Interface()
	}

	return nil
}

func (*List) Last(items any) any {
	if v, ok := items.([]any); ok && len(v) > 0 {
		return v[len(v)-1]
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(items))
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

func (*List) Compact(items any) []any {
	return slices.CompactFunc(convert.ToAnySlice(items), func(a, b any) bool {
		return reflection.CompareValues(reflect.ValueOf(a), reflect.ValueOf(b))
	})
}

func (*List) Reverse(items any) []any {
	out := convert.ToAnySlice(items)

	slices.Reverse(out)

	return out
}

func (*List) Sort(items any) ([]any, error) {
	out := convert.ToAnySlice(items)

	if len(out) <= 1 {
		return out, nil
	}

	slices.SortStableFunc(out, compareAny)

	return out, nil
}

func (*List) SortBy(key, items any) ([]any, error) {
	out := convert.ToAnySlice(items)

	if len(out) <= 1 {
		return out, nil
	}

	slices.SortStableFunc(out, func(a, b any) int {
		return compareAny(reflection.ExtractKey(a, key), reflection.ExtractKey(b, key))
	})

	return out, nil
}

func (*List) Unique(items any) []any {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s
	}

	out := make([]any, 0, len(s))

	for _, item := range s {
		var exists bool

		for _, u := range out {
			if reflection.CompareValues(reflect.ValueOf(item), reflect.ValueOf(u)) {
				exists = true

				break
			}
		}

		if !exists {
			out = append(out, item)
		}
	}

	return out
}

func (*List) UniqueBy(key, items any) []any {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s
	}

	out := make([]any, 0, len(s))

	for _, item := range s {
		var (
			exists bool

			ik = reflect.ValueOf(reflection.ExtractKey(item, key))
		)

		for _, u := range out {
			uk := reflect.ValueOf(reflection.ExtractKey(u, key))

			if reflection.CompareValues(ik, uk) {
				exists = true

				break
			}
		}

		if !exists {
			out = append(out, item)
		}
	}

	return out
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
		ra, _ = reflection.IndirectValue(reflect.ValueOf(a))
		rb, _ = reflection.IndirectValue(reflect.ValueOf(b))
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
