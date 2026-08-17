package functions

import (
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

type Math struct{}

func (*Math) Add(values ...any) float64 {
	var sum float64

	for _, v := range values {
		for _, num := range convert.ToFloat64Slice(v) {
			sum += num
		}
	}

	return sum
}

func (*Math) Sub(y, x any) float64 {
	return convert.ToFloat64(x) - convert.ToFloat64(y)
}

func (*Math) Mul(values ...any) float64 {
	var (
		res   = 1.0
		count = 0
	)

	for _, v := range values {
		for _, num := range convert.ToFloat64Slice(v) {
			res *= num

			count++
		}
	}

	if count == 0 {
		return 0
	}

	return res
}

func (*Math) Div(y, x any) float64 {
	yf := convert.ToFloat64(y)

	if yf == 0 {
		return 0
	}

	return convert.ToFloat64(x) / yf
}

func (*Math) Pow(y, x any) float64 {
	var (
		xf = convert.ToFloat64(x)
		yf = convert.ToFloat64(y)
	)

	if math.IsNaN(xf) || math.IsInf(yf, 0) {
		return 0
	}

	if yf == 0 {
		return 1
	}

	return math.Pow(xf, yf)
}

func (*Math) Mod(y, x any) float64 {
	var (
		xf = convert.ToFloat64(x)
		yf = convert.ToFloat64(y)
	)

	if yf == 0 || math.IsNaN(xf) || math.IsInf(yf, 0) {
		return 0
	}

	return math.Mod(xf, yf)
}

func (*Math) Rem(y, x any) float64 {
	var (
		xf = convert.ToFloat64(x)
		yf = convert.ToFloat64(y)
	)

	if yf == 0 || math.IsNaN(xf) || math.IsInf(yf, 0) {
		return 0
	}

	return math.Remainder(xf, yf)
}

func (*Math) Round(value any, precision ...any) float64 {
	var (
		v = convert.ToFloat64(value)
		p = 0
	)

	if len(precision) > 0 {
		p = convert.ToInt(precision[0])
	}

	pow := math.Pow(10, convert.SafeIntToFloat64(p))

	return math.Round(v*pow) / pow
}

func (*Math) Floor(value any) float64 {
	return math.Floor(convert.ToFloat64(value))
}

func (*Math) Ceil(value any) float64 {
	return math.Ceil(convert.ToFloat64(value))
}

func (*Math) Abs(value any) float64 {
	return math.Abs(convert.ToFloat64(value))
}

func (*Math) Between(minimum, maximum, value any) bool {
	v := convert.ToFloat64(value)

	return v >= convert.ToFloat64(minimum) && v <= convert.ToFloat64(maximum)
}

func (*Math) Percent(part, total any) float64 {
	t := convert.ToFloat64(total)

	if t == 0 || math.IsNaN(t) || math.IsInf(t, 0) {
		return 0
	}

	if r := (convert.ToFloat64(part) / t) * 100; !math.IsNaN(r) && !math.IsInf(r, 0) {
		return r
	}

	return 0
}

func (*Math) Min(values ...any) float64 {
	var s []float64

	for _, v := range values {
		s = append(s, convert.ToFloat64Slice(v)...)
	}

	if len(s) == 0 {
		return 0
	}

	return slices.Min(s)
}

func (*Math) Max(values ...any) float64 {
	var s []float64

	for _, v := range values {
		s = append(s, convert.ToFloat64Slice(v)...)
	}

	if len(s) == 0 {
		return 0
	}

	return slices.Max(s)
}

func (*Math) Clamp(minimum, maximum, value any) (result float64, err error) {
	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
		return 0, fmt.Errorf("got invalid value: %w", reflection.ErrNilPointer)
	}

	if !rv.CanInt() && !rv.CanFloat() {
		return 0, fmt.Errorf("value type must be a number, got `%T`", value)
	}

	return convert.Clamp(convert.ToFloat64(value), convert.ToFloat64(minimum), convert.ToFloat64(maximum)), nil
}
