package functions

import (
	"slices"
	"strings"

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
	if v := convert.ToAnyList(l); len(v) > 0 {
		return v[0]
	}

	return ""
}

func (*List) Last(l any) any {
	if v := convert.ToAnyList(l); len(v) > 0 {
		return v[len(v)-1]
	}

	return ""
}

func (*List) Concat(values ...any) []any {
	out := make([]any, 0)

	for i := range values {
		out = append(out, convert.ToAnyList(values[i])...)
	}

	return out
}

func (*List) Join(sep string, value any) string {
	return strings.Join(convert.ToStringList(value), sep)
}

func (*List) Reverse(l any) []any {
	out := convert.ToAnyList(l)

	slices.Reverse(out)

	return out
}
