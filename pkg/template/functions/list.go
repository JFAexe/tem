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
	rv, ok := sliceReflectValue(items)
	if !ok || rv.Len() == 0 {
		return nil
	}

	return rv.Index(0).Interface()
}

func (*List) Last(items any) any {
	rv, ok := sliceReflectValue(items)
	if !ok || rv.Len() == 0 {
		return nil
	}

	return rv.Index(rv.Len() - 1).Interface()
}

func (*List) Append(item any, values ...any) []any {
	return append(convert.ToAnySlice(item), values...)
}

func (*List) Prepend(item any, values ...any) []any {
	return append(values, convert.ToAnySlice(item)...)
}

func (*List) Concat(values ...any) []any {
	return listConcat(values...)
}

func (*List) Flatten(items any) []any {
	var (
		s   = convert.ToAnySlice(items)
		out = make([]any, 0, len(s))
	)

	listFlatten(s, &out)

	return out
}

func (*List) Compact(items any) []any {
	return slices.CompactFunc(convert.ToAnySlice(items), equalAny)
}

func (*List) Reverse(items any) []any {
	out := convert.ToAnySlice(items)

	slices.Reverse(out)

	return out
}

func (*List) Sort(items any) ([]any, error) {
	out := convert.ToAnySlice(items)

	if len(out) > 1 {
		slices.SortStableFunc(out, compareAny)
	}

	return out, nil
}

func (*List) SortBy(key, items any) ([]any, error) {
	out := convert.ToAnySlice(items)
	if len(out) <= 1 {
		return out, nil
	}

	type kv struct {
		key any
		val any
	}

	pairs := make([]kv, len(out))

	for i, item := range out {
		pairs[i] = kv{reflection.ExtractKey(item, key), item}
	}

	slices.SortStableFunc(pairs, func(a, b kv) int { return compareAny(a.key, b.key) })

	for i, p := range pairs {
		out[i] = p.val
	}

	return out, nil
}

func (*List) Unique(items any) []any {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s
	}

	return uniqueBy(s, convert.ToAny)
}

func (*List) UniqueBy(key, items any) []any {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s
	}

	return uniqueBy(s, func(v any) any { return reflection.ExtractKey(v, key) })
}

func listConcat(values ...any) []any {
	out := make([][]any, 0, len(values))

	for _, v := range values {
		out = append(out, convert.ToAnySlice(v))
	}

	return slices.Concat(out...)
}

func listFlatten(items []any, out *[]any) {
	for _, item := range items {
		if s, ok := item.([]any); ok {
			listFlatten(s, out)

			continue
		}

		if _, ok := sliceReflectValue(item); ok {
			listFlatten(convert.ToAnySlice(item), out)

			continue
		}

		*out = append(*out, item)
	}
}

func uniqueBy(s []any, key func(any) any) []any {
	var (
		out  = make([]any, 0, len(s))
		seen = make(map[any]struct{}, len(s))
	)

	for _, item := range s {
		var (
			k = key(item)
			t = reflect.TypeOf(k)
		)

		if k == nil || (t != nil && t.Comparable()) {
			if _, ok := seen[k]; ok {
				continue
			}

			seen[k] = struct{}{}
			out = append(out, item)

			continue
		}

		var skip bool

		for _, v := range out {
			if equalAny(k, key(v)) {
				skip = true

				break
			}
		}

		if !skip {
			out = append(out, item)
		}
	}

	return out
}

func sliceReflectValue(items any) (reflect.Value, bool) {
	rv := reflection.IndirectValue(reflect.ValueOf(items))
	if !rv.IsValid() {
		return rv, false
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		return rv, true
	}

	return rv, false
}

func equalAny(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}

	switch a := a.(type) {
	case int:
		if b, ok := b.(int); ok {
			return a == b
		}
	case string:
		if b, ok := b.(string); ok {
			return a == b
		}
	case bool:
		if b, ok := b.(bool); ok {
			return a == b
		}
	case time.Time:
		if b, ok := b.(time.Time); ok {
			return a.Equal(b)
		}
	}

	return reflection.CompareValues(reflect.ValueOf(a), reflect.ValueOf(b))
}

func compareAny(a, b any) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	}

	if ta, ok := a.(time.Time); ok {
		if tb, ok := b.(time.Time); ok {
			return ta.Compare(tb)
		}
	}

	var (
		ra = reflection.IndirectValue(reflect.ValueOf(a))
		rb = reflection.IndirectValue(reflect.ValueOf(b))
	)

	if (ra.CanInt() || ra.CanUint() || ra.CanFloat()) && (rb.CanInt() || rb.CanUint() || rb.CanFloat()) {
		return cmp.Compare(convert.ToFloat64(a), convert.ToFloat64(b))
	}

	if ra.Kind() == reflect.Bool && rb.Kind() == reflect.Bool {
		switch {
		case ra.Bool() == rb.Bool():
			return 0
		case ra.Bool():
			return 1
		default:
			return -1
		}
	}

	return cmp.Compare(strings.ToLower(convert.ToString(a)), strings.ToLower(convert.ToString(b)))
}
