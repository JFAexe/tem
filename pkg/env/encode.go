package env

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

var quoteReplacer = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

type EncoderOption func(e *Encoder)

func WithEncoderLookup(lookup LookupFunc) EncoderOption {
	return func(e *Encoder) {
		e.lookup = lookup
	}
}

func WithEncoderExpand(val bool) EncoderOption {
	return func(e *Encoder) {
		e.expand = val
	}
}

type Encoder struct {
	w      io.Writer
	lookup LookupFunc
	expand bool
}

func NewEncoder(w io.Writer, options ...EncoderOption) *Encoder {
	e := &Encoder{
		w:      w,
		expand: true,
		lookup: RawLookup,
	}

	for _, option := range options {
		option(e)
	}

	return e
}

func (e *Encoder) Encode(v any) error {
	m, ok := v.(Map)
	if !ok {
		return fmt.Errorf("encode requires map[string]string, got %T", v)
	}

	var buf bytes.Buffer

	for i, key := range slices.Sorted(maps.Keys(m)) {
		if i > 0 {
			buf.WriteByte('\n')
		}

		val := m[key]

		if e.expand {
			expanded, err := RawExpand(val, e.lookup)
			if err != nil {
				return fmt.Errorf("failed to expand value %#q: %w", val, err)
			}

			val = expanded
		}

		fmt.Fprintf(&buf, `%s="%s"`, ToKey(key), quoteReplacer.Replace(val))
	}

	if _, err := buf.WriteTo(e.w); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}

	return nil
}

func Marshal(value Map) ([]byte, error) {
	return MarshalOptions(value)
}

func MarshalOptions(value Map, options ...EncoderOption) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := NewEncoder(buf, options...).Encode(value); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
