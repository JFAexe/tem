package env

import (
	"maps"
	"os"
	"strings"
)

type Store map[string]string

func (s Store) Environ() Map {
	envs := make(Map)

	for _, env := range os.Environ() {
		key, value, err := ParseKV(env)
		if err != nil {
			continue
		}

		envs[key] = s.Expand(value)
	}

	maps.Copy(envs, s)

	return envs
}

func (s Store) RawLookup(key string) (string, bool) {
	if key == "$" {
		return "$", true
	}

	key = ToKey(key)

	if value, ok := s[key]; ok {
		return value, true
	}

	if value, ok := os.LookupEnv(key); ok {
		return value, true
	}

	return "", false
}

func (s Store) Lookup(key string) (string, bool) {
	if value, ok := s.Lookup(key); ok {
		return s.Expand(value), true
	}

	return "", false
}

func (s Store) Get(key string) string {
	value, _ := s.Lookup(key)

	return value
}

func (s Store) Or(key, defaultValue string) string {
	if value, ok := s.Lookup(key); ok && value != "" {
		return value
	}

	return defaultValue
}

func (s Store) Expand(value string) string {
	if !strings.Contains(value, "$") {
		return value
	}

	return os.Expand(value, s.Get)
}

func (s Store) IsSet(key string) bool {
	_, ok := s.Lookup(key)

	return ok
}
