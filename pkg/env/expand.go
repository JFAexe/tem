package env

import (
	"fmt"
	"os"
	"strings"
	"unicode"
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
		out   strings.Builder
		runes = []rune(value)
	)

	for i := 0; i < len(runes); i++ {
		if runes[i] != '$' {
			out.WriteRune(runes[i])
			continue
		}

		if i+1 < len(runes) && runes[i+1] == '$' {
			out.WriteRune('$')

			i++

			continue
		}

		if i+1 < len(runes) && runes[i+1] == '{' {
			j := i + 2

			for j < len(runes) && runes[j] != '}' {
				j++
			}

			if j == len(runes) {
				for k := i; k < len(runes); k++ {
					out.WriteRune(runes[k])
				}

				break
			}

			out.WriteString(expandBrace(string(runes[i+2:j]), lookup))

			i = j

			continue
		}

		if i+1 < len(runes) && isVarStart(runes[i+1]) {
			var (
				start = i + 1
				j     = start + 1
			)

			for j < len(runes) && isVarPart(runes[j]) {
				j++
			}

			val, _ := lookup(string(runes[start:j]))

			out.WriteString(val)

			i = j - 1

			continue
		}

		out.WriteByte('$')
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
			idx, op = i, c

			break
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
