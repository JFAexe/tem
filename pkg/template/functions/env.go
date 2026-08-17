package functions

import (
	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/env"
)

type Env struct{}

func EnvVarargInit(n *Env, args []any) (any, error) {
	return n.Get(convert.ToStringSlice(args)[0]), nil
}

func (*Env) Escape(value any) string {
	return env.Escape(convert.ToString(value))
}

func (*Env) Unescape(value any) string {
	return env.Unescape(convert.ToString(value))
}

func (*Env) ToKey(key any) string {
	return env.ToKey(convert.ToString(key))
}

func (*Env) Expand(value any) string {
	return env.Expand(convert.ToString(value))
}

func (*Env) Map() (env.Map, error) {
	return env.Environ()
}

func (*Env) Set(key, value any) (bool, error) {
	err := env.Set(convert.ToString(key), convert.ToString(value))

	return err == nil, err
}

func (*Env) BatchSet(m any) (bool, error) {
	err := env.BatchSet(convert.ToStringStringMap(m))

	return err == nil, err
}

func (*Env) Unset(key any) (bool, error) {
	err := env.Unset(convert.ToString(key))

	return err == nil, err
}

func (*Env) BatchUnset(keys ...any) (bool, error) {
	err := env.BatchUnset(convert.ToStringSlice(listConcat(keys...)))

	return err == nil, err
}

func (*Env) IsSet(key any) bool {
	return env.IsSet(convert.ToString(key))
}

func (*Env) Get(key any) string {
	return env.Get(convert.ToString(key))
}

func (*Env) RawGet(key any) string {
	return env.RawGet(convert.ToString(key))
}

func (*Env) RawOr(def, key any) string {
	return env.RawOr(convert.ToString(key), convert.ToString(def))
}

func (*Env) Or(def, key any) string {
	return env.Or(convert.ToString(key), convert.ToString(def))
}
