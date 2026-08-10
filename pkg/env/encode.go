package env

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"slices"
)

type EncoderOption = func(e *Encoder)

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
		lookup: Lookup,
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
			fmt.Fprint(&buf, "\n")
		}

		val := m[key]

		if e.expand {
			val = RawExpand(val, e.lookup)
		}

		fmt.Fprintf(&buf, "%s=%q", ToKey(key), val)
	}

	if _, err := buf.WriteTo(e.w); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}

	return nil
}

func Marshal(value Map, options ...EncoderOption) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := NewEncoder(buf, options...).Encode(value); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
