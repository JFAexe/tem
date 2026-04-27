package functions

import (
	"fmt"
	"maps"
	"slices"
)

type Map struct{}

func MapNamespace() func(...any) (any, error) {
	n := new(Map)

	return func(args ...any) (any, error) {
		if len(args) > 0 {
			return n.New(args...)
		}

		return n, nil
	}
}

func (*Map) New(kv ...any) (map[string]any, error) {
	out := make(map[string]any)

	if len(kv)%2 != 0 {
		return out, fmt.Errorf("amount of arguments for key-values should be even, got %d", len(kv))
	}

	for i := 0; i < len(kv); i += 2 {
		out[ToString(kv[i])] = kv[i+1]
	}

	return out, nil
}

func (*Map) Get(key string, d map[string]any) any {
	if v, ok := d[key]; ok {
		return v
	}

	return ""
}

func (*Map) GetOr(key string, defaultValue any, d map[string]any) any {
	if v, ok := d[key]; ok {
		return v
	}

	return defaultValue
}

func (*Map) Set(key string, value any, d map[string]any) map[string]any {
	d[key] = value

	return d
}

func (*Map) Unset(key string, d map[string]any) map[string]any {
	delete(d, key)

	return d
}

func (*Map) IsSet(key string, d map[string]any) bool {
	_, ok := d[key]

	return ok
}

func (*Map) Merge(from, to map[string]any) map[string]any {
	maps.Copy(to, from)

	return to
}

func (*Map) Pick(d map[string]any, keys ...string) map[string]any {
	out := make(map[string]any, len(keys))

	for _, k := range keys {
		if v, ok := d[k]; ok {
			out[k] = v
		}
	}

	return out
}

func (*Map) Omit(d map[string]any, keys ...string) map[string]any {
	out := make(map[string]any, len(d))

	for k, v := range d {
		if !slices.Contains(keys, k) {
			out[k] = v
		}
	}

	return out
}

func (*Map) Keys(d map[string]any) []any {
	out := make([]any, 0, len(d))

	for k := range d {
		out = append(out, k)
	}

	return out
}

func (*Map) Values(d map[string]any) []any {
	return slices.Collect(maps.Values(d))
}
