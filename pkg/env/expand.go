package env

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

var expandOps = []string{
	":-", "-",
	":=", "=",
	":+", "+",
	":?", "?",
}

func RawExpand(value string, lookup LookupFunc) string {
	if !strings.Contains(value, "$") {
		return value
	}

	if lookup == nil {
		lookup = noopLookup
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

			out.WriteString(expandBrace(value[s:j], lookup))

			i = j + 1

			continue
		}

		if isVarStart(nr) {
			var (
				s = i + 1
				j = s + ns
			)

			for j < n {
				rj, sj := utf8.DecodeRuneInString(value[j:])

				if !isVarPart(rj) {
					break
				}

				j += sj
			}

			v, _ := lookup(value[s:j])

			out.WriteString(v)

			i = j

			continue
		}

		out.WriteByte('$')

		i += size
	}

	return out.String()
}

func expandBrace(expr string, lookup LookupFunc) string {
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
		val, _ := lookup(strings.TrimSpace(expr))

		return val
	}

	var (
		name    = strings.TrimSpace(expr[:idx])
		value   = expr[idx+len(op):]
		val, ok = lookup(name)
		unset   = !ok
		empty   = ok && val == ""
	)

	switch op {
	case ":-":
		if unset || empty {
			return value
		}
	case "-":
		if unset {
			return value
		}
	case ":=":
		if unset || empty {
			if err := Set(name, RawExpand(value, lookup)); err != nil {
				exit(name, "failed to set env")
			}

			return value
		}
	case "=":
		if unset {
			if err := Set(name, RawExpand(value, lookup)); err != nil {
				exit(name, "failed to set env")
			}

			return value
		}
	case ":+":
		if !unset && !empty {
			return value
		}

		return ""
	case "+":
		if !unset {
			return value
		}

		return ""
	case ":?":
		if unset || empty {
			exit(name, value)
		}
	case "?":
		if unset {
			exit(name, value)
		}
	}

	return val
}

func noopLookup(value string) (string, bool) {
	return value, true
}

func isVarStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isVarPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func exit(name, message string) {
	if message == "" {
		message = "parameter is null or not set"
	}

	fmt.Fprintf(os.Stderr, "%s: %s: %s\n", os.Args[0], name, message)

	os.Exit(1)
}
