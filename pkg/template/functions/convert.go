package functions

import (
	"fmt"
	"maps"
	"net"
	"reflect"
	"time"
	"uuid"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

var convertTypes = map[string]any{
	"Any":         new(Convert).Any,
	"Bool":        new(Convert).Bool,
	"String":      new(Convert).String,
	"Rune":        new(Convert).Rune,
	"Int":         new(Convert).Int,
	"Uint":        new(Convert).Uint,
	"Float":       new(Convert).Float,
	"Duration":    new(Convert).Duration,
	"Time":        new(Convert).Time,
	"UUID":        new(Convert).UUID,
	"IP":          new(Convert).IP,
	"List":        new(Convert).List,
	"Bools":       new(Convert).Bools,
	"Strings":     new(Convert).Strings,
	"Ints":        new(Convert).Ints,
	"Uints":       new(Convert).Uints,
	"Floats":      new(Convert).Floats,
	"Durations":   new(Convert).Durations,
	"Times":       new(Convert).Times,
	"UUIDs":       new(Convert).UUIDs,
	"IPs":         new(Convert).IPs,
	"Runes":       new(Convert).Runes,
	"Bytes":       new(Convert).Bytes,
	"Map":         new(Convert).Map,
	"BoolMap":     new(Convert).BoolMap,
	"StringMap":   new(Convert).StringMap,
	"IntMap":      new(Convert).IntMap,
	"UintMap":     new(Convert).UintMap,
	"FloatMap":    new(Convert).FloatMap,
	"DurationMap": new(Convert).DurationMap,
	"TimeMap":     new(Convert).TimeMap,
	"UUIDMap":     new(Convert).UUIDMap,
	"IPMap":       new(Convert).IPMap,
}

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

func (*Convert) Int(value any) int {
	return convert.ToInt(value)
}

func (*Convert) Uint(value any) uint {
	return convert.ToUint(value)
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

func (*Convert) UUID(value any) uuid.UUID {
	return convert.ToUUID(value)
}

func (*Convert) IP(value any) net.IP {
	return convert.ToIP(value)
}

func (*Convert) List(args ...any) (any, error) {
	switch len(args) {
	case 0:
		return make([]any, 0), nil
	case 1:
		if arg := args[0]; !reflection.IsFunc(arg) {
			return convert.ToSlice(arg, convert.ToAny), nil
		}

		return convertList(make([]any, 0), args[0])
	case 2:
		return convertList(args[1], args[0])
	default:
		return nil, ErrTooManyArguments
	}
}

func (*Convert) Bools(value any) []bool {
	return convert.ToSlice(value, convert.ToBool)
}

func (*Convert) Strings(value any) []string {
	return convert.ToSlice(value, convert.ToString)
}

func (*Convert) Ints(value any) []int {
	return convert.ToSlice(value, convert.ToInt)
}

func (*Convert) Uints(value any) []uint {
	return convert.ToSlice(value, convert.ToUint)
}

func (*Convert) Floats(value any) []float64 {
	return convert.ToSlice(value, convert.ToFloat64)
}

func (*Convert) Durations(value any) []time.Duration {
	return convert.ToSlice(value, convert.ToDuration)
}

func (*Convert) Times(value any) []time.Time {
	return convert.ToSlice(value, convert.ToTime)
}

func (*Convert) UUIDs(value any) []uuid.UUID {
	return convert.ToSlice(value, convert.ToUUID)
}

func (*Convert) IPs(value any) []net.IP {
	return convert.ToSlice(value, convert.ToIP)
}

func (*Convert) Runes(value any) []rune {
	return convert.ToRuneSlice(value)
}

func (*Convert) Bytes(value any) []byte {
	return convert.ToByteSlice(value)
}

func (*Convert) Map(args ...any) (any, error) {
	switch len(args) {
	case 0:
		return make(map[string]any, 0), nil
	case 1:
		if arg := args[0]; !reflection.IsFunc(arg) {
			return convert.ToMap(arg, convert.ToAny, convert.ToAny), nil
		}

		return convertMap(make(map[string]any, 0), convert.ToAny, args[0])
	case 2:
		if !reflection.IsFunc(args[0]) {
			return nil, fmt.Errorf("expected func(any) T, got %T", args[0])
		}

		var (
			vf  = args[1]
			out = any(make(map[string]any, 0))
		)

		if !reflection.IsFunc(vf) {
			out = vf
			vf = convert.ToAny
		}

		return convertMap(out, args[0], vf)
	case 3:
		return convertMap(args[2], args[0], args[1])
	default:
		return nil, ErrTooManyArguments
	}
}

func (*Convert) BoolMap(value any) map[any]bool {
	return convert.ToMap(value, convert.ToAny, convert.ToBool)
}

func (*Convert) StringMap(value any) map[any]string {
	return convert.ToMap(value, convert.ToAny, convert.ToString)
}

func (*Convert) IntMap(value any) map[any]int {
	return convert.ToMap(value, convert.ToAny, convert.ToInt)
}

func (*Convert) UintMap(value any) map[any]uint {
	return convert.ToMap(value, convert.ToAny, convert.ToUint)
}

func (*Convert) FloatMap(value any) map[any]float64 {
	return convert.ToMap(value, convert.ToAny, convert.ToFloat64)
}

func (*Convert) DurationMap(value any) map[any]time.Duration {
	return convert.ToMap(value, convert.ToAny, convert.ToDuration)
}

func (*Convert) TimeMap(value any) map[any]time.Time {
	return convert.ToMap(value, convert.ToAny, convert.ToTime)
}

func (*Convert) UUIDMap(value any) map[any]uuid.UUID {
	return convert.ToMap(value, convert.ToAny, convert.ToUUID)
}

func (*Convert) IPMap(value any) map[any]net.IP {
	return convert.ToMap(value, convert.ToAny, convert.ToIP)
}

func (*Convert) Type() any {
	return maps.Clone(convertTypes)
}

func convertList(l, v any) (any, error) {
	vf, vt, err := adapt(v)
	if err != nil {
		return nil, err
	}

	return convert.ToSliceDynamic(l, vf, vt)
}

func convertMap(m, k, v any) (any, error) {
	kf, kt, err := adapt(k)
	if err != nil {
		return nil, fmt.Errorf("key converter: %w", err)
	}

	vf, vt, err := adapt(v)
	if err != nil {
		return nil, fmt.Errorf("value converter: %w", err)
	}

	return convert.ToMapDynamic(m, kf, vf, kt, vt)
}

func adapt(v any) (func(any) any, reflect.Type, error) {
	rv := reflection.IndirectValue(v)

	if !rv.IsValid() || rv.Kind() != reflect.Func || rv.Type().NumIn() != 1 || rv.Type().NumOut() != 1 {
		return nil, nil, fmt.Errorf("expected func(any) T, got %T", v)
	}

	return func(a any) any {
		return rv.Call([]reflect.Value{reflect.ValueOf(a)})[0].Interface()
	}, rv.Type().Out(0), nil
}
