package functions

import (
	"crypto/rand"
	"errors"
	"fmt"
	"maps"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strings"

	"github.com/JFAexe/tem/pkg/convert"
)

var (
	ErrEmptyList           = errors.New("can't select value from empty list")
	ErrUpperNegativeOrZero = errors.New("upper boundary must be greater than 0")
	ErrLowerGreaterEqual   = errors.New("lower boundary must be less than upper boundary")
	ErrRangeTooLarge       = errors.New("range is too large")
)

type Random struct {
	runes *Rune
}

func NewRandomFuncs(runeFuncs *Rune) *Random {
	return &Random{
		runes: runeFuncs,
	}
}

func (f *Random) Pick(values ...any) (any, error) {
	return f.PickFrom(values)
}

func (*Random) PickFrom(value any) (any, error) {
	if reflect.ValueOf(value).Kind() == reflect.Map {
		value = slices.Sorted(maps.Keys(convert.ToAnyMap(value)))
	}

	var (
		values = convert.ToAnyList(value)
		count  = int64(len(values))
	)

	if count == 0 {
		return nil, ErrEmptyList
	}

	idx, err := randInt64(count, false)
	if err != nil {
		return nil, err
	}

	return values[convert.Clamp(idx, 0, count)], nil
}

func (*Random) Int(args ...int64) (int64, error) {
	return randInt64Range(args, false)
}

func (*Random) IntInclusive(args ...int64) (int64, error) {
	return randInt64Range(args, true)
}

func (*Random) Float(args ...float64) (float64, error) {
	return randFloat64Range(args, false)
}

func (*Random) FloatInclusive(args ...float64) (float64, error) {
	return randFloat64Range(args, true)
}

func (f *Random) Bool(args ...float64) (bool, error) {
	switch len(args) {
	case 0:
		return randBool(0.5)
	default:
		return randBool(args[len(args)-1])
	}
}

func (f *Random) String(length int64, args ...any) (_ string, err error) {
	var set []rune

	switch len(args) {
	case 0:
		if set, err = f.runes.RegexSet(`[a-zA-Z0-9._-]`); err != nil {
			return "", err
		}
	case 2:
		set = f.runes.rangeSet(convert.ToRune(args[0]), convert.ToRune(args[1]))
	default:
		set = convert.ToRuneList(args[0])
	}

	return randString(length, set)
}

func (f *Random) ASCII(length int64) (string, error) {
	set, err := f.runes.RegexSet(`[[:ascii:]]`)
	if err != nil {
		return "", err
	}

	return randString(length, set)
}

func (f *Random) Alpha(length int64) (string, error) {
	set, err := f.runes.RegexSet(`[[:alpha:]]`)
	if err != nil {
		return "", err
	}

	return randString(length, set)
}

func (f *Random) Numeric(length int64) (string, error) {
	set, err := f.runes.RegexSet(`[[:digit:]]`)
	if err != nil {
		return "", err
	}

	return randString(length, set)
}

func (f *Random) AlphaNumeric(length int64) (string, error) {
	set, err := f.runes.RegexSet(`[[:alnum:]]`)
	if err != nil {
		return "", err
	}

	return randString(length, set)
}

func (f *Random) Hex(length int64) (string, error) {
	set, err := f.runes.RegexSet(`[[:xdigit:]]`)
	if err != nil {
		return "", err
	}

	return randString(length, set)
}

func (f *Random) Graphic(length int64) (string, error) {
	set, err := f.runes.RegexSet(`[[:graph:]]`)
	if err != nil {
		return "", err
	}

	return randString(length, set)
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

		if upper > math.MaxInt-lower {
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
		)

		if upper > math.MaxFloat64-lower {
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
	switch {
	case n < 0:
		return 0, ErrUpperNegativeOrZero
	case n == 0:
		if inclusive {
			return 0, nil
		}

		return 0, ErrUpperNegativeOrZero
	}

	upper := big.NewInt(int64(n))

	if inclusive {
		upper.Add(upper, big.NewInt(1))
	}

	value, err := rand.Int(rand.Reader, upper)
	if err != nil {
		return 0, fmt.Errorf("failed to read crypto/rand: %w", err)
	}

	return clampBigInt(value, 0, math.MaxInt64).Int64(), nil
}

func randFloat64(inclusive bool) (float64, error) {
	upper := big.NewInt(math.MaxInt64)

	if inclusive {
		upper.Add(upper, big.NewInt(1))
	}

	value, err := rand.Int(rand.Reader, upper)
	if err != nil {
		return 0, fmt.Errorf("failed to read crypto/rand: %w", err)
	}

	float := new(big.Float).SetInt(value)
	float.Quo(float, new(big.Float).SetInt(upper))

	result, _ := clampBigFloat(float, 0, 1).Float64()

	return result, nil
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
	length = max(0, length)

	if len(set) == 0 {
		return "", nil
	}

	var (
		builder strings.Builder

		count = big.NewInt(int64(len(set)))
	)

	for range length {
		idx, err := rand.Int(rand.Reader, count)
		if err != nil {
			return "", err
		}

		builder.WriteRune(set[idx.Int64()])
	}

	return builder.String(), nil
}

func clampBigInt(v *big.Int, mi, ma int64) *big.Int {
	if m := big.NewInt(mi); v.Cmp(m) < 0 {
		return m
	}

	if m := big.NewInt(ma); v.Cmp(m) > 0 {
		return m
	}

	return v
}

func clampBigFloat(v *big.Float, mi, ma float64) *big.Float {
	if m := big.NewFloat(mi); v.Cmp(m) < 0 {
		return m
	}

	if m := big.NewFloat(ma); v.Cmp(m) > 0 {
		return m
	}

	return v
}
