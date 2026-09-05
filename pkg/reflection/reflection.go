package reflection

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

var (
	ErrNilPointer   = errors.New("nil pointer")
	ErrInvalidIndex = errors.New("invalid index")
	ErrOutOfRange   = errors.New("index out of range")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrKeyMissing   = errors.New("key missing")
	ErrNotSettable  = errors.New("target is not settable")
)

func Lookup(target, key any) (reflect.Value, error) {
	v := IndirectValue(target)
	if !v.IsValid() {
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

func Set(target, key, value any) error {
	v := IndirectValue(target)
	if !v.IsValid() {
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

func Unset(target, key any) error {
	v := IndirectValue(target)
	if !v.IsValid() {
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
		return ErrNotSettable
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
