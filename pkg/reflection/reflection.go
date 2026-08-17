package reflection

import (
	"errors"
	"fmt"
	"iter"
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

func IndirectValue(v reflect.Value) reflect.Value {
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

func CompareValues(v, target reflect.Value) bool {
	if !v.IsValid() || !target.IsValid() {
		return v.IsValid() == target.IsValid()
	}

	if v = IndirectValue(v); !v.IsValid() {
		return false
	}

	if v.Type() == target.Type() {
		return reflect.DeepEqual(v.Interface(), target.Interface())
	}

	if v.Type().ConvertibleTo(target.Type()) {
		return reflect.DeepEqual(v.Convert(target.Type()).Interface(), target.Interface())
	}

	return false
}

func ToIndex(i any) (int64, error) {
	if i == nil {
		return 0, ErrInvalidIndex
	}

	switch v := i.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case uintptr:
		return int64(v), nil
	}

	return 0, ErrInvalidIndex
}

func ExtractKey(item any, key any) any {
	if item == nil {
		return nil
	}

	rv := IndirectValue(reflect.ValueOf(item))
	if !rv.IsValid() {
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

func ResolveKey(m reflect.Value, key any) (reflect.Value, error) {
	kt := m.Type().Key()

	if key == nil {
		return reflect.Zero(kt), nil
	}

	kv := reflect.ValueOf(key)

	if !kv.Type().ConvertibleTo(kt) {
		return reflect.Value{}, ErrTypeMismatch
	}

	return kv.Convert(kt), nil
}

func ResolveField(v reflect.Value, key any) (reflect.Value, error) {
	ks, ok := key.(string)
	if !ok {
		return reflect.Value{}, ErrInvalidIndex
	}

	for name, field := range ExportedFields(v) {
		if name == ks {
			return field, nil
		}
	}

	return reflect.Value{}, ErrKeyMissing
}

func ExportedFields(rv reflect.Value) iter.Seq2[string, reflect.Value] {
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
