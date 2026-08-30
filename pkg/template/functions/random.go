package functions

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strings"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

const maxRangeLength = 1 << 20

var one = big.NewInt(1)

type Random struct{}

func (*Random) Pick(args ...any) (any, error) {
	if len(args) == 1 {
		rv := reflection.IndirectValue(args[0])
		if !rv.IsValid() {
			return nil, fmt.Errorf("%w: expected non-empty list, or map, or struct, or ...any", ErrValueRequired)
		}

		var values []any

		switch rv.Kind() {
		case reflect.Struct:
			values = make([]any, 0, rv.NumField())

			for _, field := range reflection.Fields(rv) {
				values = append(values, field.Interface())
			}
		case reflect.Map:
			values = make([]any, 0, rv.Len())

			for iter := rv.MapRange(); iter.Next(); {
				values = append(values, iter.Value().Interface())
			}
		default:
			values = convert.ToAnySlice(rv.Interface())
		}

		args = values
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("%w: expected non-empty list, or map, or struct, or ...any", ErrValueRequired)
	}

	idx, err := randInt64(int64(len(args)), false)
	if err != nil {
		return nil, err
	}

	return args[idx], nil
}

func (*Random) Shuffle(items any) ([]any, error) {
	out := convert.ToAnySlice(items)

	for i := len(out) - 1; i > 0; i-- {
		j, err := randInt64(int64(i+1), false)
		if err != nil {
			return nil, err
		}

		out[i], out[j] = out[j], out[i]
	}

	return out, nil
}

func (*Random) Int(args ...any) (int64, error) {
	return randInt64Range(convert.ToInt64Slice(args), false)
}

func (*Random) IntInclusive(args ...any) (int64, error) {
	return randInt64Range(convert.ToInt64Slice(args), true)
}

func (*Random) Float(args ...any) (float64, error) {
	return randFloat64Range(convert.ToFloat64Slice(args), false)
}

func (*Random) FloatInclusive(args ...any) (float64, error) {
	return randFloat64Range(convert.ToFloat64Slice(args), true)
}

func (f *Random) Bool(args ...any) (bool, error) {
	switch fs := convert.ToFloat64Slice(args); len(fs) {
	case 0:
		return randBool(0.5)
	default:
		return randBool(fs[len(fs)-1])
	}
}

func (f *Random) String(length any, args ...any) (_ string, err error) {
	var set []rune

	switch len(args) {
	case 0:
		if set, err = cachedRegexSet("[a-zA-Z0-9._-]"); err != nil {
			return "", err
		}
	case 1:
		if set, err = cachedRegexSet(convert.ToString(args[0])); err != nil {
			set = convert.ToRuneSlice(args[0])
		}
	default:
		set = cachedRangeSet(args[0], args[1])
	}

	return randString(convert.ToInt64(length), set)
}

func (f *Random) ASCII(length any) (string, error) {
	set, err := cachedRegexSet("[[:ascii:]]")
	if err != nil {
		return "", err
	}

	return randString(convert.ToInt64(length), set)
}

func (f *Random) Alpha(length any) (string, error) {
	set, err := cachedRegexSet("[[:alpha:]]")
	if err != nil {
		return "", err
	}

	return randString(convert.ToInt64(length), set)
}

func (f *Random) Numeric(length any) (string, error) {
	set, err := cachedRegexSet("[[:digit:]]")
	if err != nil {
		return "", err
	}

	return randString(convert.ToInt64(length), set)
}

func (f *Random) AlphaNumeric(length any) (string, error) {
	set, err := cachedRegexSet("[[:alnum:]]")
	if err != nil {
		return "", err
	}

	return randString(convert.ToInt64(length), set)
}

func (f *Random) Hex(length any) (string, error) {
	set, err := cachedRegexSet("[[:xdigit:]]")
	if err != nil {
		return "", err
	}

	return randString(convert.ToInt64(length), set)
}

func (f *Random) Graphic(length any) (string, error) {
	set, err := cachedRegexSet("[[:graph:]]")
	if err != nil {
		return "", err
	}

	return randString(convert.ToInt64(length), set)
}

