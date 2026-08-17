package convert

import (
	"reflect"
	"time"

	"github.com/JFAexe/tem/pkg/reflection"
)

func ToMap[K comparable, V any, M map[K]V](value any, kf ConvertKeyFunc[K], vf ConvertFunc[V]) M {
	if value == nil {
		return make(M)
	}

	rv := reflection.IndirectValue(reflect.ValueOf(value))
	if !rv.IsValid() {
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

func ToStringAnyMap(value any) map[string]any {
	return ToMap(value, ToString, ToAny)
}

func ToStringBoolMap(value any) map[string]bool {
	return ToMap(value, ToString, ToBool)
}

func ToStringStringMap(value any) map[string]string {
	return ToMap(value, ToString, ToString)
}

func ToStringIntMap(value any) map[string]int {
	return ToMap(value, ToString, ToInt)
}

func ToStringInt8Map(value any) map[string]int8 {
	return ToMap(value, ToString, ToInt8)
}

func ToStringInt16Map(value any) map[string]int16 {
	return ToMap(value, ToString, ToInt16)
}

func ToStringInt32Map(value any) map[string]int32 {
	return ToMap(value, ToString, ToInt32)
}

func ToStringInt64Map(value any) map[string]int64 {
	return ToMap(value, ToString, ToInt64)
}

func ToStringUintMap(value any) map[string]uint {
	return ToMap(value, ToString, ToUint)
}

func ToStringUint8Map(value any) map[string]uint8 {
	return ToMap(value, ToString, ToUint8)
}

func ToStringUint16Map(value any) map[string]uint16 {
	return ToMap(value, ToString, ToUint16)
}

func ToStringUint32Map(value any) map[string]uint32 {
	return ToMap(value, ToString, ToUint32)
}

func ToStringUint64Map(value any) map[string]uint64 {
	return ToMap(value, ToString, ToUint64)
}

func ToStringFloat32Map(value any) map[string]float32 {
	return ToMap(value, ToString, ToFloat32)
}

func ToStringFloat64Map(value any) map[string]float64 {
	return ToMap(value, ToString, ToFloat64)
}

func ToStringDurationMap(value any) map[string]time.Duration {
	return ToMap(value, ToString, ToDuration)
}

func ToStringTimeMap(value any) map[string]time.Time {
	return ToMap(value, ToString, ToTime)
}
