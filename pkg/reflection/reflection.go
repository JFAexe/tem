package reflection

import (
	"errors"
	"fmt"
	"iter"
	"math"
	"reflect"
	"strings"
)

var (
	ErrNilPointer   = errors.New("nil pointer")
	ErrInvalidIndex = errors.New("invalid index")
	ErrOutOfRange   = errors.New("index out of range")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrKeyMissing   = errors.New("key missing")
)

var structTags = []string{
	"json",
	"yaml",
	"toml",
}

func Zero(value any) any {
	if value == nil {
		return nil
	}

	return reflect.Zero(reflect.TypeOf(value)).Interface()
}

func IsZero(value any) bool {
	v := IndirectValue(value)

	return !v.IsValid() || v.IsZero()
}

func IsEmpty(value any) bool {
	v := IndirectValue(value)
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	}

	return v.IsZero()
}

func IsBool(v any) bool {
	rv := IndirectValue(v)

	return rv.IsValid() && rv.Kind() == reflect.Bool
}

func IsNumber(v any) bool {
	rv := IndirectValue(v)

	return rv.IsValid() && (rv.CanInt() || rv.CanUint() || rv.CanFloat())
}

func IsSlice(v any) bool {
	_, ok := SliceValue(v)

	return ok
}

func IsMap(v any) bool {
	_, ok := MapValue(v)

	return ok
}

func SliceValue(v any) (reflect.Value, bool) {
	rv := IndirectValue(v)

	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		return rv, true
	}

	return rv, false
}

func MapValue(v any) (reflect.Value, bool) {
	rv := IndirectValue(v)

	if rv.IsValid() && rv.Kind() == reflect.Map {
		return rv, true
	}

	return rv, false
}

func IndirectValue(value any) reflect.Value {
	v, ok := value.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(value)
	}

	for {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return reflect.Value{}
			}

			v = v.Elem()
		default:
			return v
		}
	}
}

func Lookup(v reflect.Value, key any) (reflect.Value, error) {
	if v = IndirectValue(v); !v.IsValid() {
		return reflect.Value{}, ErrNilPointer
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		idx, err := ResolveIndex(key)
		if err != nil {
			return reflect.Value{}, err
		}

		if idx < 0 || idx >= int64(v.Len()) {
			return reflect.Value{}, ErrOutOfRange
		}

		return v.Index(int(idx)), nil
	case reflect.Map:
		kv, err := ResolveKey(v, key)
		if err != nil {
			return reflect.Value{}, err
		}

		val := v.MapIndex(kv)

		if !val.IsValid() {
			return reflect.Value{}, ErrKeyMissing
		}

		return val, nil
	case reflect.Struct:
		return ResolveField(v, key)
	}

	return reflect.Value{}, fmt.Errorf("cannot index into type %v", v.Kind())
}

func Set(v reflect.Value, key, value any) error {
	if v = IndirectValue(v); !v.IsValid() {
		return ErrNilPointer
	}

	val := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		idx, err := ResolveIndex(key)
		if err != nil {
			return fmt.Errorf("index: %w", err)
		}

		if idx < 0 || idx >= int64(v.Len()) {
			return fmt.Errorf("%w: %d not in [0:%d]", ErrOutOfRange, idx, v.Len())
		}

		return SetTarget(v.Index(int(idx)), val)
	case reflect.Map:
		if v.IsNil() {
			return fmt.Errorf("cannot set key in nil map")
		}

		kv, err := ResolveKey(v, key)
		if err != nil {
			return fmt.Errorf("key: %w", err)
		}

		converted, err := Convert(val, v.Type().Elem())
		if err != nil {
			return err
		}

		v.SetMapIndex(kv, converted)

		return nil
	case reflect.Struct:
		target, err := ResolveField(v, key)
		if err != nil {
			return fmt.Errorf("key: %w", err)
		}

		return SetTarget(target, val)
	}

	return fmt.Errorf("cannot set on type %v", v.Kind())
}

