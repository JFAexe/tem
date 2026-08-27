package functions

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/JFAexe/tem/pkg/convert"
)

var regexCache sync.Map

type Regex struct{}

func (*Regex) Escape(str any) string {
	return regexp.QuoteMeta(convert.ToString(str))
}

func (*Regex) Match(regex, str any) (bool, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return false, err
	}

	return exp.MatchString(convert.ToString(str)), nil
}

func (*Regex) Find(regex, str any) (string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return "", err
	}

	return exp.FindString(convert.ToString(str)), nil
}

func (*Regex) FindAll(regex, n, str any) ([]string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return make([]string, 0), err
	}

	return exp.FindAllString(convert.ToString(str), convert.ToInt(n)), nil
}

func (*Regex) Replace(regex, rpl, str any) (string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return "", err
	}

	return exp.ReplaceAllString(convert.ToString(str), convert.ToString(rpl)), nil
}

func (*Regex) Split(regex, n, str any) ([]string, error) {
	exp, err := cachedRegex(convert.ToString(regex))
	if err != nil {
		return make([]string, 0), err
	}

	return exp.Split(convert.ToString(str), convert.ToInt(n)), nil
}

func cachedRegex(regex string) (*regexp.Regexp, error) {
	if exp, ok := regexCache.Load(regex); ok {
		return exp.(*regexp.Regexp), nil
	}

	exp, err := regexp.Compile(regex)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	regexCache.Store(regex, exp)

	return exp, nil
}
