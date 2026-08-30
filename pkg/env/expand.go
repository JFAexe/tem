package env

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUnset       = errors.New("parameter is null or not set")
	ErrEmptyName   = fmt.Errorf("empty variable name")
	ErrExpandDepth = errors.New("expansion depth limit exceeded")
)

const maxExpandDepth = 16

var expandOps = []string{
	":-", "-",
	":=", "=",
	":+", "+",
	":?", "?",
}

type expander struct {
	lookup LookupFunc
	depth  int
}

func (e *expander) expand(value string) (string, error) {
	if e.depth > maxExpandDepth {
		return "", fmt.Errorf("%w (%d)", ErrExpandDepth, maxExpandDepth)
	}

	if !strings.Contains(value, "$") {
		return value, nil
	}

	var (
		out strings.Builder

		n = len(value)
	)

	out.Grow(n)

	for i := 0; i < n; {
		r, size := utf8.DecodeRuneInString(value[i:])
		if r != '$' {
			out.WriteString(value[i : i+size])

			i += size

			continue
		}

		if i+1 >= n {
			out.WriteByte('$')

			break
		}

		nr, ns := utf8.DecodeRuneInString(value[i+1:])

		if nr == '$' {
			out.WriteByte('$')

			i += 1 + ns

			continue
		}

		if nr == '{' {
			var (
				s = i + 1 + ns
				j = s
			)

			for j < n {
				rj, sj := utf8.DecodeRuneInString(value[j:])

				if rj == '}' {
					break
				}

				j += sj
			}

			if j >= n {
				out.WriteString(value[i:])

				break
			}

			v, err := e.expandBrace(value[s:j])
			if err != nil {
				return "", err
			}

			out.WriteString(v)

			i = j + 1

			continue
		}

		if unicode.IsLetter(nr) || nr == '_' {
			var (
				s = i + 1
				j = s + ns
			)

			for j < n {
				rj, sj := utf8.DecodeRuneInString(value[j:])

				if !unicode.IsLetter(rj) && !unicode.IsDigit(rj) && rj != '_' {
					break
				}

				j += sj
			}

			v, err := e.fetch(value[s:j])
			if err != nil {
				return "", err
			}

			out.WriteString(v)

			i = j

			continue
		}

		out.WriteByte('$')

		i += size
	}

	return out.String(), nil
}

func (e *expander) fetch(name string) (string, error) {
	v, ok := e.lookup(name)
	if !ok {
		return "", nil
	}

	return (&expander{lookup: e.lookup, depth: e.depth + 1}).expand(v)
}

func (e *expander) expandBrace(expr string) (string, error) {
	var (
		op string

		idx = -1
	)

	for _, c := range expandOps {
		if i := strings.Index(expr, c); i > 0 {
			if idx == -1 || i < idx {
				idx, op = i, c
			}
		}
	}

	if idx <= 0 {
		return e.fetch(strings.TrimSpace(expr))
	}

	var (
		name    = strings.TrimSpace(expr[:idx])
		word    = expr[idx+len(op):]
		val, ok = e.lookup(name)
		unset   = !ok
		empty   = ok && val == ""
	)

	switch op {
	case ":-":
		if name == "" {
			return "", ErrEmptyName
		}

		if unset || empty {
			return e.expand(word)
		}
	case "-":
		if unset {
			return e.expand(word)
		}
	case ":=":
		if unset || empty {
			w, err := e.expand(word)
			if err != nil {
				return "", err
			}

			if err := Set(name, w); err != nil {
				return "", fmt.Errorf("expand %s: %w", name, err)
			}

			return w, nil
		}
	case "=":
		if unset {
			w, err := e.expand(word)
			if err != nil {
				return "", err
			}

			if err := Set(name, w); err != nil {
				return "", fmt.Errorf("expand %s: %w", name, err)
			}

			return w, nil
		}
	case ":+":
		if !unset && !empty {
			return e.expand(word)
		}

		return "", nil
	case "+":
		if !unset {
			return e.expand(word)
		}

		return "", nil
	case ":?":
		if unset || empty {
			msg, err := e.expand(word)
			if err != nil {
				return "", err
			}

			if msg == "" {
				msg = "parameter is null or not set"
			}

			return "", fmt.Errorf("%w: %s: %s", ErrUnset, name, msg)
		}
	case "?":
		if unset {
			msg, err := e.expand(word)
			if err != nil {
				return "", err
			}

			if msg == "" {
				msg = "parameter is null or not set"
			}

			return "", fmt.Errorf("%w: %s: %s", ErrUnset, name, msg)
		}
	}

	return val, nil
}
