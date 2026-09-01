package functions

import "github.com/JFAexe/tem/pkg/reflection"

type Type struct{}

func (*Type) IsBool(value any) bool {
	return reflection.IsBool(value)
}

func (*Type) IsString(value any) bool {
	return reflection.IsString(value)
}

func (*Type) IsInt(value any) bool {
	return reflection.IsInt(value)
}

func (*Type) IsUint(value any) bool {
	return reflection.IsUint(value)
}

func (*Type) IsFloat(value any) bool {
	return reflection.IsFloat(value)
}

func (*Type) IsNumber(value any) bool {
	return reflection.IsNumber(value)
}

func (*Type) IsList(value any) bool {
	return reflection.IsSlice(value)
}

func (*Type) IsMap(value any) bool {
	return reflection.IsMap(value)
}

func (*Type) IsStruct(value any) bool {
	return reflection.IsStruct(value)
}
