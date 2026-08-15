package convert

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/JFAexe/tem/pkg/reflection"
)

const (
	MinInt32ToFloat32 = -(1 << 23)
	MaxInt32ToFloat32 = 1<<23 - 1
	MinInt64ToFloat64 = -(1 << 53)
	MaxInt64ToFloat64 = 1<<53 - 1
)

var timeLayouts = []string{
	"15:04:05 02-01-2006",
	"2006-01-02T15:04:05",
	time.DateTime,
	time.RFC3339,
	time.RFC3339Nano,
	time.ANSIC,
	time.UnixDate,
}

type (
	ConvertFunc[T any]           = func(any) T
	ConvertKeyFunc[K comparable] = func(any) K
)

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Float interface {
	~float32 | ~float64
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

func SafeFloatToInt32[T Float](f T) int32 {
	f64 := float64(f)

	if math.IsNaN(f64) || math.IsInf(f64, 0) {
		return 0
	}

	return int32(Clamp(math.Trunc(f64), float64(math.MinInt32), float64(math.MaxInt32)))
}

func SafeFloatToInt64[T Float](f T) int64 {
	f64 := float64(f)

	if math.IsNaN(f64) || math.IsInf(f64, 0) {
		return 0
	}

	return int64(Clamp(math.Trunc(f64), float64(math.MinInt64), float64(math.MaxInt64)))
}

func ToAny(v any) any {
	return v
}

func ToBool(value any) bool {
	if value == nil {
		return false
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return false
	}

	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		b, _ := strconv.ParseBool(rv.String())
		return b
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
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
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	case fmt.GoStringer:
		return v.GoString()
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return ""
	}

	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		if target := reflect.TypeFor[string](); rv.Type().ConvertibleTo(target) {
			return rv.Convert(target).String()
		}
	}

	switch rv.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	}

	return fmt.Sprint(value)
}

func ToRune(value any) rune {
	if value == nil {
		return 0
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return SafeInt32(rv.Int())
	case reflect.Float32, reflect.Float64:
		return SafeFloatToInt32(rv.Float())
	case reflect.String:
		if s := rv.String(); s != "" {
			r, _ := utf8.DecodeRuneInString(s)

			return r
		}
	}

	return ToRune(ToString(value))
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
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}

		if f, err := v.Float64(); err == nil {
			return int64(f)
		}
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

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Float32, reflect.Float64:
		return SafeFloatToInt64(rv.Float())
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
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	case bool:
		if v {
			return 1
		}
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return SafeFloat64(rv.Int())
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
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}

		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return time.Duration(i)
		}
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Duration(rv.Int())
	case reflect.Float32, reflect.Float64:
		return time.Duration(rv.Float() * float64(time.Second))
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
	case time.Duration:
		return time.Unix(0, int64(v))
	case string:
		for _, layout := range timeLayouts {
			if t, err := time.Parse(layout, v); err == nil {
				return t
			}
		}

		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return time.Unix(i, 0)
		}
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return time.Time{}
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Unix(rv.Int(), 0)
	case reflect.Float32, reflect.Float64:
		var (
			f = rv.Float()
			s = int64(f)
		)

		return time.Unix(s, int64((f-float64(s))*1e9))
	}

	return ToTime(ToString(value))
}

func ToSlice[T any, S []T](value any, fn ConvertFunc[T]) S {
	if value == nil {
		return make(S, 0)
	}

	if v, ok := value.(S); ok {
		return slices.Clone(v)
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return make(S, 0)
	}

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return S{fn(value)}
	}

	var (
		typ = reflect.TypeFor[T]()
		out = make(S, 0, rv.Len())
	)

	for i := range rv.Len() {
		var (
			v T

			e = rv.Index(i)
		)

		if e.IsValid() && e.Type().ConvertibleTo(typ) {
			v = e.Convert(typ).Interface().(T)
		} else {
			v = fn(e.Interface())
		}

		out = append(out, v)
	}

	return out
}

