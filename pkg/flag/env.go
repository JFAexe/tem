package flag

import (
	"flag"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/JFAexe/tem/pkg/env"
)

var _ flag.Value = (*EnvMap)(nil)

type EnvMap map[string]string

func (e *EnvMap) Set(arg string) error {
	if *e == nil {
		*e = make(EnvMap)
	}

	key, value, err := env.ParseKV(arg)
	if err != nil {
		return err
	}

	(*e)[key] = value

	return nil
}

func (e *EnvMap) String() string {
	var builder strings.Builder

	for i, key := range slices.Sorted(maps.Keys(*e)) {
		if i > 0 {
			builder.WriteByte('\n')
		}

		builder.WriteString(key)
		builder.WriteByte('=')

		if value, ok := (*e)[key]; ok {
			fmt.Fprint(&builder, strconv.Quote(value))
		}
	}

	return builder.String()
}