func Unset(v reflect.Value, key any) error {
	if v = IndirectValue(v); !v.IsValid() {
		return ErrNilPointer
	}

	switch v.Kind() {
	case reflect.Map:
		if v.IsNil() {
			return nil
		}

		kv, err := ResolveKey(v, key)
		if err != nil {
			return fmt.Errorf("key: %w", err)
		}

		v.SetMapIndex(kv, reflect.Value{})

		return nil
	case reflect.Struct:
		target, err := ResolveField(v, key)
		if err != nil {
			return fmt.Errorf("key: %w", err)
		}

		return SetTarget(target, reflect.Zero(target.Type()))
	case reflect.Slice, reflect.Array:
		return fmt.Errorf("cannot unset elements from slice or array, use Set with zero value instead")
	}

	return fmt.Errorf("cannot unset on type %v", v.Kind())
}

func SetTarget(target, val reflect.Value) error {
	if !target.CanSet() {
		return fmt.Errorf("target is not settable (unaddressable)")
	}

	converted, err := Convert(val, target.Type())
	if err != nil {
		return err
	}

	target.Set(converted)

	return nil
}

func Convert(v reflect.Value, target reflect.Type) (reflect.Value, error) {
	if !v.IsValid() {
		return reflect.Zero(target), nil
	}

	if v.Type() == target {
		return v, nil
	}

	if !v.Type().ConvertibleTo(target) {
		return reflect.Value{}, fmt.Errorf("value type %v not convertible to %v", v.Type(), target)
	}

	return v.Convert(target), nil
}

func Compare(v, target reflect.Value) bool {
	v, target = IndirectValue(v), IndirectValue(target)

	if !v.IsValid() || !target.IsValid() {
		return v.IsValid() == target.IsValid()
	}

	if v.Type() == target.Type() {
		return reflect.DeepEqual(v.Interface(), target.Interface())
	}

	if IsNumber(v) && IsNumber(target) {
		var (
			cv = target.Convert(v.Type())
			rt = cv.Convert(target.Type())
		)

		if rt.Interface() == target.Interface() {
			return v.Interface() == cv.Interface()
		}

		return false
	}

	return false
}

func ResolveIndex(i any) (int64, error) {
	switch v := reflect.ValueOf(i); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if u := v.Uint(); u <= math.MaxInt64 {
			return int64(u), nil
		}
	case reflect.Float32, reflect.Float64:
		if f := v.Float(); !math.IsNaN(f) && !math.IsInf(f, 0) && f == math.Trunc(f) && f >= math.MinInt64 && f < math.MaxInt64 {
			return int64(f), nil
		}
	}

	return 0, ErrInvalidIndex
}

func ResolveKey(m reflect.Value, key any) (reflect.Value, error) {
	kt := m.Type().Key()

	if key == nil {
		return reflect.Zero(kt), nil
	}

	kv, err := Convert(reflect.ValueOf(key), kt)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("%w: %v", ErrTypeMismatch, err)
	}

	return kv, nil
}

func ResolveField(v reflect.Value, key any) (reflect.Value, error) {
	ks, ok := key.(string)
	if !ok {
		return reflect.Value{}, ErrInvalidIndex
	}

	for name, field := range Fields(v) {
		if name == ks {
			return field, nil
		}
	}

	return reflect.Value{}, ErrKeyMissing
}

func Fields(rv reflect.Value) iter.Seq2[string, reflect.Value] {
	return func(yield func(name string, field reflect.Value) bool) {
		if rv.Kind() != reflect.Struct {
			return
		}

		typ := rv.Type()

		for i := range typ.NumField() {
			field := typ.Field(i)

			if !field.IsExported() {
				continue
			}

			var (
				name = field.Name
				skip = false
			)

			for _, tag := range structTags {
				val := field.Tag.Get(tag)

				if val == "" {
					continue
				}

				if before, _, ok := strings.Cut(val, ","); ok {
					val = before
				}

				if val == "-" {
					skip = true

					break
				}

				if val != "" {
					name = val

					break
				}
			}

			if skip {
				continue
			}

			if !yield(name, rv.Field(i)) {
				break
			}
		}
	}
}

func Values(v any) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		rv := IndirectValue(v)

		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			for i := range rv.Len() {
				if !yield(rv.Index(i).Interface(), nil) {
					return
				}
			}
		case reflect.Map:
			for iter := rv.MapRange(); iter.Next(); {
				if !yield(iter.Value().Interface(), nil) {
					return
				}
			}
		case reflect.Struct:
			for _, field := range Fields(rv) {
				if !yield(field.Interface(), nil) {
					return
				}
			}
		default:
			yield(v, nil)
		}
	}
}
