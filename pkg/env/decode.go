package env

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode"
)

var (
	ErrNotKV           = errors.New("not an env key-value")
	ErrInvalidKVFormat = errors.New("invalid env key-value format")
	ErrUnmatchedQuote  = errors.New("unmatched quote in value")
)

type DecoderOption = func(d *Decoder)

func WithDecoderLookup(lookup LookupFunc) DecoderOption {
	return func(d *Decoder) {
		d.lookup = lookup
	}
}

func WithDecoderExpand(val bool) DecoderOption {
	return func(d *Decoder) {
		d.expand = val
	}
}

type Decoder struct {
	r      io.Reader
	lookup LookupFunc
	expand bool
}

func NewDecoder(r io.Reader, options ...DecoderOption) *Decoder {
	d := &Decoder{
		r:      r,
		expand: true,
		lookup: Lookup,
	}

	for _, option := range options {
		option(d)
	}

	return d
}

func (d *Decoder) Decode(v any) error {
	if v == nil {
		return fmt.Errorf("env: cannot decode into nil")
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("env: decode requires a pointer, got %T", v)
	}

	if rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	out, err := d.decode()
	if err != nil {
		return fmt.Errorf("env: failed to decode data: %w", err)
	}

	switch target := rv.Elem(); target.Kind() {
	case reflect.Interface:
		target.Set(reflect.ValueOf(out))
	case reflect.Map:
		if target.Type().Key().Kind() != reflect.String || target.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("env: only types like map[string]string are supported, got %s", target.Type())
		}

		if target.IsNil() {
			target.Set(reflect.MakeMap(target.Type()))
		}

		for k, val := range out {
			target.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(val))
		}
	default:
		return fmt.Errorf("env: only *map[string]string or *any is supported, got %T", v)
	}

	return nil
}

func (d *Decoder) decode() (Map, error) {
	var (
		multiline bool
		quote     rune
		key       string
		buf       []string

		out     = make(Map)
		scanner = bufio.NewScanner(d.r)
	)

	lookup := func(key string) (string, bool) {
		key = ToKey(key)

		if val, ok := out[key]; ok {
			return val, true
		}

		if d.lookup != nil {
			return d.lookup(key)
		}

		return "", false
	}

	save := func(key, val string) {
		if d.expand {
			val = RawExpand(val, lookup)
		}

		out[key] = val
	}

	saveMultiline := func(key string, buffer []string) {
		inner := strings.Join(buffer, "\n")

		parsed, err := parseQuotedValue(string(quote) + inner + string(quote))
		if err != nil {
			parsed = inner
		}

		save(key, parsed)
	}

	for scanner.Scan() {
		line := scanner.Text()

		if multiline {
			buf = append(buf, line)

			if line = strings.TrimRight(line, " \t\r"); len(line) == 0 {
				continue
			}

			runes := []rune(line)

			if idx := len(runes) - 1; runes[idx] == quote && !isEscaped(runes, idx) {
				buf[len(buf)-1] = string(runes[:idx])

				saveMultiline(key, buf)

				multiline = false
				quote = 0
				key = ""
				buf = nil
			}

			continue
		}

		if line = strings.TrimSpace(line); line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		if k = ToKey(strings.TrimSpace(k)); k == "" {
			continue
		}

		if v = strings.TrimSpace(v); v == "" {
			save(k, "")

			continue
		}

		switch runes := []rune(v); runes[0] {
		case '"', '\'':
			if parsed, err := parseValue(v); err == nil {
				save(k, parsed)
			} else if errors.Is(err, ErrUnmatchedQuote) {
				multiline = true
				quote = runes[0]
				key = k
				buf = []string{string(runes[1:])}
			}

			continue
		}

		if parsed, err := parseValue(v); err == nil {
			save(k, parsed)
		}
	}

	if multiline {
		saveMultiline(key, buf)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func Unmarshal(data []byte, v any, options ...DecoderOption) error {
	if err := NewDecoder(bytes.NewReader(data), options...).Decode(v); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func isEscaped(r []rune, i int) bool {
	if i == 0 {
		return false
	}

	var count int

	for j := i - 1; j >= 0 && r[j] == '\\'; j-- {
		count++
	}

	return count%2 == 1
}

func parseValue(s string) (string, error) {
	if s = strings.TrimSpace(s); s == "" {
		return s, nil
	}

	runes := []rune(s)

	if r := runes[0]; r == '"' || r == '\'' {
		return parseQuotedValue(s)
	}

	var (
		b       strings.Builder
		escaped bool
	)

	for i, r := range runes {
		if escaped {
			escaped = false

			b.WriteRune(r)

			continue
		}

		if r == '\\' {
			escaped = true

			continue
		}

		if r == '#' && (i == 0 || unicode.IsSpace(runes[i-1])) {
			break
		}

		b.WriteRune(r)
	}

	return strings.TrimSpace(b.String()), nil
}

func parseQuotedValue(s string) (string, error) {
	var (
		b       strings.Builder
		escaped bool

		runes = []rune(s)
		quote = runes[0]
		tail  = runes[1:]
	)

	for i, r := range tail {
		if escaped {
			escaped = false

			b.WriteRune(r)

			continue
		}

		if r == '\\' {
			escaped = true

			continue
		}

		if r != quote {
			b.WriteRune(r)

			continue
		}

		remaining := string(tail[i+1:])

		if idx := strings.IndexByte(remaining, '#'); idx >= 0 {
			if pre := remaining[:idx]; strings.TrimSpace(pre) == "" {
				remaining = ""
			}
		}

		if strings.TrimSpace(remaining) != "" {
			return "", fmt.Errorf("unexpected characters after closing quote: %#q", remaining)
		}

		return b.String(), nil
	}

	return "", ErrUnmatchedQuote
}
