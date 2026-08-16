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
	Float32ExactIntMax = 1 << 24
	Float64ExactIntMax = 1 << 53
)

var timeLayouts = []string{
	time.DateTime,
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
}

type (
	ConvertFunc[T any]           func(any) T
	ConvertKeyFunc[K comparable] func(any) K
)

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Float interface {
	~float32 | ~float64
}

func Clamp[T cmp.Ordered](v, mi, ma T) T {
	return min(max(v, mi), ma)
}

func ClampFloat[F Float, I Int | Uint](v F, mi, ma I) I {
	f := float64(v)

	if f <= float64(mi) {
		return mi
	}

	if f >= float64(ma) {
		return ma
	}

	return I(math.Trunc(f))
}

func SafeInt[T Int](v T) int {
	return int(Clamp(int64(v), math.MinInt, math.MaxInt))
}

func SafeInt8[T Int](v T) int8 {
	return int8(Clamp(int64(v), math.MinInt8, math.MaxInt8))
}

func SafeInt16[T Int](v T) int16 {
	return int16(Clamp(int64(v), math.MinInt16, math.MaxInt16))
}

func SafeInt32[T Int](v T) int32 {
	return int32(Clamp(int64(v), math.MinInt32, math.MaxInt32))
}

func SafeIntToUint8[T Int](v T) uint8 {
	return uint8(Clamp(int64(v), 0, math.MaxUint8))
}

func SafeIntToUint16[T Int](v T) uint16 {
	return uint16(Clamp(int64(v), 0, math.MaxUint16))
}

func SafeIntToUint32[T Int](v T) uint32 {
	return uint32(Clamp(int64(v), 0, math.MaxUint32))
}

func SafeIntToUint64[T Int](v T) uint64 {
	return uint64(max(v, 0))
}

func SafeIntToFloat32[T Int](v T) float32 {
	return float32(Clamp(int64(v), -Float32ExactIntMax, Float32ExactIntMax))
}

func SafeIntToFloat64[T Int](v T) float64 {
	return float64(Clamp(int64(v), -Float64ExactIntMax, Float64ExactIntMax))
}

func SafeUint[T Uint](v T) uint {
	return uint(Clamp(uint64(v), 0, math.MaxUint))
}

func SafeUint8[T Uint](v T) uint8 {
	return uint8(Clamp(uint64(v), 0, math.MaxUint8))
}

func SafeUint16[T Uint](v T) uint16 {
	return uint16(Clamp(uint64(v), 0, math.MaxUint16))
}

func SafeUint32[T Uint](v T) uint32 {
	return uint32(Clamp(uint64(v), 0, math.MaxUint32))
}

func SafeUintToInt8[T Uint](v T) int8 {
	return int8(Clamp(uint64(v), 0, math.MaxInt8))
}

func SafeUintToInt16[T Uint](v T) int16 {
	return int16(Clamp(uint64(v), 0, math.MaxInt16))
}

func SafeUintToInt32[T Uint](v T) int32 {
	return int32(Clamp(uint64(v), 0, math.MaxInt32))
}

func SafeUintToInt64[T Uint](v T) int64 {
	return int64(Clamp(uint64(v), 0, math.MaxInt64))
}

func SafeUintToFloat32[T Uint](v T) float32 {
	return float32(Clamp(uint64(v), 0, Float32ExactIntMax))
}

func SafeUintToFloat64[T Uint](v T) float64 {
	return float64(Clamp(uint64(v), 0, Float64ExactIntMax))
}

func SafeFloat32[T Float](v T) float32 {
	return float32(Clamp(SafeFloat64(v), -math.MaxFloat32, math.MaxFloat32))
}

func SafeFloat64[T Float](v T) float64 {
	f := float64(v)

	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}

	return f
}

func SafeFloatToInt8[T Float](v T) int8 {
	return ClampFloat[float64, int8](SafeFloat64(v), math.MinInt8, math.MaxInt8)
}

func SafeFloatToInt16[T Float](v T) int16 {
	return ClampFloat[float64, int16](SafeFloat64(v), math.MinInt16, math.MaxInt16)
}

func SafeFloatToInt32[T Float](v T) int32 {
	return ClampFloat[float64, int32](SafeFloat64(v), math.MinInt32, math.MaxInt32)
}

func SafeFloatToInt64[T Float](v T) int64 {
	return ClampFloat[float64, int64](SafeFloat64(v), math.MinInt64, math.MaxInt64)
}

func SafeFloatToUint8[T Float](v T) uint8 {
	return ClampFloat[float64, uint8](SafeFloat64(v), 0, math.MaxUint8)
}

func SafeFloatToUint16[T Float](v T) uint16 {
	return ClampFloat[float64, uint16](SafeFloat64(v), 0, math.MaxUint16)
}

func SafeFloatToUint32[T Float](v T) uint32 {
	return ClampFloat[float64, uint32](SafeFloat64(v), 0, math.MaxUint32)
}

func SafeFloatToUint64[T Float](v T) uint64 {
	return ClampFloat[float64, uint64](SafeFloat64(v), 0, math.MaxUint64)
}