func ToAnySlice(value any) []any {
	return ToSlice(value, ToAny)
}

func ToBoolSlice(value any) []bool {
	return ToSlice(value, ToBool)
}

func ToStringSlice(value any) []string {
	return ToSlice(value, ToString)
}

func ToIntSlice(value any) []int {
	return ToSlice(value, ToInt)
}

func ToInt32Slice(value any) []int32 {
	return ToSlice(value, ToInt32)
}

func ToInt64Slice(value any) []int64 {
	return ToSlice(value, ToInt64)
}

func ToFloat32Slice(value any) []float32 {
	return ToSlice(value, ToFloat32)
}

func ToFloat64Slice(value any) []float64 {
	return ToSlice(value, ToFloat64)
}

func ToDurationSlice(value any) []time.Duration {
	return ToSlice(value, ToDuration)
}

func ToTimeSlice(value any) []time.Time {
	return ToSlice(value, ToTime)
}

func ToRuneSlice(value any) []rune {
	switch v := value.(type) {
	case []rune:
		return v
	case string:
		return []rune(v)
	case nil:
		return nil
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return nil
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []rune{SafeInt32(rv.Int())}
	case reflect.Float32, reflect.Float64:
		return []rune{SafeFloatToInt32(rv.Float())}
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		out := make([]rune, 0, rv.Len())

		for i := range rv.Len() {
			e := rv.Index(i)

			if !e.IsValid() {
				continue
			}

			switch e.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				out = append(out, SafeInt32(e.Int()))
			case reflect.Float32, reflect.Float64:
				out = append(out, int32(math.Trunc(e.Float())))
			default:
				out = append(out, ToRune(e.Interface()))
			}
		}

		return out
	}

	return []rune(ToString(value))
}

func ToByteSlice(value any) []byte {
	if value == nil {
		return nil
	}

	if v, ok := value.([]byte); ok {
		return v
	}

	return []byte(ToString(value))
}

func ToMap[K comparable, V any, M map[K]V](value any, kf ConvertKeyFunc[K], vf ConvertFunc[V]) M {
	if value == nil {
		return make(M)
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return make(M)
	}

	switch rv.Kind() {
	case reflect.Map:
		out := make(M, rv.Len())

		for iter := rv.MapRange(); iter.Next(); {
			out[kf(iter.Key().Interface())] = vf(iter.Value().Interface())
		}

		return out
	case reflect.Slice, reflect.Array:
		out := make(M, rv.Len())

		for i := range rv.Len() {
			out[kf(i)] = vf(rv.Index(i).Interface())
		}

		return out
	case reflect.Struct:
		out := make(M, rv.NumField())

		for name, field := range reflection.ExportedFields(rv) {
			out[kf(name)] = vf(field.Interface())
		}

		return out
	}

	return M{kf(0): vf(value)}
}

func ToAnyMap(value any) map[string]any {
	return ToMap(value, ToString, ToAny)
}

func ToBoolMap(value any) map[string]bool {
	return ToMap(value, ToString, ToBool)
}

func ToStringMap(value any) map[string]string {
	return ToMap(value, ToString, ToString)
}

func ToIntMap(value any) map[string]int {
	return ToMap(value, ToString, ToInt)
}

func ToInt32Map(value any) map[string]int32 {
	return ToMap(value, ToString, ToInt32)
}

func ToInt64Map(value any) map[string]int64 {
	return ToMap(value, ToString, ToInt64)
}

func ToFloat32Map(value any) map[string]float32 {
	return ToMap(value, ToString, ToFloat32)
}

func ToFloat64Map(value any) map[string]float64 {
	return ToMap(value, ToString, ToFloat64)
}

func ToDurationMap(value any) map[string]time.Duration {
	return ToMap(value, ToString, ToDuration)
}

func ToTimeMap(value any) map[string]time.Time {
	return ToMap(value, ToString, ToTime)
}
