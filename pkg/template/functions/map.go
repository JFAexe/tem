package functions

import (
	"fmt"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

type Map struct{}

func MapVarargInit(n *Map, args []any) (any, error) {
	return n.New(args...)
}

func (*Map) New(args ...any) (map[any]any, error) {
	if len(args) == 1 {
		var (
			nested = listArgSlice(args)
			flat   = make([]any, 0, len(nested)*2)
		)

		for _, item := range nested {
			if pair, ok := item.([]any); ok {
				if len(pair) >= 2 {
					flat = append(flat, pair[0], pair[1])
				}

				continue
			}

			flat = append(flat, item)
		}

		args = flat
	}

	if len(args)%2 != 0 {
		return nil, fmt.Errorf("amount of arguments for key-value pairs should be even, got %d: %v", len(args), args)
	}

	out := make(map[any]any, len(args)/2)

	for i := 0; i < len(args); i += 2 {
		out[args[i]] = args[i+1]
	}

	return out, nil
}

func (*Map) Merge(args ...any) map[any]any {
	if len(args) == 0 {
		return make(map[any]any)
	}

	var (
		to, with = popOne(args)
		t        = convert.ToAnyMap(to)
	)

	for _, w := range with {
		rv, ok := reflection.MapValue(w)
		if !ok {
			continue
		}

		for iter := rv.MapRange(); iter.Next(); {
			t[iter.Key().Interface()] = iter.Value().Interface()
		}
	}

	return t
}

func (*Map) Pick(args ...any) map[any]any {
	out := make(map[any]any)

	if len(args) == 0 {
		return out
	}

	m, keys := popOne(args)

	rv, ok := reflection.MapValue(m)
	if !ok {
		return out
	}

	for _, k := range listFlatten(keys) {
		kv, err := reflection.ResolveKey(rv, k)
		if err != nil {
			continue
		}

		if val := rv.MapIndex(kv); val.IsValid() {
			out[k] = val.Interface()
		}
	}

	return out
}

func (*Map) Omit(args ...any) map[any]any {
	out := make(map[any]any)

	if len(args) == 0 {
		return out
	}

	var (
		m, keys = popOne(args)
		set     = make(map[any]struct{}, len(keys))
	)

	for _, k := range listFlatten(keys) {
		set[k] = struct{}{}
	}

	rv, ok := reflection.MapValue(m)
	if !ok {
		return out
	}

	iter := rv.MapRange()

	for iter.Next() {
		k := iter.Key().Interface()

		if _, ok := set[k]; !ok {
			out[k] = iter.Value().Interface()
		}
	}

	return out
}

func (*Map) Keys(m any) []any {
	rv, ok := reflection.MapValue(m)
	if !ok {
		return make([]any, 0)
	}

	var (
		out  = make([]any, 0, rv.Len())
		iter = rv.MapRange()
	)

	for iter.Next() {
		out = append(out, iter.Key().Interface())
	}

	return out
}

func (*Map) Values(m any) []any {
	rv, ok := reflection.MapValue(m)
	if !ok {
		return make([]any, 0)
	}

	var (
		out  = make([]any, 0, rv.Len())
		iter = rv.MapRange()
	)

	for iter.Next() {
		out = append(out, iter.Value().Interface())
	}

	return out
}