func randInt64Range(args []int64, inclusive bool) (int64, error) {
	switch len(args) {
	case 0:
		return randInt64(math.MaxInt64, inclusive)
	case 1:
		upper := args[0]

		value, err := randInt64(upper, inclusive)
		if err != nil {
			return 0, fmt.Errorf("%w: upper=%d", err, upper)
		}

		return value, nil
	default:
		var (
			lower = slices.Min(args)
			upper = slices.Max(args)
		)

		if d := uint64(upper) - uint64(lower); d > math.MaxInt64 {
			return 0, ErrRangeTooLarge
		}

		if !inclusive && lower == upper {
			return 0, fmt.Errorf("%w: lower=%d upper=%d", ErrLowerGreaterEqual, lower, upper)
		}

		offset, err := randInt64(upper-lower, inclusive)
		if err != nil {
			return 0, fmt.Errorf("%w: lower=%d upper=%d", err, lower, upper)
		}

		return lower + offset, nil
	}
}

func randFloat64Range(args []float64, inclusive bool) (float64, error) {
	switch len(args) {
	case 0:
		return randFloat64(inclusive)
	case 1:
		upper := args[0]

		value, err := randFloat64(inclusive)
		if err != nil {
			return 0, fmt.Errorf("%w: upper=%f", err, upper)
		}

		return value * upper, nil
	default:
		var (
			lower = slices.Min(args)
			upper = slices.Max(args)
			diff  = upper - lower
		)

		if math.IsInf(diff, 1) || math.IsNaN(diff) {
			return 0, ErrRangeTooLarge
		}

		if !inclusive && lower == upper {
			return 0, fmt.Errorf("%w: lower=%f upper=%f", ErrLowerGreaterEqual, lower, upper)
		}

		offset, err := randFloat64(inclusive)
		if err != nil {
			return 0, fmt.Errorf("%w: lower=%f upper=%f", err, lower, upper)
		}

		return lower + offset*(upper-lower), nil
	}
}

func randInt64(n int64, inclusive bool) (int64, error) {
	if n < 0 || (n == 0 && !inclusive) {
		return 0, ErrUpperNegativeOrZero
	}

	upper := big.NewInt(int64(n))

	if inclusive {
		upper.Add(upper, one)
	}

	value, err := rand.Int(rand.Reader, upper)
	if err != nil {
		return 0, fmt.Errorf("failed to read crypto/rand: %w", err)
	}

	return value.Int64(), nil
}

func randFloat64(inclusive bool) (float64, error) {
	upper := new(big.Int).Lsh(one, 53)

	if inclusive {
		upper.Add(upper, one)
	}

	value, err := rand.Int(rand.Reader, upper)
	if err != nil {
		return 0, fmt.Errorf("failed to read crypto/rand: %w", err)
	}

	return float64(value.Int64()) / float64(1<<53), nil
}

func randBool(p float64) (bool, error) {
	if p <= 0.0 {
		return false, nil
	}

	if p >= 1.0 {
		return true, nil
	}

	float, err := randFloat64(false)
	if err != nil {
		return false, err
	}

	return float < p, nil
}

func randString(length int64, set []rune) (string, error) {
	if length = convert.Clamp(length, 0, maxRangeLength); len(set) == 0 || length == 0 {
		return "", nil
	}

	var (
		setLen = len(set)
		width  = 1
	)

	for 1<<(8*width) < setLen {
		width++
	}

	var (
		wordMax = uint64(1) << (8 * width)
		lim     = wordMax - wordMax%uint64(setLen)

		builder strings.Builder
		buf     = make([]byte, 64*width)
		pos     = len(buf)
	)

	builder.Grow(int(length) * 4)

	for i := int64(0); i < length; {
		if pos+width > len(buf) {
			if _, err := rand.Read(buf); err != nil {
				return "", err
			}

			pos = 0
		}

		var word uint64

		for k := range width {
			word = word<<8 | uint64(buf[pos+k])
		}

		pos += width

		if word < lim {
			builder.WriteRune(set[int(word%uint64(setLen))])

			i++
		}
	}

	return builder.String(), nil
}
