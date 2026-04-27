package functions

import (
	"fmt"
	"regexp"
)

type Regex struct {
	cache map[string]*regexp.Regexp
}

func RegexNamespace() func() any {
	n := &Regex{
		cache: make(map[string]*regexp.Regexp),
	}

	return func() any {
		return n
	}
}

func (*Regex) Escape(str string) string {
	return regexp.QuoteMeta(str)
}

func (f *Regex) Match(regex string, str string) (bool, error) {
	exp, err := f.cached(regex)
	if err != nil {
		return false, err
	}

	return exp.MatchString(str), nil
}

func (f *Regex) Find(regex string, str string) (string, error) {
	exp, err := f.cached(regex)
	if err != nil {
		return "", err
	}

	return exp.FindString(str), nil
}

func (f *Regex) FindAll(regex string, n int, str string) ([]string, error) {
	exp, err := f.cached(regex)
	if err != nil {
		return make([]string, 0), err
	}

	return exp.FindAllString(str, n), nil
}

func (f *Regex) Replace(regex string, rpl string, str string) (string, error) {
	exp, err := f.cached(regex)
	if err != nil {
		return "", err
	}

	return exp.ReplaceAllString(str, rpl), nil
}

func (f *Regex) Split(regex string, n int, str string) ([]string, error) {
	exp, err := f.cached(regex)
	if err != nil {
		return make([]string, 0), err
	}

	return exp.Split(str, n), nil
}

func (f *Regex) cached(regex string) (*regexp.Regexp, error) {
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
