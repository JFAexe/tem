package functions

import (
	"fmt"
	"reflect"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

type Map struct{}

func MapVarargInit(n *Map, args []any) (any, error) {
	return n.New(args...)
}

func (*Map) New(kv ...any) (map[string]any, error) {
	out := make(map[string]any, len(kv)/2)

	if len(kv)%2 != 0 {
		return out, fmt.Errorf("amount of arguments for key-values should be even, got %d", len(kv))
	}

	for i := 0; i < len(kv); i += 2 {
		out[convert.ToString(kv[i])] = kv[i+1]
	}

	return out, nil
}

func (*Map) Merge(to any, with ...any) map[string]any {
	t := convert.ToStringAnyMap(to)

	for _, w := range with {
		rv, ok := mapReflectValue(w)
		if !ok {
			continue
		}

		for iter := rv.MapRange(); iter.Next(); {
			t[convert.ToString(iter.Key().Interface())] = iter.Value().Interface()
		}
	}

	return t
}

func (*Map) Pick(m any, keys ...any) map[string]any {
	out := make(map[string]any, len(keys))

	rv, ok := mapReflectValue(m)
	if !ok {
		return out
	}

	for _, k := range keys {
		kv, err := reflection.ResolveKey(rv, k)
		if err != nil {
			continue
		}

		if val := rv.MapIndex(kv); val.IsValid() {
			out[convert.ToString(k)] = val.Interface()
		}
	}

	return out
}

func (*Map) Omit(m any, keys ...any) map[string]any {
	set := make(map[string]struct{}, len(keys))

	for _, k := range keys {
		set[convert.ToString(k)] = struct{}{}
	}

	rv, ok := mapReflectValue(m)
	if !ok {
		return make(map[string]any)
	}

	var (
		out  = make(map[string]any, rv.Len())
		iter = rv.MapRange()
	)

	for iter.Next() {
		ks := convert.ToString(iter.Key().Interface())

		if _, ok := set[ks]; !ok {
			out[ks] = iter.Value().Interface()
		}
	}

	return out
}

func (*Map) Keys(m any) []any {
	rv, ok := mapReflectValue(m)
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
	rv, ok := mapReflectValue(m)
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

func mapReflectValue(m any) (reflect.Value, bool) {
	rv := reflection.IndirectValue(reflect.ValueOf(m))
	if !rv.IsValid() || rv.Kind() != reflect.Map {
		return rv, false
	}

	return rv, true
}
