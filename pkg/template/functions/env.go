package functions

import "github.com/JFAexe/tem/pkg/env"

type Env struct{}

func EnvNamespace() func(...any) any {
	n := new(Env)

	return func(args ...any) any {
		if len(args) > 0 {
			return n.Get(ToStringList(args)[0])
		}

		return n
	}
}

func (*Env) Escape(value string) string {
	return env.Escape(value)
}

func (*Env) Unescape(value string) string {
	return env.Unescape(value)
}

func (*Env) ToKey(key string) string {
	return env.ToKey(key)
}

func (*Env) Expand(value string) string {
	return env.Expand(value)
}

func (*Env) Map() (env.Map, error) {
	return env.Environ()
}

func (*Env) Set(key string, value any) error {
	return env.Set(key, ToString(value))
}

func (*Env) BatchSet(m any) error {
	return env.BatchSet(ToStringMap(m))
}

func (*Env) Unset(key string) error {
	return env.Unset(key)
}

func (*Env) BatchUnset(keys ...any) error {
	return env.BatchUnset(ToStringList(keys))
}

func (*Env) IsSet(key string) bool {
	return env.IsSet(key)
}

func (*Env) Get(key string) string {
	return env.Get(key)
}

func (*Env) RawGet(key string) string {
	return env.RawGet(key)
}

func (*Env) RawOr(def, key string) string {
	return env.RawOr(key, def)
}

func (*Env) Or(def, key string) string {
	return env.Or(key, def)
}
