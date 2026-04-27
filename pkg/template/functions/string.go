package functions

import (
	"fmt"
	"strings"
)

type String struct{}

func StringNamespace() func() any {
	n := new(String)

	return func() any {
		return n
	}
}

func (*String) Quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func (*String) Squote(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, `'`, `''`))
}

func (*String) Bquote(s string) string {
	return fmt.Sprintf("`%s`", strings.ReplaceAll(s, "`", "``"))
}

func (*String) EqualFold(t, s string) bool {
	return strings.EqualFold(s, t)
}

func (*String) Lower(s string) string {
	return strings.ToLower(s)
}

func (*String) Upper(s string) string {
	return strings.ToUpper(s)
}

func (*String) Title(s string) string {
	return strings.ToTitle(s)
}

func (*String) TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

func (*String) Trim(cutset, s string) string {
	return strings.Trim(s, cutset)
}

func (*String) TrimLeft(cutset, s string) string {
	return strings.TrimLeft(s, cutset)
}

func (*String) TrimRight(cutset, s string) string {
	return strings.TrimRight(s, cutset)
}

func (*String) TrimPrefix(prefix, s string) string {
	return strings.TrimPrefix(s, prefix)
}

func (*String) TrimSuffix(suffix, s string) string {
	return strings.TrimSuffix(s, suffix)
}

func (*String) HasPrefix(prefix, s string) bool {
	return strings.HasPrefix(s, prefix)
}

func (*String) HasSuffix(suffix, s string) bool {
	return strings.HasSuffix(s, suffix)
}

func (*String) Contains(sub, s string) bool {
	return strings.Contains(s, sub)
}

func (*String) Replace(old, new, src string) string {
	return strings.ReplaceAll(src, old, new)
}

func (*String) Repeat(count int, s string) string {
	return strings.Repeat(s, count)
}

func (*String) Split(sep string, s string) []string {
	return strings.Split(s, sep)
}

func (*String) Join(sep string, values ...any) string {
	return strings.Join(ToStringList(values), sep)
}

func (*String) Truncate(size int, str string) string {
	runes := []rune(str)

	if size < 0 && len(runes)+size > 0 {
		return string(runes[len(runes)+size:])
	}

	if size >= 0 && len(runes) > size {
		return string(runes[:size])
	}

	return str
}

func (*String) Indent(level int, str string) string {
	if level <= 0 || str == "" {
		return str
	}

	var (
		builder strings.Builder

		prefix = strings.Repeat(" ", level)
	)

	for i, s := range strings.Split(str, "\n") {
		if i > 0 {
			builder.WriteByte('\n')
		}

		if strings.TrimSpace(s) != "" {
			builder.WriteString(prefix)
			builder.WriteString(s)
		}
	}

	return builder.String()
}
