package env

import (
	"maps"
	"os"
)

type Store map[string]string

func (s Store) Environ() Map {
	envs := Environ()

	maps.Copy(envs, s)

	return envs
}

func (s Store) Set(key, value string) {
	s[ToKey(key)] = value
}

func (s Store) IsSet(key string) bool {
	_, ok := s.RawLookup(key)

	return ok
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
	if value, ok := s.RawLookup(key); ok {
		return s.Expand(value), true
	}

	return "", false
}

func (s Store) RawGet(key string) string {
	value, _ := s.RawLookup(key)

	return value
}

func (s Store) Get(key string) string {
	value, _ := s.Lookup(key)

	return value
}

func (s Store) RawOr(key, defaultValue string) string {
	if value, ok := s.RawLookup(key); ok && value != "" {
		return value
	}

	return defaultValue
}

func (s Store) Or(key, defaultValue string) string {
	if value, ok := s.Lookup(key); ok && value != "" {
		return value
	}

	return defaultValue
}

func (s Store) Expand(value string) string {
	return RawExpand(value, s.Lookup)
}

func (s Store) Copy(m Map) {
	for k, v := range m {
		s[ToKey(k)] = v
	}
}
