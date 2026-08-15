package functions

import (
	"fmt"
	"regexp"

	"github.com/JFAexe/tem/pkg/convert"
)

type Regex struct {
	cache map[string]*regexp.Regexp
}

func (*Regex) Escape(str any) string {
	return regexp.QuoteMeta(convert.ToString(str))
}

func (f *Regex) Match(regex, str any) (bool, error) {
	exp, err := f.cached(convert.ToString(regex))
	if err != nil {
		return false, err
	}

	return exp.MatchString(convert.ToString(str)), nil
}

func (f *Regex) Find(regex, str any) (string, error) {
	exp, err := f.cached(convert.ToString(regex))
	if err != nil {
		return "", err
	}

	return exp.FindString(convert.ToString(str)), nil
}

func (f *Regex) FindAll(regex, n, str any) ([]string, error) {
	exp, err := f.cached(convert.ToString(regex))
	if err != nil {
		return make([]string, 0), err
	}

	return exp.FindAllString(convert.ToString(str), convert.ToInt(n)), nil
}

func (f *Regex) Replace(regex, rpl, str any) (string, error) {
	exp, err := f.cached(convert.ToString(regex))
	if err != nil {
		return "", err
	}

	return exp.ReplaceAllString(convert.ToString(str), convert.ToString(rpl)), nil
}

func (f *Regex) Split(regex, n, str any) ([]string, error) {
	exp, err := f.cached(convert.ToString(regex))
	if err != nil {
		return make([]string, 0), err
	}

	return exp.Split(convert.ToString(str), convert.ToInt(n)), nil
}

func (f *Regex) cached(regex string) (*regexp.Regexp, error) {
	if f.cache == nil {
		f.cache = make(map[string]*regexp.Regexp)
	}

	if exp, ok := f.cache[regex]; ok {
		return exp, nil
	}

	exp, err := regexp.Compile(regex)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	f.cache[regex] = exp

	return exp, nil
}
