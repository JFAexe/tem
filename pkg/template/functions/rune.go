package functions

import (
	"fmt"
	"regexp/syntax"
	"strings"
	"unicode"

	"github.com/JFAexe/tem/pkg/convert"
)

type runeRange struct{ lo, hi rune }

type Rune struct {
	rangeCache map[runeRange][]rune
	regexCache map[string][]rune
}

func (f *Rune) RangeSet(lower, upper any) []rune {
	var (
		lr     = convert.ToRune(lower)
		ur     = convert.ToRune(upper)
		lo     = min(lr, ur)
		hi     = max(lr, ur)
		set, _ = f.cached(runeRange{lo, hi})
	)

	return set
}

func (f *Rune) RegexSet(pattern any) ([]rune, error) {
	return f.cached(strings.TrimSpace(convert.ToString(pattern)))
}

func (f *Rune) cached(key any) (set []rune, err error) {
	var ok bool

	switch k := key.(type) {
	case runeRange:
		if f.rangeCache == nil {
			f.rangeCache = make(map[runeRange][]rune)
		}

		if set, ok = f.rangeCache[k]; ok {
			return set, nil
		}

		set = expandRange(k.lo, k.hi)

		f.rangeCache[k] = set
	case string:
		if f.regexCache == nil {
			f.regexCache = make(map[string][]rune)
		}

		if set, ok = f.regexCache[k]; ok {
			return set, nil
		}

		if set = f.fromUnicode(k); set != nil {
			f.regexCache[k] = set

			return set, nil
		}

		if set, err = f.syntaxSet(k); err != nil {
			return nil, err
		}

		f.regexCache[k] = set
	}

	return set, nil
}

func (f *Rune) syntaxSet(pattern string) ([]rune, error) {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return nil, fmt.Errorf("parse regex %q: %w", pattern, err)
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
			return nil, fmt.Errorf("multi-character literal %q not supported", pattern)
		}

		return []rune{re.Rune[0]}, nil
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return f.RangeSet(0, unicode.MaxRune), nil
	default:
		return nil, fmt.Errorf("pattern %q is not a simple character set (op: %s)", pattern, re.Op)
	}
}

func (f *Rune) fromUnicode(pattern string) []rune {
	name := strings.TrimSpace(pattern)

	if strings.HasPrefix(name, `\p{`) && strings.HasSuffix(name, `}`) {
		name = name[3 : len(name)-1]
	} else if strings.HasPrefix(name, `[\p{`) && strings.HasSuffix(name, `}]`) {
		name = name[4 : len(name)-2]
	}

	if name == "" {
		return nil
	}

	if rt, ok := unicode.Scripts[name]; ok {
		return expandTable(rt)
	}

	if rt, ok := unicode.Categories[name]; ok {
		return expandTable(rt)
	}

	if rt, ok := unicode.Properties[name]; ok {
		return expandTable(rt)
	}

	return nil
}

func expandRange(lo, hi rune) []rune {
	if hi < lo {
		lo, hi = hi, lo
	}

	runes := make([]rune, 0, int(hi-lo)+1)

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
