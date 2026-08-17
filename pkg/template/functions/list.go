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

	rv := reflection.IndirectValue(reflect.ValueOf(items))
	if !rv.IsValid() {
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

	rv := reflection.IndirectValue(reflect.ValueOf(items))
	if !rv.IsValid() {
		return nil
	}

	if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Len() > 0 {
		return rv.Index(rv.Len() - 1).Interface()
	}

	return nil
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

	if len(out) <= 1 {
		return out, nil
	}

	slices.SortStableFunc(out, compareAny)

	return out, nil
}

func (*List) SortBy(key, items any) ([]any, error) {
	unsorted := convert.ToAnySlice(items)

	if len(unsorted) <= 1 {
		return unsorted, nil
	}

	keys := make([]any, len(unsorted))

	for i, item := range unsorted {
		keys[i] = reflection.ExtractKey(item, key)
	}

	indices := make([]int, len(unsorted))

	for i := range indices {
		indices[i] = i
	}

	slices.SortStableFunc(indices, func(i, j int) int {
		return compareAny(keys[i], keys[j])
	})

	out := make([]any, len(unsorted))

	for i, idx := range indices {
		out[i] = unsorted[idx]
	}

	return out, nil
}

func (*List) Unique(items any) []any {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s
	}

	var (
		out          = make([]any, 0, len(s))
		seen         = make(map[any]struct{}, len(s))
		uncomparable = false
	)

	for _, item := range s {
		if !uncomparable && reflect.TypeOf(item) != nil && reflect.TypeOf(item).Comparable() {
			if _, ok := seen[item]; ok {
				continue
			}

			seen[item] = struct{}{}
			out = append(out, item)

			continue
		}

		uncomparable = true

		var exists bool

		for _, u := range out {
			if equalAny(item, u) {
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

	var (
		out          = make([]any, 0, len(s))
		seen         = make(map[any]struct{}, len(s))
		uncomparable = false
	)

	for _, item := range s {
		k := reflection.ExtractKey(item, key)

		if !uncomparable && reflect.TypeOf(k) != nil && reflect.TypeOf(k).Comparable() {
			if _, ok := seen[k]; ok {
				continue
			}

			seen[k] = struct{}{}
			out = append(out, item)

			continue
		}

		uncomparable = true

		var exists bool

		for _, u := range out {
			if equalAny(k, reflection.ExtractKey(u, key)) {
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

func listFlatten(items []any, out *[]any) {
	for _, item := range items {
		if s, ok := item.([]any); ok {
			listFlatten(s, out)

			continue
		}

		if rv := reflection.IndirectValue(reflect.ValueOf(item)); rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
			listFlatten(convert.ToAnySlice(item), out)

			continue
		}

		*out = append(*out, item)
	}
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

	switch a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		switch b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return cmp.Compare(convert.ToFloat64(a), convert.ToFloat64(b))
		}
	}

	if ba, ok := a.(bool); ok {
		if bb, ok := b.(bool); ok {
			if ba == bb {
				return 0
			}

			if ba {
				return 1
			}

			return -1
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