func ToAny(v any) any {
	return v
}

func ToBool(value any) bool {
	if value == nil {
		return false
	}

	if v, ok := value.(bool); ok {
		return v
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
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
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
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 32)
	case reflect.Float64:
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
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return SafeUintToInt32(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return SafeFloatToInt32(rv.Float())
	case reflect.String:
		if s := rv.String(); s != "" {
			r, _ := utf8.DecodeRuneInString(s)

			return r
		}

		return 0
	}

	return ToRune(ToString(value))
}

func ToInt(value any) int {
	return SafeInt(ToInt64(value))
}

func ToInt8(value any) int8 {
	return SafeInt8(ToInt64(value))
}

func ToInt16(value any) int16 {
	return SafeInt16(ToInt64(value))
}

func ToInt32(value any) int32 {
	return SafeInt32(ToInt64(value))
}

func ToInt64(value any) int64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case bool:
		if v {
			return 1
		}

		return 0
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}

		if f, err := v.Float64(); err == nil {
			return SafeFloatToInt64(f)
		}
	case string:
		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return i
		}

		if u, err := strconv.ParseUint(v, 0, 64); err == nil {
			return SafeUintToInt64(u)
		}

		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return SafeFloatToInt64(f)
		}
	case time.Time:
		return v.UTC().Unix()
	case time.Duration:
		return int64(v)
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return SafeUintToInt64(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return SafeFloatToInt64(rv.Float())
	}

	s := ToString(value)

	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return i
	}

	if u, err := strconv.ParseUint(s, 0, 64); err == nil {
		return SafeUintToInt64(u)
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return SafeFloatToInt64(f)
	}

	return 0
}

func ToUint(value any) uint {
	return SafeUint(ToUint64(value))
}

func ToUint8(value any) uint8 {
	return SafeUint8(ToUint64(value))
}

func ToUint16(value any) uint16 {
	return SafeUint16(ToUint64(value))
}

func ToUint32(value any) uint32 {
	return SafeUint32(ToUint64(value))
}

func ToUint64(value any) uint64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case bool:
		if v {
			return 1
		}

		return 0
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return SafeIntToUint64(i)
		}

		if f, err := v.Float64(); err == nil {
			return SafeFloatToUint64(f)
		}
	case string:
		if u, err := strconv.ParseUint(v, 0, 64); err == nil {
			return u
		}

		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return SafeIntToUint64(i)
		}

		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return SafeFloatToUint64(f)
		}
	case time.Time:
		return SafeIntToUint64(v.UTC().Unix())
	case time.Duration:
		return SafeIntToUint64(v)
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return SafeIntToUint64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint()
	case reflect.Float32, reflect.Float64:
		return SafeFloatToUint64(rv.Float())
	}

	s := ToString(value)

	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return SafeIntToUint64(i)
	}

	if u, err := strconv.ParseUint(s, 0, 64); err == nil {
		return u
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return SafeFloatToUint64(f)
	}

	return 0
}

func ToFloat32(value any) float32 {
	return SafeFloat32(ToFloat64(value))
}

func ToFloat64(value any) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case bool:
		if v {
			return 1
		}

		return 0
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}

		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return SafeIntToFloat64(i)
		}

		if u, err := strconv.ParseUint(v, 0, 64); err == nil {
			return SafeUintToFloat64(u)
		}
	case time.Time:
		return SafeIntToFloat64(v.UTC().Unix())
	case time.Duration:
		return SafeIntToFloat64(v)
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return SafeFloat64(rv.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return SafeIntToFloat64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return SafeUintToFloat64(rv.Uint())
	}

	s := ToString(value)

	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return SafeIntToFloat64(i)
	}

	if u, err := strconv.ParseUint(s, 0, 64); err == nil {
		return SafeUintToFloat64(u)
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
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
	case time.Time:
		return v.UTC().Sub(time.Unix(0, 0).UTC())
	case bool:
		if v {
			return time.Second
		}

		return 0
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return time.Duration(i)
		}

		if f, err := v.Float64(); err == nil {
			return time.Duration(SafeFloatToInt64(f * float64(time.Second)))
		}
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}

		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return time.Duration(i)
		}

		if u, err := strconv.ParseUint(v, 0, 64); err == nil {
			return time.Duration(SafeUintToInt64(u))
		}

		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Duration(SafeFloatToInt64(f * float64(time.Second)))
		}

		return 0
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Duration(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return time.Duration(SafeUintToInt64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		return time.Duration(SafeFloatToInt64(rv.Float() * float64(time.Second)))
	}

	return time.Duration(ToInt64(value))
}

