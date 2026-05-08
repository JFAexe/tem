package convert

import (
	"cmp"
	"fmt"
	"maps"
	"math"
	"reflect"
	"slices"
	"strconv"
	"time"
	"unicode/utf8"
)

const (
	MinInt32ToFloat32 = -(1 << 23)
	MaxInt32ToFloat32 = 1<<23 - 1
	MinInt64ToFloat64 = -(1 << 53)
	MaxInt64ToFloat64 = 1<<53 - 1
)

var timeLayouts = []string{
	"2006-01-02T15:04:05",
	time.DateTime,
	time.RFC3339,
	time.RFC3339Nano,
	time.ANSIC,
	time.UnixDate,
}

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type ConvertFunc[T any] = func(any) T

func ToAny(v any) any {
	return v
}

func ToBool(value any) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		b, _ := strconv.ParseBool(v)

		return b
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	}

	b, _ := strconv.ParseBool(ToString(value))

	return b
}

func ToString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case []rune:
		return string(v)
	case fmt.Stringer:
		return v.String()
	case fmt.GoStringer:
		return v.GoString()
	}

	if rv := reflect.ValueOf(value); rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return ""
		}

		return ToString(rv.Elem().Interface())
	}

	return fmt.Sprint(value)
}

func ToRune(value any) rune {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return SafeInt32(reflect.ValueOf(v).Int())
	case float32, float64:
		return ToInt32(reflect.ValueOf(v).Float())
	case string:
		if v != "" {
			r, _ := utf8.DecodeRuneInString(v)

			return r
		}
	default:
		return ToRune(ToString(value))
	}

	return 0
}

func Clamp[T cmp.Ordered](v, mi, ma T) T {
	return min(max(v, mi), ma)
}

func SafeInt[T Int](v T) int {
	return int(Clamp(int64(v), math.MinInt, math.MaxInt))
}

func SafeInt32[T Int](v T) int32 {
	return int32(Clamp(int64(v), math.MinInt32, math.MaxInt32))
}

func SafeFloat32[T Int](v T) float32 {
	return float32(Clamp(int64(v), MinInt32ToFloat32, MaxInt32ToFloat32))
}

func SafeFloat64[T Int](v T) float64 {
	return float64(Clamp(int64(v), MinInt64ToFloat64, MaxInt64ToFloat64))
}

func ToInt(value any) int {
	return SafeInt(ToInt64(value))
}

func ToInt32(value any) int32 {
	return SafeInt32(ToInt64(value))
}

func ToInt64(value any) int64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int()
	case float32, float64:
		return int64(math.Trunc(reflect.ValueOf(v).Float()))
	case string:
		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return i
		}

		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int64(f)
		}
	case bool:
		if v {
			return 1
		}
	}

	if i, err := strconv.ParseInt(ToString(value), 0, 64); err == nil {
		return i
	}

	return 0
}

func ToFloat32(value any) float32 {
	return float32(Clamp(ToFloat64(value), -math.MaxFloat32, math.MaxFloat32))
}

func ToFloat64(value any) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case float32, float64:
		return reflect.ValueOf(v).Float()
	case int, int8, int16, int32, int64:
		return SafeFloat64(reflect.ValueOf(v).Int())
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	case bool:
		if v {
			return 1
		}
	}

	if f, err := strconv.ParseFloat(ToString(value), 64); err == nil {
		return f
	}

	return 0
}

func ToDuration(value any) time.Duration {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case time.Duration:
		return v
	case int, int8, int16, int32, int64:
		return time.Duration(reflect.ValueOf(v).Int())
	case float32, float64:
		return time.Duration(reflect.ValueOf(v).Float() * float64(time.Second))
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}

		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return time.Duration(i)
		}
	}

	return time.Duration(ToInt64(value))
}

