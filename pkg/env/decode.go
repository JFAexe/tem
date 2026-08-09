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
	ErrNotPair        = errors.New("not a key-value pair")
	ErrBadKey         = errors.New("bad key")
	ErrUnmatchedQuote = errors.New("unmatched quote in value")
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
		return fmt.Errorf("cannot decode into nil value")
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("decode requires a pointer, got %T", v)
	}

	if rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	var (
		re = rv.Elem()
		rk = re.Kind()
	)

	if rk != reflect.Interface && rk != reflect.Map {
		return fmt.Errorf("only *map[string]string or *any is supported, got %T", v)
	}

	if rk == reflect.Map && (re.Type().Key().Kind() != reflect.String || re.Type().Elem().Kind() != reflect.String) {
		return fmt.Errorf("only types like map[string]string are supported, got %s", re.Type())
	}

	out, err := d.decode()
	if err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	switch rk {
	case reflect.Interface:
		re.Set(reflect.ValueOf(out))
	case reflect.Map:
		if re.IsNil() {
			re.Set(reflect.MakeMap(re.Type()))
		}

		for k, val := range out {
			re.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(val))
		}
	}

	return nil
}

func (d *Decoder) decode() (Map, error) {
	var (
		multiline    bool
		quote        rune
		multilineKey string
		rawValue     strings.Builder

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

	flush := func() {
		raw := rawValue.String() + string(quote)

		parsed, err := ParseValue(raw)
		if err != nil {
			parsed = strings.TrimPrefix(raw, string(quote))
		}

		save(multilineKey, parsed)
	}

	for scanner.Scan() {
		line := scanner.Text()

		if multiline {
			trimmed := strings.TrimRight(line, " \t\r")
			if len(trimmed) == 0 {
				rawValue.WriteString("\n")
				rawValue.WriteString(line)

				continue
			}

			runes := []rune(trimmed)

			if idx := len(runes) - 1; runes[idx] == quote && !isEscaped(runes, idx) {
				rawValue.WriteString("\n")
				rawValue.WriteString(string(runes[:idx]))

				flush()

				multiline = false
				quote = 0
				multilineKey = ""

				rawValue.Reset()
			} else {
				rawValue.WriteString("\n")
				rawValue.WriteString(line)
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

		k, err := ParseKey(k)
		if err != nil {
			return nil, err
		}

		if v = strings.TrimSpace(v); v == "" {
			save(k, "")

			continue
		}

		if parsed, err := ParseValue(v); err == nil {
			save(k, parsed)
		} else if errors.Is(err, ErrUnmatchedQuote) {
			multiline = true
			quote = rune(v[0])
			multilineKey = k

			rawValue.WriteString(v)
		}
	}

	if multiline {
		flush()
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

func ParseMap(e Map) (m Map, err error) {
	m = make(Map, len(e))

	for key, val := range e {
		if key, err = ParseKey(key); err != nil {
			return nil, err
		}

		if val, err = ParseValue(val); err != nil {
			return nil, err
		}

		m[key] = val
	}

	return m, nil
}

func ParseLine(line string) (key, val string, err error) {
	key, val, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", ErrNotPair
	}

	if key, err = ParseKey(key); err != nil {
		return "", "", err
	}

	if val, err = ParseValue(val); err != nil {
		return "", "", err
	}

	return key, val, nil
}

func ParseKey(s string) (string, error) {
	if key := ToKey(s); key != "" {
		return key, nil
	}

	return "", ErrBadKey
}

func ParseValue(s string) (string, error) {
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
