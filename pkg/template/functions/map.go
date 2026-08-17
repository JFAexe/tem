package functions

import (
	"fmt"
	"maps"
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
	t := convert.ToAnyMap(to)

	for _, w := range with {
		maps.Copy(t, convert.ToAnyMap(w))
	}

	return t
}

func (*Map) Pick(m any, keys ...any) map[string]any {
	out := make(map[string]any, len(keys))

	if d, ok := m.(map[string]any); ok {
		for _, k := range keys {
			ks := convert.ToString(k)

			if v, ok := d[ks]; ok {
				out[ks] = v
			}
		}

		return out
	}

	rv := reflection.IndirectValue(reflect.ValueOf(m))
	if !rv.IsValid() || rv.Kind() != reflect.Map {
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

	if d, ok := m.(map[string]any); ok {
		out := make(map[string]any, len(d))

		for k, v := range d {
			if _, ok := set[k]; !ok {
				out[k] = v
			}
		}

		return out
	}

	rv := reflection.IndirectValue(reflect.ValueOf(m))
	if !rv.IsValid() || rv.Kind() != reflect.Map {
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
	if v, ok := m.(map[string]any); ok {
		out := make([]any, 0, len(v))

		for k := range v {
			out = append(out, k)
		}

		return out
	}

	rv := reflection.IndirectValue(reflect.ValueOf(m))
	if !rv.IsValid() || rv.Kind() != reflect.Map {
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
	if v, ok := m.(map[string]any); ok {
		out := make([]any, 0, len(v))

		for _, v := range v {
			out = append(out, v)
		}

		return out
	}

	rv := reflection.IndirectValue(reflect.ValueOf(m))
	if !rv.IsValid() || rv.Kind() != reflect.Map {
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
