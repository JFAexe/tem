package env

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
)

type EncoderOption = func(d *Encoder)

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
		return fmt.Errorf("env: encode requires map[string]string, got %T", v)
	}

	for i, key := range slices.Sorted(maps.Keys(m)) {
		if i > 0 {
			if _, err := fmt.Fprint(e.w, "\n"); err != nil {
				return fmt.Errorf("env: encoding error: %w", err)
			}
		}

		val := m[key]

		if e.expand {
			val = RawExpand(val, e.lookup)
		}

		if _, err := fmt.Fprintf(e.w, "%s=%s", key, strconv.Quote(val)); err != nil {
			return fmt.Errorf("env: encoding error: %w", err)
		}
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
