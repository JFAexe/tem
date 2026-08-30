package functions

import (
	"fmt"
	"regexp/syntax"
	"strings"
	"unicode"

	"github.com/JFAexe/tem/pkg/cache"
	"github.com/JFAexe/tem/pkg/convert"
)

var (
	runeRangeCache = cache.NewSyncCache[runeRange, []rune]()
	runeRegexCache = cache.NewSyncCache[string, []rune]()
)

type runeRange struct{ lo, hi rune }

type Rune struct{}

func (f *Rune) RangeSet(lower, upper any) []rune {
	return cachedRangeSet(lower, upper)
}

func (f *Rune) RegexSet(pattern any) ([]rune, error) {
	return cachedRegexSet(strings.TrimSpace(convert.ToString(pattern)))
}

func cachedRangeSet(lower, upper any) []rune {
	var (
		lr  = convert.ToRune(lower)
		ur  = convert.ToRune(upper)
		lo  = min(lr, ur)
		hi  = max(lr, ur)
		key = runeRange{lo, hi}
	)

	set, _ := runeRangeCache.Get(key, func() ([]rune, error) {
		return expandRange(lo, hi), nil
	})

	return set
}

func cachedRegexSet(key string) ([]rune, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("failed to compute set: %w", ErrEmptyRegex)
	}

	return runeRegexCache.Get(key, func() ([]rune, error) {
		if s := fromUnicode(key); s != nil {
			return s, nil
		}

		set, err := syntaxSet(key)
		if err != nil {
			return nil, fmt.Errorf("failed to compute set: %w", err)
		}

		return set, nil
	})
}

func syntaxSet(pattern string) ([]rune, error) {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return nil, fmt.Errorf("parse regex %#q: %w", pattern, err)
	}

	re = re.Simplify()

	for re.Op == syntax.OpCapture && len(re.Sub) > 0 {
		re = re.Sub[0]
	}

	switch re.Op {
	case syntax.OpCharClass:
		var runes []rune

		for i := 0; i+1 < len(re.Rune); i += 2 {
			runes = append(runes, expandRange(re.Rune[i], re.Rune[i+1])...)
		}

		return runes, nil
	case syntax.OpLiteral:
		if len(re.Rune) != 1 {
			return nil, fmt.Errorf("multi-character literal %#q not supported", pattern)
		}

		return []rune{re.Rune[0]}, nil
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return cachedRangeSet(0, unicode.MaxRune), nil
	default:
		return nil, fmt.Errorf("pattern %#q is not a simple character set (op: %s)", pattern, re.Op)
	}
}

func fromUnicode(pattern string) []rune {
	switch {
	case strings.HasPrefix(pattern, `[\p{`) && strings.HasSuffix(pattern, `}]`):
		pattern = pattern[4 : len(pattern)-2]
	case strings.HasPrefix(pattern, `\p{`) && strings.HasSuffix(pattern, `}`):
		pattern = pattern[3 : len(pattern)-1]
	default:
		return nil
	}

	if pattern == "" {
		return nil
	}

	if rt, ok := unicode.Scripts[pattern]; ok {
		return expandTable(rt)
	}

	if rt, ok := unicode.Categories[pattern]; ok {
		return expandTable(rt)
	}

	if rt, ok := unicode.Properties[pattern]; ok {
		return expandTable(rt)
	}

	return nil
}

func expandRange(lo, hi rune) []rune {
	runes := make([]rune, 0, int(hi-lo)/4+16)

	for r := lo; r <= hi; r++ {
		if unicode.IsGraphic(r) {
			runes = append(runes, r)
		}
	}

	return runes
}

func expandTable(rt *unicode.RangeTable) []rune {
	if rt == nil {
		return nil
	}

	var runes []rune

	for _, r := range rt.R16 {
		for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
			if unicode.IsGraphic(c) {
				runes = append(runes, c)
			}
		}
	}

	for _, r := range rt.R32 {
		for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
			if unicode.IsGraphic(c) {
				runes = append(runes, c)
			}
		}
	}

	return runes
}
