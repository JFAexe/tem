package flag

import (
	"flag"
	"slices"
	"strconv"
	"strings"
)

var _ flag.Value = (*StringSlice)(nil)

type StringSlice []string

func (e *StringSlice) Set(value string) error {
	if strings.TrimSpace(value) != "" {
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