func ToTime(value any) time.Time {
	if value == nil {
		return time.Time{}
	}

	switch v := value.(type) {
	case time.Time:
		return v
	case string:
		for _, layout := range timeLayouts {
			if t, err := time.Parse(layout, v); err == nil {
				return t
			}
		}

		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return time.Unix(i, 0)
		}
	case int, int8, int16, int32, int64:
		return time.Unix(reflect.ValueOf(v).Int(), 0)
	case float32, float64:
		sec := ToInt64(v)

		return time.Unix(sec, int64((reflect.ValueOf(v).Float()-float64(sec))*1e9))
	case time.Duration:
		return time.Unix(0, int64(v))
	}

	return ToTime(ToString(value))
}

func ToAnyList(value any) []any {
	return ToList(value, ToAny)
}

func ToBoolList(value any) []bool {
	return ToList(value, ToBool)
}

func ToStringList(value any) []string {
	return ToList(value, ToString)
}

func ToIntList(value any) []int {
	return ToList(value, ToInt)
}

func ToInt32List(value any) []int32 {
	return ToList(value, ToInt32)
}

func ToInt64List(value any) []int64 {
	return ToList(value, ToInt64)
}

func ToFloat32List(value any) []float32 {
	return ToList(value, ToFloat32)
}

func ToFloat64List(value any) []float64 {
	return ToList(value, ToFloat64)
}

func ToRuneList(value any) []rune {
	switch v := value.(type) {
	case []rune:
		return v
	case int, int8, int16, int32, int64:
		return []rune{ToRune(v)}
	case []int, []int8, []int16, []int64:
		var (
			rv    = reflect.ValueOf(value)
			runes = make([]rune, 0, rv.Len())
		)

		for i := range rv.Len() {
			if item := rv.Index(i).Interface(); item != nil {
				runes = append(runes, ToRune(item))
			}
		}

		return runes
	case string:
		return []rune(v)
	}

	return ToRuneList(ToString(value))
}

func ToList[T any, L []T](value any, fn ConvertFunc[T]) L {
	if value == nil {
		return make(L, 0)
	}

	if v, ok := value.(L); ok {
		return slices.Clone(v)
	}

	rv := reflect.ValueOf(value)

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return L{fn(value)}
	}

	out := make(L, 0, rv.Len())

	for i := range rv.Len() {
		if item := rv.Index(i).Interface(); item != nil {
			out = append(out, fn(item))
		}
	}

	return out
}

func ToAnyMap(value any) map[string]any {
	return ToMap(value, ToAny)
}

func ToBoolMap(value any) map[string]bool {
	return ToMap(value, ToBool)
}

func ToStringMap(value any) map[string]string {
	return ToMap(value, ToString)
}

func ToIntMap(value any) map[string]int {
	return ToMap(value, ToInt)
}

func ToInt32Map(value any) map[string]int32 {
	return ToMap(value, ToInt32)
}

func ToInt64Map(value any) map[string]int64 {
	return ToMap(value, ToInt64)
}

func ToFloat32Map(value any) map[string]float32 {
	return ToMap(value, ToFloat32)
}

func ToFloat64Map(value any) map[string]float64 {
	return ToMap(value, ToFloat64)
}

func ToDurationMap(value any) map[string]time.Duration {
	return ToMap(value, ToDuration)
}

func ToMap[T any, M map[string]T](value any, fn ConvertFunc[T]) M {
	if value == nil {
		return make(M)
	}

	if v, ok := value.(M); ok {
		return maps.Clone(v)
	}

	switch rv := reflect.ValueOf(value); rv.Kind() {
	case reflect.Map:
		out := make(M, rv.Len())

		for iter := rv.MapRange(); iter.Next(); {
			out[ToString(iter.Key().Interface())] = fn(iter.Value().Interface())
		}

		return out
	case reflect.Slice, reflect.Array:
		out := make(M, rv.Len())

		for i := range rv.Len() {
			out[ToString(i)] = fn(rv.Index(i).Interface())
		}

		return out
	}

	return M{"0": fn(value)}
}
