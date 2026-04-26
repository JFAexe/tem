package functions

import "regexp"

type Regex struct{}

func (Regex) Escape(str string) string {
	return regexp.QuoteMeta(str)
}

func (Regex) Match(regex string, str string) (bool, error) {
	return regexp.MatchString(regex, str)
}

func (Regex) Find(regex string, str string) (string, error) {
	exp, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}

	return exp.FindString(str), nil
}

func (Regex) FindAll(regex string, n int, str string) ([]string, error) {
	exp, err := regexp.Compile(regex)
	if err != nil {
		return make([]string, 0), err
	}

	return exp.FindAllString(str, n), nil
}

func (Regex) Replace(regex string, rpl string, str string) (string, error) {
	exp, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}

	return exp.ReplaceAllString(str, rpl), nil
}

func (Regex) Split(regex string, n int, str string) ([]string, error) {
	exp, err := regexp.Compile(regex)
	if err != nil {
		return make([]string, 0), err
	}

	return exp.Split(str, n), nil
}
