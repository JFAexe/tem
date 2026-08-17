package convert

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/JFAexe/tem/pkg/reflection"
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
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	case string:
		return boolFromString(v)
	case []byte:
		return boolFromString(string(v))
	case []rune:
		return boolFromString(string(v))
	case time.Duration:
		return v != 0
	case time.Time:
		return !v.UTC().IsZero()
	case error:
		return boolFromString(v.Error())
	case fmt.Stringer:
		return boolFromString(v.String())
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
		return false
	}

	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		return boolFromString(rv.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	}

	return boolFromString(ToString(value))
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
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
		return ""
	}

	switch rv.Kind() {
	case reflect.String:
		return rv.String()
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
	case reflect.Slice, reflect.Array:
		if rv.Type().ConvertibleTo(reflect.TypeFor[string]()) {
			return rv.Convert(reflect.TypeFor[string]()).String()
		}
	}

	return fmt.Sprint(value)
}

func ToRune(value any) rune {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return SafeInt32(int64(v))
	case int8:
		return SafeInt32(int64(v))
	case int16:
		return SafeInt32(int64(v))
	case int32:
		return SafeInt32(int64(v))
	case int64:
		return SafeInt32(v)
	case uint:
		return SafeUintToInt32(uint64(v))
	case uint8:
		return SafeUintToInt32(uint64(v))
	case uint16:
		return SafeUintToInt32(uint64(v))
	case uint32:
		return SafeUintToInt32(uint64(v))
	case uint64:
		return SafeUintToInt32(v)
	case float32:
		return SafeFloatToInt32(v)
	case float64:
		return SafeFloatToInt32(v)
	case string:
		return runeFromString(v)
	case []byte:
		if len(v) > 0 {
			r, _ := utf8.DecodeRune(v)

			return r
		}

		return 0
	case []rune:
		if len(v) > 0 {
			return v[0]
		}

		return 0
	case error:
		return ToRune(v.Error())
	case fmt.Stringer:
		return ToRune(v.String())
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
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
		return runeFromString(rv.String())
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
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return SafeUintToInt64(v)
	case uint8:
		return SafeUintToInt64(v)
	case uint16:
		return SafeUintToInt64(v)
	case uint32:
		return SafeUintToInt64(v)
	case uint64:
		return SafeUintToInt64(v)
	case float32:
		return SafeFloatToInt64(v)
	case float64:
		return SafeFloatToInt64(v)
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
		return int64FromString(v)
	case []byte:
		return int64FromString(string(v))
	case []rune:
		return int64FromString(string(v))
	case time.Duration:
		return int64(v)
	case time.Time:
		return v.UTC().Unix()
	case error:
		return ToInt64(v.Error())
	case fmt.Stringer:
		return ToInt64(v.String())
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return SafeUintToInt64(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return SafeFloatToInt64(rv.Float())
	case reflect.String:
		return int64FromString(rv.String())
	}

	return int64FromString(ToString(value))
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
	case int:
		return SafeIntToUint64(v)
	case int8:
		return SafeIntToUint64(v)
	case int16:
		return SafeIntToUint64(v)
	case int32:
		return SafeIntToUint64(v)
	case int64:
		return SafeIntToUint64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case float32:
		return SafeFloatToUint64(v)
	case float64:
		return SafeFloatToUint64(v)
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
		return uint64FromString(v)
	case []byte:
		return uint64FromString(string(v))
	case []rune:
		return uint64FromString(string(v))
	case time.Duration:
		return SafeIntToUint64(v)
	case time.Time:
		return SafeIntToUint64(v.UTC().Unix())
	case error:
		return ToUint64(v.Error())
	case fmt.Stringer:
		return ToUint64(v.String())
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return SafeIntToUint64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint()
	case reflect.Float32, reflect.Float64:
		return SafeFloatToUint64(rv.Float())
	case reflect.String:
		return uint64FromString(rv.String())
	}

	return uint64FromString(ToString(value))
}

func ToFloat32(value any) float32 {
	return SafeFloat32(ToFloat64(value))
}

