package env

import (
	"os"
	"strings"
	"unicode"
)

type (
	Map        = map[string]string
	LookupFunc = func(value string) (string, bool)
	ExpandFunc = func(value string) string
)

func Escape(value string) string {
	return strings.ReplaceAll(value, "$", "$$")
}

func Unescape(value string) string {
	return strings.ReplaceAll(value, "$$", "$")
}

func Environ() Map {
	envs := make(Map)

	_ = NewDecoder(strings.NewReader(strings.Join(os.Environ(), "\n")), WithDecoderExpand(false)).Decode(&envs)

	return envs
}

func IsSet(key string) bool {
	_, ok := RawLookup(key)

	return ok
}

func RawLookup(key string) (string, bool) {
	if key == "$" {
		return "$", true
	}

	return os.LookupEnv(ToKey(key))
}

func Lookup(key string) (string, bool) {
	if value, ok := RawLookup(key); ok {
		return Expand(value), true
	}

	return "", false
}

func RawGet(key string) string {
	value, _ := RawLookup(key)

	return value
}

func Get(key string) string {
	value, _ := Lookup(key)

	return value
}

func RawOr(key, defaultValue string) string {
	if value, ok := RawLookup(key); ok && value != "" {
		return value
	}

	return defaultValue
}

func Or(key, defaultValue string) string {
	if value, ok := Lookup(key); ok && value != "" {
		return value
	}

	return defaultValue
}

func Expand(value string) string {
	return RawExpand(value, Lookup)
}

func ToKey(key string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToUpper(r)
		}

		return '_'
	}, key)
}
