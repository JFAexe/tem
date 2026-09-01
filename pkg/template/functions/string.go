package functions

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/JFAexe/tem/pkg/convert"
)

var (
	singleQuoteReplacer = strings.NewReplacer(`\"`, `"`, `'`, `\'`)
	stringCaser         = cases.Title(language.Und)
)

type String struct{}

func (*String) Quote(value any) string {
	return fmt.Sprintf("%q", convert.ToString(value))
}

func (*String) Squote(value any) string {
	return fmt.Sprint("'", singleQuoteReplacer.Replace(strings.Trim(fmt.Sprintf("%q", convert.ToString(value)), `"`)), "'")
}

func (*String) Bquote(value any) string {
	return fmt.Sprintf("`%s`", convert.ToString(value))
}

func (*String) EqualFold(other, value any) bool {
	return stringEqualFold(other, value)
}

func (*String) ToValidUTF8(replacement, value any) string {
	return strings.ToValidUTF8(convert.ToString(value), convert.ToString(replacement))
}

func (*String) Lower(value any) string {
	return strings.ToLower(convert.ToString(value))
}

func (*String) Upper(value any) string {
	return strings.ToUpper(convert.ToString(value))
}

func (*String) Title(value any) string {
	return stringCaser.String(convert.ToString(value))
}

func (*String) Cut(separator, value any) []string {
	before, after, _ := strings.Cut(convert.ToString(value), convert.ToString(separator))

	return []string{before, after}
}

func (*String) CutPrefix(prefix, value any) string {
	after, _ := strings.CutPrefix(convert.ToString(value), convert.ToString(prefix))

	return after
}

func (*String) CutSuffix(suffix, value any) string {
	before, _ := strings.CutSuffix(convert.ToString(value), convert.ToString(suffix))

	return before
}

func (*String) CutLast(separator, value any) []string {
	before, after, _ := strings.CutLast(convert.ToString(value), convert.ToString(separator))

	return []string{before, after}
}

func (*String) TrimSpace(value any) string {
	return strings.TrimSpace(convert.ToString(value))
}

func (*String) Trim(cutset, value any) string {
	return strings.Trim(convert.ToString(value), convert.ToString(cutset))
}

func (*String) TrimLeft(cutset, value any) string {
	return strings.TrimLeft(convert.ToString(value), convert.ToString(cutset))
}

func (*String) TrimRight(cutset, value any) string {
	return strings.TrimRight(convert.ToString(value), convert.ToString(cutset))
}

func (*String) TrimPrefix(prefix, value any) string {
	return strings.TrimPrefix(convert.ToString(value), convert.ToString(prefix))
}

func (*String) TrimSuffix(suffix, value any) string {
	return strings.TrimSuffix(convert.ToString(value), convert.ToString(suffix))
}

func (*String) HasPrefix(prefix, value any) bool {
	return stringHasPrefix(prefix, value)
}

func (*String) HasSuffix(suffix, value any) bool {
	return stringHasSuffix(suffix, value)
}

func (*String) Contains(subvalue, value any) bool {
	return stringContains(subvalue, value)
}

func (*String) ContainsAny(charset, value any) bool {
	return stringContainsAny(charset, value)
}

func (*String) Count(subvalue, value any) int {
	return strings.Count(convert.ToString(value), convert.ToString(subvalue))
}

func (*String) Replace(old, new, value any) string {
	return strings.ReplaceAll(convert.ToString(value), convert.ToString(old), convert.ToString(new))
}

func (*String) ReplaceN(old, new, count, value any) string {
	return strings.Replace(convert.ToString(value), convert.ToString(old), convert.ToString(new), convert.ToInt(count))
}

func (*String) Repeat(count, value any) string {
	return strings.Repeat(convert.ToString(value), max(0, convert.ToInt(count)))
}

func (*String) Split(separator, value any) []string {
	return strings.Split(convert.ToString(value), convert.ToString(separator))
}

func (*String) SplitN(separator, count, value any) []string {
	return strings.SplitN(convert.ToString(value), convert.ToString(separator), convert.ToInt(count))
}

func (*String) SplitAfter(separator, value any) []string {
	return strings.SplitAfter(convert.ToString(value), convert.ToString(separator))
}

func (*String) SplitAfterN(separator, count, value any) []string {
	return strings.SplitAfterN(convert.ToString(value), convert.ToString(separator), convert.ToInt(count))
}

