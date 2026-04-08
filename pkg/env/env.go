package env

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	ErrNotKV           = errors.New("not an env key-value")
	ErrInvalidKVFormat = errors.New("invalid env key-value format")
)

type Map = map[string]string

func Escape(key string) string {
	return strings.ReplaceAll(key, "$", "$$")
}

func Unescape(key string) string {
	return strings.ReplaceAll(key, "$$", "$")
}

func Environ() Map {
	envs := make(Map)

	for _, env := range os.Environ() {
		key, value, err := ParseKV(env)
		if err != nil {
			continue
		}

		envs[key] = Expand(value)
	}

	return envs
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

func Get(key string) string {
	value, _ := Lookup(key)

	return value
}

func Or(key, defaultValue string) string {
	if value, ok := Lookup(key); ok && value != "" {
		return value
	}

	return defaultValue
}

func Expand(value string) string {
	if value == "" || !strings.Contains(value, "$") {
		return value
	}

	return os.Expand(value, Get)
}

func IsSet(key string) bool {
	_, ok := Lookup(key)

	return ok
}

func ToKey(key string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 'a' + 'A'
		}

		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}

		return '_'
	}, key)
}

func Unmarshal(data []byte) (Map, error) {
	return UnmarshalString(string(data))
}

func UnmarshalString(data string) (Map, error) {
	return Decode(strings.NewReader(data))
}

func Decode(r io.Reader) (Map, error) {
	var (
		scanner = bufio.NewScanner(r)
		envs    = make(Map)
	)

	for scanner.Scan() {
		key, value, err := ParseKV(strings.TrimSpace(scanner.Text()))
		if err != nil {
			continue
		}

		envs[key] = Expand(value)
	}

	return envs, scanner.Err()
}

func ParseKV(kv string) (string, string, error) {
	if kv = strings.TrimSpace(kv); kv == "" || strings.HasPrefix(kv, "#") {
		return "", "", ErrNotKV
	}

	parts := strings.SplitN(kv, "=", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("%w: %#q (expected KEY=value)", ErrInvalidKVFormat, kv)
	}

	var (
		key   = strings.TrimSpace(parts[0])
		value = strings.TrimSpace(parts[1])
	)

	if key == "" {
		return "", "", fmt.Errorf("%w: key is empty", ErrInvalidKVFormat)
	}

	if unquoted, err := strconv.Unquote(value); err == nil {
		value = unquoted
	}

	return ToKey(key), value, nil
}
