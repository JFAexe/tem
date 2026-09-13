package convert

import (
	"fmt"
	"net"
	"reflect"
	"time"
	"uuid"

	"github.com/JFAexe/tem/pkg/reflection"
)

func ToMap[K comparable, V any, M map[K]V](value any, kf ConvertKeyFunc[K], vf ConvertFunc[V]) M {
	if value == nil {
		return make(M)
	}

	rv := reflection.IndirectValue(value)
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

		for name, field := range reflection.Fields(rv) {
			out[kf(name)] = vf(field.Interface())
		}

		return out
	}

	return M{kf(0): vf(value)}
}

func ToMapDynamic(value any, kf, vf func(any) any, kt, vt reflect.Type) (any, error) {
	if !kt.Comparable() {
		return nil, fmt.Errorf("key type %s not comparable", kt)
	}

	rv := reflection.IndirectValue(value)
	if !rv.IsValid() {
		return reflect.MakeMap(reflect.MapOf(kt, vt)).Interface(), nil
	}

	var (
		out = reflect.MakeMap(reflect.MapOf(kt, vt))
		set = func(k any, v any) error {
			kv := reflect.ValueOf(kf(k))
			if !kv.IsValid() {
				kv = reflect.Zero(kt)
			}

			vv := reflect.ValueOf(vf(v))
			if !vv.IsValid() {
				vv = reflect.Zero(vt)
			}

			if !kv.Type().AssignableTo(kt) || !vv.Type().AssignableTo(vt) {
				return fmt.Errorf("type mismatch: got %s/%s, want %s/%s", kv.Type(), vv.Type(), kt, vt)
			}

			out.SetMapIndex(kv, vv)

			return nil
		}
	)

	switch rv.Kind() {
	case reflect.Map:
		for it := rv.MapRange(); it.Next(); {
			if err := set(it.Key().Interface(), it.Value().Interface()); err != nil {
				return nil, err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range rv.Len() {
			if err := set(i, rv.Index(i).Interface()); err != nil {
				return nil, err
			}
		}
	case reflect.Struct:
		for name, field := range reflection.Fields(rv) {
			if err := set(name, field.Interface()); err != nil {
				return nil, err
			}
		}
	default:
		if err := set(0, value); err != nil {
			return nil, err
		}
	}

	return out.Interface(), nil
}

func ToAnyMap(value any) map[any]any {
	return ToMap(value, ToAny, ToAny)
}

func ToBoolMap(value any) map[any]bool {
	return ToMap(value, ToAny, ToBool)
}

func ToStringMap(value any) map[any]string {
	return ToMap(value, ToAny, ToString)
}

func ToIntMap(value any) map[any]int {
	return ToMap(value, ToAny, ToInt)
}

func ToInt8Map(value any) map[any]int8 {
	return ToMap(value, ToAny, ToInt8)
}

func ToInt16Map(value any) map[any]int16 {
	return ToMap(value, ToAny, ToInt16)
}

func ToInt32Map(value any) map[any]int32 {
	return ToMap(value, ToAny, ToInt32)
}

func ToInt64Map(value any) map[any]int64 {
	return ToMap(value, ToAny, ToInt64)
}

func ToUintMap(value any) map[any]uint {
	return ToMap(value, ToAny, ToUint)
}

func ToUint8Map(value any) map[any]uint8 {
	return ToMap(value, ToAny, ToUint8)
}

func ToUint16Map(value any) map[any]uint16 {
	return ToMap(value, ToAny, ToUint16)
}

func ToUint32Map(value any) map[any]uint32 {
	return ToMap(value, ToAny, ToUint32)
}

func ToUint64Map(value any) map[any]uint64 {
	return ToMap(value, ToAny, ToUint64)
}

func ToFloat32Map(value any) map[any]float32 {
	return ToMap(value, ToAny, ToFloat32)
}

func ToFloat64Map(value any) map[any]float64 {
	return ToMap(value, ToAny, ToFloat64)
}

func ToDurationMap(value any) map[any]time.Duration {
	return ToMap(value, ToAny, ToDuration)
}

func ToTimeMap(value any) map[any]time.Time {
	return ToMap(value, ToAny, ToTime)
}

func ToUUIDMap(value any) map[any]uuid.UUID {
	return ToMap(value, ToAny, ToUUID)
}

func ToIPMap(value any) map[any]net.IP {
	return ToMap(value, ToAny, ToIP)
}