func (*String) Join(separator any, values ...any) string {
	return strings.Join(convert.ToStringSlice(values), convert.ToString(separator))
}

func (*String) Fields(value any) []string {
	return strings.Fields(convert.ToString(value))
}

func (*String) Truncate(length, value any) string {
	var (
		str  = convert.ToString(value)
		size = convert.ToInt(length)
	)

	if str == "" {
		return str
	}

	var (
		total    = utf8.RuneCountInString(str)
		keep     = total + size
		negative = size < 0 && keep > 0
		positive = size >= 0 && total > size

		count, start int
		end          = len(str)
	)

	for i, w := 0, 0; i < len(str); i += w {
		_, w = utf8.DecodeRuneInString(str[i:])

		if negative && count == keep {
			start = i

			break
		}

		if positive && count == size {
			end = i

			break
		}

		count++
	}

	return str[start:end]
}

func (*String) IndentWith(char, level, value any) string {
	var (
		str = convert.ToString(value)
		lvl = convert.ToInt(level)
		chr = convert.ToRune(char)
	)

	if lvl <= 0 || str == "" {
		return str
	}

	var (
		builder strings.Builder

		newlines = strings.Count(str, "\n")
		prefix   = strings.Repeat(string(chr), lvl)
	)

	builder.Grow(len(str) + lvl*newlines + 1)

	for part := range strings.SplitSeq(str, "\n") {
		if !isSpace(part) {
			builder.WriteString(prefix)
			builder.WriteString(part)
		}

		builder.WriteByte('\n')
	}

	return builder.String()
}

func (f *String) Indent(level, value any) string {
	return f.IndentWith(' ', level, value)
}

func (f *String) IndentWithN(char, level, value any) string {
	str := f.IndentWith(char, level, value)

	if isSpace(str) || strings.HasPrefix(str, "\n") {
		return str
	}

	return "\n" + str
}

func (f *String) IndentN(level, value any) string {
	return f.IndentWithN(' ', level, value)
}

func (f *String) Fold(length, value any) string {
	return f.FoldBy(length, "", value)
}

func (*String) FoldBy(length, separator, value any) string {
	var (
		str = convert.ToString(value)
		lnt = convert.ToInt(length)
		sep = convert.ToString(separator)
	)

	if lnt <= 0 || str == "" {
		return str
	}

	total := utf8.RuneCountInString(str)

	if total <= lnt {
		return str
	}

	var (
		start, count int
		builder      strings.Builder

		size = len(str)
	)

	builder.Grow(len(str) + utf8.RuneCountInString(str)/lnt)

	for i := 0; i < size; {
		_, w := utf8.DecodeRuneInString(str[i:])

		i += w

		count++

		if count < lnt {
			continue
		}

		if j := strings.LastIndex(str[start:i], sep); j > 0 {
			builder.WriteString(str[start : start+j])

			start += j + len(sep)
			count = utf8.RuneCountInString(str[start:i])
		} else {
			builder.WriteString(str[start:i])

			start = i
			count = 0
		}

		if start < len(str) {
			builder.WriteByte('\n')
		}
	}

	if start < size {
		builder.WriteString(str[start:])
	}

	return builder.String()
}

func (*String) TrimTrailingSpace(value any) string {
	str := convert.ToString(value)

	if str == "" {
		return str
	}

	var builder strings.Builder

	builder.Grow(len(str))

	for part := range strings.SplitSeq(str, "\n") {
		if !isSpace(part) {
			builder.WriteString(strings.TrimRightFunc(part, unicode.IsSpace))
		}

		builder.WriteByte('\n')
	}

	return builder.String()
}

func stringEqualFold(other, value any) bool {
	return strings.EqualFold(convert.ToString(value), convert.ToString(other))
}

func stringHasPrefix(prefix, value any) bool {
	return strings.HasPrefix(convert.ToString(value), convert.ToString(prefix))
}

func stringHasSuffix(suffix, value any) bool {
	return strings.HasSuffix(convert.ToString(value), convert.ToString(suffix))
}

func stringContains(subvalue, value any) bool {
	return strings.Contains(convert.ToString(value), convert.ToString(subvalue))
}

func stringContainsAny(charset, value any) bool {
	return strings.ContainsAny(convert.ToString(value), convert.ToString(charset))
}

func isSpace(value string) bool {
	for _, r := range value {
		if !unicode.IsSpace(r) {
			return false
		}
	}

	return true
}