func ToFloat64(value any) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return SafeIntToFloat64(v)
	case int8:
		return SafeIntToFloat64(v)
	case int16:
		return SafeIntToFloat64(v)
	case int32:
		return SafeIntToFloat64(v)
	case int64:
		return SafeIntToFloat64(v)
	case uint:
		return SafeUintToFloat64(v)
	case uint8:
		return SafeUintToFloat64(v)
	case uint16:
		return SafeUintToFloat64(v)
	case uint32:
		return SafeUintToFloat64(v)
	case uint64:
		return SafeUintToFloat64(v)
	case float32:
		return SafeFloat64(v)
	case float64:
		return SafeFloat64(v)
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
		return float64FromString(v)
	case []byte:
		return float64FromString(string(v))
	case []rune:
		return float64FromString(string(v))
	case time.Duration:
		return SafeIntToFloat64(v)
	case time.Time:
		return SafeIntToFloat64(v.UTC().Unix())
	case error:
		return ToFloat64(v.Error())
	case fmt.Stringer:
		return ToFloat64(v.String())
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return SafeFloat64(rv.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return SafeIntToFloat64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return SafeUintToFloat64(rv.Uint())
	case reflect.String:
		return float64FromString(rv.String())
	}

	return float64FromString(ToString(value))
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
	case int:
		return time.Duration(v)
	case int8:
		return time.Duration(v)
	case int16:
		return time.Duration(v)
	case int32:
		return time.Duration(v)
	case int64:
		return time.Duration(v)
	case uint:
		return time.Duration(SafeUintToInt64(v))
	case uint8:
		return time.Duration(SafeUintToInt64(v))
	case uint16:
		return time.Duration(SafeUintToInt64(v))
	case uint32:
		return time.Duration(SafeUintToInt64(v))
	case uint64:
		return time.Duration(SafeUintToInt64(v))
	case float32:
		return durationFromFloat(SafeFloat64(v))
	case float64:
		return durationFromFloat(SafeFloat64(v))
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
			return durationFromFloat(f)
		}
	case string:
		return durationFromString(v)
	case []byte:
		return durationFromString(string(v))
	case []rune:
		return durationFromString(string(v))
	case error:
		return ToDuration(v.Error())
	case fmt.Stringer:
		return ToDuration(v.String())
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
		return 0
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Duration(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return time.Duration(SafeUintToInt64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		return durationFromFloat(rv.Float())
	}

	return durationFromString(ToString(value))
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
	case int:
		return time.Unix(int64(v), 0).UTC()
	case int8:
		return time.Unix(int64(v), 0).UTC()
	case int16:
		return time.Unix(int64(v), 0).UTC()
	case int32:
		return time.Unix(int64(v), 0).UTC()
	case int64:
		return time.Unix(v, 0).UTC()
	case uint:
		return time.Unix(SafeUintToInt64(v), 0).UTC()
	case uint8:
		return time.Unix(SafeUintToInt64(v), 0).UTC()
	case uint16:
		return time.Unix(SafeUintToInt64(v), 0).UTC()
	case uint32:
		return time.Unix(SafeUintToInt64(v), 0).UTC()
	case uint64:
		return time.Unix(SafeUintToInt64(v), 0).UTC()
	case float32:
		return unixTimeFromFloat(SafeFloat64(v))
	case float64:
		return unixTimeFromFloat(SafeFloat64(v))
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
			return unixTimeFromFloat(f)
		}
	case string:
		return timeFromString(v)
	case []byte:
		return timeFromString(string(v))
	case []rune:
		return timeFromString(string(v))
	case error:
		return ToTime(v.Error())
	case fmt.Stringer:
		return ToTime(v.String())
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
		return time.Time{}
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Unix(rv.Int(), 0).UTC()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return time.Unix(SafeUintToInt64(rv.Uint()), 0).UTC()
	case reflect.Float32, reflect.Float64:
		return unixTimeFromFloat(rv.Float())
	}

	return timeFromString(ToString(value))
}

func boolFromString(s string) bool {
	b, _ := strconv.ParseBool(s)

	return b
}

func runeFromString(s string) rune {
	if s != "" {
		r, _ := utf8.DecodeRuneInString(s)

		return r
	}

	return 0
}

func int64FromString(s string) int64 {
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

func uint64FromString(s string) uint64 {
	if u, err := strconv.ParseUint(s, 0, 64); err == nil {
		return u
	}

	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return SafeIntToUint64(i)
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return SafeFloatToUint64(f)
	}

	return 0
}

func float64FromString(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return SafeIntToFloat64(i)
	}

	if u, err := strconv.ParseUint(s, 0, 64); err == nil {
		return SafeUintToFloat64(u)
	}

	return 0
}

func durationFromFloat(f float64) time.Duration {
	return time.Duration(SafeFloatToInt64(f * float64(time.Second)))
}

func durationFromString(s string) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}

	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return time.Duration(i)
	}

	if u, err := strconv.ParseUint(s, 0, 64); err == nil {
		return time.Duration(SafeUintToInt64(u))
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return durationFromFloat(f)
	}

	return 0
}

func timeFromString(s string) time.Time {
	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}

	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return time.Unix(i, 0).UTC()
	}

	if u, err := strconv.ParseUint(s, 0, 64); err == nil {
		return time.Unix(SafeUintToInt64(u), 0).UTC()
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return unixTimeFromFloat(f)
	}

	return time.Time{}
}

func unixTimeFromFloat(f float64) time.Time {
	s := SafeFloatToInt64(f)

	return time.Unix(s, unixNsec(s, f)).UTC()
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
