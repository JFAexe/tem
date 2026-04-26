package functions

import "github.com/JFAexe/tem/pkg/env"

type Env struct {
	envs env.Store
}

func (f Env) Escape(value string) string {
	return env.Escape(value)
}

func (f Env) Unescape(value string) string {
	return env.Unescape(value)
}

func (f Env) Map() env.Map {
	return f.envs.Environ()
}

func (f Env) Copy(m env.Map) Env {
	f.envs.Copy(m)

	return f
}

func (f Env) Set(key string, value any) Env {
	f.envs.Set(key, ToString(value))

	return f
}

func (f Env) IsSet(key string) bool {
	return f.envs.IsSet(key)
}

func (f Env) Get(key string) string {
	return f.envs.Get(key)
}

func (f Env) RawGet(key string) string {
	return f.envs.RawGet(key)
}

func (f Env) RawOr(def, key string) string {
	return f.envs.RawOr(key, def)
}

func (f Env) Or(def, key string) string {
	return f.envs.Or(key, def)
}

func (f Env) Expand(value string) string {
	return f.envs.Expand(value)
}
