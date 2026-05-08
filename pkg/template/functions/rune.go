package functions

import (
	"fmt"
	"regexp"
	"regexp/syntax"
	"strings"
	"unicode"

	"github.com/JFAexe/tem/pkg/convert"
)

var regexUnicodeSet = regexp.MustCompile(`^\[?\\p\{([^}]+)}\]?$`)

type Rune struct {
	cacheRange map[string][]rune
	cacheRegex map[string][]rune
}

func NewRuneFuncs() *Rune {
	return &Rune{
		cacheRange: make(map[string][]rune),
		cacheRegex: make(map[string][]rune),
	}
}

func (f *Rune) RangeSet(lower, upper any) []rune {
	return f.rangeSet(convert.ToRune(lower), convert.ToRune(upper))
}

func (f *Rune) RegexSet(regex string) ([]rune, error) {
	if set, ok := f.cacheRegex[regex]; ok {
		return set, nil
	}

	if set := f.fromUnicode(regex); set != nil {
		f.cacheRegex[regex] = set

		return set, nil
	}

	set, err := f.syntaxSet(regex)
	if err != nil {
		return nil, err
	}

	f.cacheRegex[regex] = set

	return set, nil
}

func (f *Rune) rangeSet(lower, upper rune) []rune {
	var (
		lo = min(lower, upper)
		hi = max(lower, upper)
		id = fmt.Sprintf("%d:%d", lo, hi)
	)

	if set, ok := f.cacheRange[id]; ok {
		return set
	}

	runes := make([]rune, 0, int(hi-lo)+1)

	for r := lo; r <= hi; r++ {
		if unicode.IsGraphic(r) {
			runes = append(runes, r)
		}
	}

	f.cacheRange[id] = runes

	return runes
}

func (f *Rune) syntaxSet(regex string) ([]rune, error) {
	re, err := syntax.Parse(regex, syntax.Perl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse regex %#q: %w", regex, err)
	}

	re = re.Simplify()

	for re.Op == syntax.OpCapture && len(re.Sub) > 0 {
		re = re.Sub[0]
	}

	switch re.Op {
	case syntax.OpCharClass:
		var capacity int

		for i := 0; i < len(re.Rune); i += 2 {
			capacity += int(re.Rune[i+1]-re.Rune[i]) + 1
		}

		runes := make([]rune, 0, capacity)

		for i := 0; i < len(re.Rune); i += 2 {
			for r := rune(re.Rune[i]); r <= rune(re.Rune[i+1]); r++ {
				if unicode.IsGraphic(r) {
					runes = append(runes, r)
				}
			}
		}

		return runes, nil
	case syntax.OpLiteral:
		if len(re.Rune) != 1 {
			return nil, fmt.Errorf("multi-character literal not supported")
		}

		return []rune{re.Rune[0]}, nil
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return f.rangeSet(0, unicode.MaxRune), nil
	default:
		return nil, fmt.Errorf("pattern %q is not a simple character set (op: %s)", regex, re.Op)
	}
}

func (f *Rune) fromUnicode(pattern string) []rune {
	pattern = strings.TrimSpace(pattern)

	if matches := regexUnicodeSet.FindStringSubmatch(pattern); matches != nil {
		pattern = matches[1]
	}

	if pattern == "" {
		return nil
	}

	if rt, ok := unicode.Scripts[pattern]; ok {
		return f.fromRangeTable(rt)
	}

	if rt, ok := unicode.Categories[pattern]; ok {
		return f.fromRangeTable(rt)
	}

	if rt, ok := unicode.Properties[pattern]; ok {
		return f.fromRangeTable(rt)
	}

	return nil
}

func (*Rune) fromRangeTable(rt *unicode.RangeTable) (runes []rune) {
	if rt == nil {
		return nil
	}

	for _, r16 := range rt.R16 {
		for r := rune(r16.Lo); r <= rune(r16.Hi); r += rune(r16.Stride) {
			if unicode.IsGraphic(r) {
				runes = append(runes, r)
			}
		}
	}

	for _, r32 := range rt.R32 {
		for r := rune(r32.Lo); r <= rune(r32.Hi); r += rune(r32.Stride) {
			if unicode.IsGraphic(r) {
				runes = append(runes, r)
			}
		}
	}

	return runes
}
