package flag

import (
	"flag"
	"fmt"

	"github.com/JFAexe/tem/pkg/env"
)

var (
	_ fmt.Stringer = (*EnvMap)(nil)
	_ flag.Value   = (*EnvMap)(nil)
)

type EnvMap map[string]string

func (e EnvMap) Set(arg string) error {
	return env.Unmarshal([]byte(arg), &e, env.WithDecoderExpand(false))
}

func (e EnvMap) String() string {
	raw, _ := env.Marshal(e, env.WithEncoderExpand(false))

	return string(raw)
}
