package reflection

import (
	"iter"
	"reflect"
	"strings"
)

var structTags = []string{
	"json",
	"yaml",
	"toml",
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
	_, ok := BoolValue(v)

	return ok
}

func IsString(v any) bool {
	_, ok := StringValue(v)

	return ok
}

func IsInt(v any) bool {
	_, ok := IntValue(v)

	return ok
}

func IsUint(v any) bool {
	_, ok := UintValue(v)

	return ok
}

func IsFloat(v any) bool {
	_, ok := FloatValue(v)

	return ok
}

func IsNumber(v any) bool {
	_, ok := NumberValue(v)

	return ok
}

func IsSlice(v any) bool {
	_, ok := SliceValue(v)

	return ok
}

func IsMap(v any) bool {
	_, ok := MapValue(v)

	return ok
}

func IsStruct(v any) bool {
	_, ok := StructValue(v)

	return ok
}

func BoolValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && rv.Kind() == reflect.Bool {
		return rv, true
	}

	return reflect.Value{}, false
}

func StringValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && rv.Kind() == reflect.String {
		return rv, true
	}

	return reflect.Value{}, false
}

func IntValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && rv.CanInt() {
		return rv, true
	}

	return reflect.Value{}, false
}

func UintValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && rv.CanUint() {
		return rv, true
	}

	return reflect.Value{}, false
}

func FloatValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && rv.CanFloat() {
		return rv, true
	}

	return reflect.Value{}, false
}

func NumberValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && (rv.CanInt() || rv.CanUint() || rv.CanFloat()) {
		return rv, true
	}

	return reflect.Value{}, false
}

func SliceValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		return rv, true
	}

	return reflect.Value{}, false
}

func MapValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && rv.Kind() == reflect.Map {
		return rv, true
	}

	return reflect.Value{}, false
}

func StructValue(v any) (reflect.Value, bool) {
	if rv := IndirectValue(v); rv.IsValid() && rv.Kind() == reflect.Struct {
		return rv, true
	}

	return reflect.Value{}, false
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