func ToTime(value any) time.Time {
	if value == nil {
		return time.Time{}
	}

	switch v := value.(type) {
	case time.Time:
		return v.UTC()
	case time.Duration:
		return time.Unix(0, int64(v)).UTC()
	case bool:
		if v {
			return time.Unix(1, 0).UTC()
		}

		return time.Time{}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return time.Unix(i, 0).UTC()
		}

		if f, err := v.Float64(); err == nil {
			s := SafeFloatToInt64(f)

			return time.Unix(s, unixNsec(s, f)).UTC()
		}
	case string:
		for _, layout := range timeLayouts {
			if t, err := time.Parse(layout, v); err == nil {
				return t.UTC()
			}
		}

		if i, err := strconv.ParseInt(v, 0, 64); err == nil {
			return time.Unix(i, 0).UTC()
		}

		if u, err := strconv.ParseUint(v, 0, 64); err == nil {
			return time.Unix(SafeUintToInt64(u), 0).UTC()
		}

		if f, err := strconv.ParseFloat(v, 64); err == nil {
			s := SafeFloatToInt64(f)

			return time.Unix(s, unixNsec(s, f)).UTC()
		}

		return time.Time{}
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return time.Time{}
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Unix(rv.Int(), 0).UTC()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return time.Unix(SafeUintToInt64(rv.Uint()), 0).UTC()
	case reflect.Float32, reflect.Float64:
		var (
			f = rv.Float()
			s = SafeFloatToInt64(f)
		)

		return time.Unix(s, unixNsec(s, f)).UTC()
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

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
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
	case reflect.Map:
		out := make(S, 0, rv.Len())

		for iter := rv.MapRange(); iter.Next(); {
			out = append(out, fn(iter.Value().Interface()))
		}

		return out
	case reflect.Struct:
		out := make(S, 0, rv.NumField())

		for _, field := range reflection.ExportedFields(rv) {
			out = append(out, fn(field.Interface()))
		}

		return out
	}

	return S{fn(value)}
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

func ToInt8Slice(value any) []int8 {
	return ToSlice(value, ToInt8)
}

func ToInt16Slice(value any) []int16 {
	return ToSlice(value, ToInt16)
}

func ToInt32Slice(value any) []int32 {
	return ToSlice(value, ToInt32)
}

func ToInt64Slice(value any) []int64 {
	return ToSlice(value, ToInt64)
}

func ToUintSlice(value any) []uint {
	return ToSlice(value, ToUint)
}

func ToUint8Slice(value any) []uint8 {
	return ToSlice(value, ToUint8)
}

func ToUint16Slice(value any) []uint16 {
	return ToSlice(value, ToUint16)
}

func ToUint32Slice(value any) []uint32 {
	return ToSlice(value, ToUint32)
}

func ToUint64Slice(value any) []uint64 {
	return ToSlice(value, ToUint64)
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
	case nil:
		return make([]rune, 0)
	case []rune:
		return slices.Clone(v)
	case string:
		return []rune(v)
	}

	rv, err := reflection.IndirectValue(reflect.ValueOf(value))
	if err != nil || !rv.IsValid() {
		return make([]rune, 0)
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []rune{SafeInt32(rv.Int())}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []rune{SafeUintToInt32(rv.Uint())}
	case reflect.Float32, reflect.Float64:
		return []rune{SafeFloatToInt32(rv.Float())}
	case reflect.Slice, reflect.Array:
		out := make([]rune, 0, rv.Len())

		for i := range rv.Len() {
			e := rv.Index(i)

			if !e.IsValid() {
				continue
			}

			switch e.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				out = append(out, SafeInt32(e.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				out = append(out, SafeUintToInt32(e.Uint()))
			case reflect.Float32, reflect.Float64:
				out = append(out, SafeFloatToInt32(e.Float()))
			default:
				out = append(out, ToRune(e.Interface()))
			}
		}

		return out
	}

	return []rune(ToString(value))
}

func ToByteSlice(value any) []byte {
	switch v := value.(type) {
	case nil:
		return make([]byte, 0)
	case []byte:
		return slices.Clone(v)
	case string:
		return []byte(v)
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

func ToInt8Map(value any) map[string]int8 {
	return ToMap(value, ToString, ToInt8)
}

func ToInt16Map(value any) map[string]int16 {
	return ToMap(value, ToString, ToInt16)
}

func ToInt32Map(value any) map[string]int32 {
	return ToMap(value, ToString, ToInt32)
}

func ToInt64Map(value any) map[string]int64 {
	return ToMap(value, ToString, ToInt64)
}

func ToUintMap(value any) map[string]uint {
	return ToMap(value, ToString, ToUint)
}

func ToUint8Map(value any) map[string]uint8 {
	return ToMap(value, ToString, ToUint8)
}

func ToUint16Map(value any) map[string]uint16 {
	return ToMap(value, ToString, ToUint16)
}

func ToUint32Map(value any) map[string]uint32 {
	return ToMap(value, ToString, ToUint32)
}

func ToUint64Map(value any) map[string]uint64 {
	return ToMap(value, ToString, ToUint64)
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

func unixNsec(s int64, f float64) int64 {
	if f >= float64(math.MaxInt64) || f <= float64(math.MinInt64) {
		return 0
	}

	nsec := int64((f - float64(s)) * 1e9)

	if (s == math.MaxInt64 && nsec > 0) || (s == math.MinInt64 && nsec < 0) {
		return 0
	}

	return nsec
}
