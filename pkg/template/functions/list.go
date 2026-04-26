package functions

import (
	"slices"
	"strings"
)

type List struct{}

func (List) New(values ...any) []any {
	return values
}

func (List) First(l any) any {
	if v := ToList(l); len(v) > 0 {
		return v[0]
	}

	return ""
}

func (List) Last(l any) any {
	if v := ToList(l); len(v) > 0 {
		return v[len(v)-1]
	}

	return ""
}

func (List) Concat(values ...any) []any {
	out := make([]any, 0)

	for i := range values {
		out = append(out, ToList(values[i])...)
	}

	return out
}

func (List) Join(sep string, value any) string {
	return strings.Join(ToStringList(value), sep)
}

func (List) Reverse(l any) []any {
	out := ToList(l)

	slices.Reverse(out)

	return out
}
