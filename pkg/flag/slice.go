package flag

import (
	"flag"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var (
	_ fmt.Stringer = (*StringSlice)(nil)
	_ flag.Value   = (*StringSlice)(nil)
)

type StringSlice []string

func (e *StringSlice) Set(value string) error {
	if value = strings.TrimSpace(value); value != "" {
		*e = append(*e, value)
	}

	return nil
}

func (e *StringSlice) String() string {
	if e == nil || len(*e) == 0 {
		return ""
	}

	values := slices.Clone(*e)

	for i, str := range values {
		values[i] = strconv.Quote(str)
	}

	return strings.Join(values, ", ")
}
