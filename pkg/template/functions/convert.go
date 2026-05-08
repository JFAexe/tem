package functions

import (
	"time"

	"github.com/JFAexe/tem/pkg/convert"
)

type Convert struct{}

func (*Convert) Any(value any) any {
	return convert.ToAny(value)
}

func (*Convert) Bool(value any) bool {
	return convert.ToBool(value)
}

func (*Convert) String(value any) string {
	return convert.ToString(value)
}

func (*Convert) Rune(value any) rune {
	return convert.ToRune(value)
}

func (*Convert) Int(value any) int64 {
	return convert.ToInt64(value)
}

func (*Convert) Float(value any) float64 {
	return convert.ToFloat64(value)
}

func (*Convert) Duration(value any) time.Duration {
	return convert.ToDuration(value)
}

func (*Convert) Time(value any) time.Time {
	return convert.ToTime(value)
}

func (*Convert) List(value any) []any {
	return convert.ToAnyList(value)
}

func (*Convert) Bools(value any) []bool {
	return convert.ToBoolList(value)
}

func (*Convert) Strings(value any) []string {
	return convert.ToStringList(value)
}

func (*Convert) Ints(value any) []int64 {
	return convert.ToInt64List(value)
}

func (*Convert) Floats(value any) []float64 {
	return convert.ToFloat64List(value)
}

func (*Convert) Runes(value any) []rune {
	return convert.ToRuneList(value)
}

func (*Convert) Map(value any) map[string]any {
	return convert.ToAnyMap(value)
}

func (*Convert) BoolMap(value any) map[string]bool {
	return convert.ToBoolMap(value)
}

func (*Convert) StringMap(value any) map[string]string {
	return convert.ToStringMap(value)
}

func (*Convert) IntMap(value any) map[string]int64 {
	return convert.ToInt64Map(value)
}

func (*Convert) FloatMap(value any) map[string]float64 {
	return convert.ToFloat64Map(value)
}
