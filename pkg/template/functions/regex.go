package functions

import (
	"fmt"
	"regexp"

	"github.com/JFAexe/tem/pkg/cache"
	"github.com/JFAexe/tem/pkg/convert"
)

var regexCache = cache.NewSyncCache[string, *regexp.Regexp]()

type Regex struct{}

func (*Regex) Escape(value any) string {
	return regexp.QuoteMeta(convert.ToString(value))
}

func (*Regex) Match(regex, value any) (bool, error) {
	return regexMatch(regex, value)
}

func (*Regex) Find(regex, value any) (string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return "", err
	}

	return exp.FindString(convert.ToString(value)), nil
}

func (*Regex) FindAll(regex, n, value any) ([]string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return make([]string, 0), err
	}

	return exp.FindAllString(convert.ToString(value), convert.ToInt(n)), nil
}

func (*Regex) Replace(regex, replacement, value any) (string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return "", err
	}

	return exp.ReplaceAllString(convert.ToString(value), convert.ToString(replacement)), nil
}

func (*Regex) ReplaceLiteral(regex, replacement, value any) (string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return "", err
	}

	return exp.ReplaceAllLiteralString(convert.ToString(value), convert.ToString(replacement)), nil
}

func (*Regex) Split(regex, n, value any) ([]string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return make([]string, 0), err
	}

	return exp.Split(convert.ToString(value), convert.ToInt(n)), nil
}

func regexMatch(regex, value any) (bool, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return false, err
	}

	return exp.MatchString(convert.ToString(value)), nil
}

func cachedRegex(regex string) (*regexp.Regexp, error) {
	return regexCache.Get(regex, func() (*regexp.Regexp, error) {
		exp, err := regexp.Compile(regex)
		if err != nil {
			return nil, fmt.Errorf("compile regex %#q: %w", regex, err)
		}

		return exp, nil
	})
}
