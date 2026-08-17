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
	return convert.ToAnySlice(value)
}

func (*Convert) Bools(value any) []bool {
	return convert.ToBoolSlice(value)
}

func (*Convert) Strings(value any) []string {
	return convert.ToStringSlice(value)
}

func (*Convert) Ints(value any) []int64 {
	return convert.ToInt64Slice(value)
}

func (*Convert) Floats(value any) []float64 {
	return convert.ToFloat64Slice(value)
}

func (*Convert) Durations(value any) []time.Duration {
	return convert.ToDurationSlice(value)
}

func (*Convert) Times(value any) []time.Time {
	return convert.ToTimeSlice(value)
}

func (*Convert) Runes(value any) []rune {
	return convert.ToRuneSlice(value)
}

func (*Convert) Bytes(value any) []byte {
	return convert.ToByteSlice(value)
}

func (*Convert) Map(value any) map[string]any {
	return convert.ToStringAnyMap(value)
}

func (*Convert) BoolMap(value any) map[string]bool {
	return convert.ToStringBoolMap(value)
}

func (*Convert) StringMap(value any) map[string]string {
	return convert.ToStringStringMap(value)
}

func (*Convert) IntMap(value any) map[string]int64 {
	return convert.ToStringInt64Map(value)
}

func (*Convert) FloatMap(value any) map[string]float64 {
	return convert.ToStringFloat64Map(value)
}

func (*Convert) DurationMap(value any) map[string]time.Duration {
	return convert.ToStringDurationMap(value)
}

func (*Convert) TimeMap(value any) map[string]time.Time {
	return convert.ToStringTimeMap(value)
}
