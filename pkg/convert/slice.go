package convert

import (
	"reflect"
	"slices"
	"strconv"
	"time"
	"uuid"

	"github.com/JFAexe/tem/pkg/reflection"
)

func ToSlice[T any, S []T](value any, fn ConvertFunc[T]) S {
	if value == nil {
		return make(S, 0)
	}

	if v, ok := value.(S); ok {
		if v == nil {
			return make(S, 0)
		}

		return slices.Clone(v)
	}

	rv := reflection.IndirectValue(value)
	if !rv.IsValid() {
		return make(S, 0)
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		var (
			out = make(S, 0, rv.Len())
			typ = reflect.TypeFor[T]()
		)

		for i := range rv.Len() {
			e := rv.Index(i)

			if e.Type() == typ {
				out = append(out, e.Interface().(T))
			} else {
				out = append(out, fn(e.Interface()))
			}
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

func ToUUIDSlice(value any) []uuid.UUID {
	return ToSlice(value, ToUUID)
}

func ToRuneSlice(value any) []rune {
	switch v := value.(type) {
	case nil:
		return make([]rune, 0)
	case []rune:
		return slices.Clone(v)
	case string:
		return []rune(v)
	case []byte:
		return []rune(string(v))
	case int:
		return []rune{SafeIntToRune(v)}
	case int8:
		return []rune{SafeIntToRune(v)}
	case int16:
		return []rune{SafeIntToRune(v)}
	case int32:
		return []rune{SafeIntToRune(v)}
	case int64:
		return []rune{SafeIntToRune(v)}
	case uint:
		return []rune{SafeUintToRune(v)}
	case uint8:
		return []rune{SafeUintToRune(v)}
	case uint16:
		return []rune{SafeUintToRune(v)}
	case uint32:
		return []rune{SafeUintToRune(v)}
	case uint64:
		return []rune{SafeUintToRune(v)}
	case float32:
		return []rune{SafeFloatToRune(v)}
	case float64:
		return []rune{SafeFloatToRune(v)}
	}

	rv := reflection.IndirectValue(value)
	if !rv.IsValid() {
		return make([]rune, 0)
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []rune{SafeIntToRune(rv.Int())}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []rune{SafeUintToRune(rv.Uint())}
	case reflect.Float32, reflect.Float64:
		return []rune{SafeFloatToRune(rv.Float())}
	case reflect.Slice, reflect.Array:
		out := make([]rune, 0, rv.Len())

		for i := range rv.Len() {
			e := reflect.Indirect(rv.Index(i))

			if !e.IsValid() {
				continue
			}

			switch e.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				out = append(out, SafeIntToRune(e.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				out = append(out, SafeUintToRune(e.Uint()))
			case reflect.Float32, reflect.Float64:
				out = append(out, SafeFloatToRune(e.Float()))
			case reflect.String:
				out = append(out, runeFromString(e.String()))
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
	case []rune:
		return []byte(string(v))
	case int:
		return []byte(strconv.FormatInt(int64(v), 10))
	case int8:
		return []byte(strconv.FormatInt(int64(v), 10))
	case int16:
		return []byte(strconv.FormatInt(int64(v), 10))
	case int32:
		return []byte(strconv.FormatInt(int64(v), 10))
	case int64:
		return []byte(strconv.FormatInt(v, 10))
	case uint:
		return []byte(strconv.FormatUint(uint64(v), 10))
	case uint8:
		return []byte(strconv.FormatUint(uint64(v), 10))
	case uint16:
		return []byte(strconv.FormatUint(uint64(v), 10))
	case uint32:
		return []byte(strconv.FormatUint(uint64(v), 10))
	case uint64:
		return []byte(strconv.FormatUint(v, 10))
	case float32:
		return []byte(strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		return []byte(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		return []byte(strconv.FormatBool(v))
	case uuid.UUID:
		return v[:]
	}

	return []byte(ToString(value))
}
