package reflection

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrNilPointer   = errors.New("nil pointer")
	ErrInvalidIndex = errors.New("invalid index")
	ErrOutOfRange   = errors.New("index out of range")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrKeyMissing   = errors.New("map key missing")
)

var structTags = []string{
	"json",
	"yaml",
	"toml",
}

func IndirectValue(v reflect.Value) (reflect.Value, error) {
	for {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return reflect.Value{}, ErrNilPointer
			}

			v = v.Elem()
		default:
			return v, nil
		}
	}
}

func Traverse(v reflect.Value, key any) (reflect.Value, error) {
	v, err := IndirectValue(v)
	if err != nil {
		return reflect.Value{}, err
	}

	if !v.IsValid() {
		return reflect.Value{}, ErrInvalidIndex
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		idx, err := ToIndex(key)
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

func ConvertValue(v reflect.Value, target reflect.Type) (reflect.Value, error) {
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

func ToIndex(i any) (int64, error) {
	iv := reflect.ValueOf(i)

	if !iv.IsValid() {
		return 0, ErrInvalidIndex
	}

	switch iv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return iv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(iv.Uint()), nil
	}

	return 0, ErrInvalidIndex
}

func ResolveKey(m reflect.Value, key any) (reflect.Value, error) {
	kv := reflect.ValueOf(key)

	if !kv.IsValid() {
		kv = reflect.Zero(m.Type().Key())
	}

	if !kv.Type().ConvertibleTo(m.Type().Key()) {
		return reflect.Value{}, ErrTypeMismatch
	}

	return kv.Convert(m.Type().Key()), nil
}

func ExtractKey(item any, key any) any {
	rv, err := IndirectValue(reflect.ValueOf(item))
	if err != nil || !rv.IsValid() {
		return nil
	}

	switch rv.Kind() {
	case reflect.Map:
		kv, err := ResolveKey(rv, key)
		if err != nil {
			return nil
		}

		if v := rv.MapIndex(kv); v.IsValid() {
			return v.Interface()
		}
	case reflect.Struct:
		f, err := ResolveField(rv, key)
		if err != nil || !f.IsValid() {
			return nil
		}

		return f.Interface()
	}

	return nil
}

func ResolveField(v reflect.Value, key any) (reflect.Value, error) {
	ks, ok := key.(string)
	if !ok {
		return reflect.Value{}, ErrInvalidIndex
	}

	var (
		value reflect.Value
		found bool
	)

	ForEachExportedField(v, func(name string, field reflect.Value) bool {
		if name == ks {
			value = field
			found = true

			return false
		}

		return true
	})

	if !found {
		return reflect.Value{}, ErrKeyMissing
	}

	return value, nil
}

func ForEachExportedField(rv reflect.Value, fn func(name string, field reflect.Value) bool) {
	if rv.Kind() != reflect.Struct {
		return
	}

	typ := rv.Type()

	for i := 0; i < typ.NumField(); i++ {
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

		if !fn(name, rv.Field(i)) {
			break
		}
	}
}
