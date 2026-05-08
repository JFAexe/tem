package functions

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/JFAexe/tem/pkg/convert"
)

var ErrNilArgument = errors.New("nil arguments are not allowed")

type Math struct{}

func (*Math) Percent(part, total any) float64 {
	if t := convert.ToFloat64(total); t != 0 {
		return (convert.ToFloat64(part) / t) * 100
	}

	return 0
}

func (*Math) Clamp(mi, ma, v any) (result any, err error) {
	if mi == nil || ma == nil || v == nil {
		return nil, ErrNilArgument
	}

	var (
		rv = reflect.ValueOf(v)
		ri = reflect.ValueOf(mi)
		ra = reflect.ValueOf(ma)
	)

	switch {
	case rv.CanInt():
		if ri.CanFloat() {
			ri = ri.Convert(rv.Type())
		}

		if ra.CanFloat() {
			ra = ra.Convert(rv.Type())
		}

		if ri.CanInt() && ra.CanInt() {
			result = convert.Clamp(rv.Int(), ri.Int(), ra.Int())
		}
	case rv.CanFloat():
		if ri.CanInt() {
			ri = ri.Convert(rv.Type())
		}

		if ra.CanInt() {
			ra = ra.Convert(rv.Type())
		}

		if ri.CanFloat() && ra.CanFloat() {
			result = convert.Clamp(rv.Float(), ri.Float(), ra.Float())
		}
	default:
		err = fmt.Errorf("value type must be a number, got `%T`", v)
	}

	if result == nil && err == nil {
		err = fmt.Errorf("value type `%T` is incompatible with types for min `%T` and max `%T` values", v, mi, ma)
	}

	if err != nil {
		return v, err
	}

	return reflect.ValueOf(result).Convert(rv.Type()).Interface(), nil
}
