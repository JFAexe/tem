package functions

import (
	"encoding/binary"
	"reflect"
	"time"
	"uuid"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

type UUID struct{}

func (*UUID) Nil() uuid.UUID {
	return uuid.Nil()
}

func (*UUID) Max() uuid.UUID {
	return uuid.Max()
}

func (*UUID) V4() uuid.UUID {
	return uuid.NewV4()
}

func (*UUID) V7() uuid.UUID {
	return uuid.NewV7()
}

func (*UUID) Version(value any) int {
	return uuidVersion(value)
}

func (*UUID) Time(value any) time.Time {
	if u := convert.ToUUID(value); uuidIsV7(u) {
		return time.UnixMilli(int64(binary.BigEndian.Uint64(u[:8]) >> 16)).UTC()
	}

	return time.Time{}
}

func (*UUID) IsNil(value any) bool {
	return uuidIsNil(value)
}

func (*UUID) IsMax(value any) bool {
	return uuidIsMax(value)
}

func (*UUID) IsV4(value any) bool {
	return uuidIsV4(value)
}

func (*UUID) IsV7(value any) bool {
	return uuidIsV7(value)
}

func (*UUID) IsValid(value any) bool {
	return uuidIsValid(value)
}

func (*UUID) IsEqual(other, value any) bool {
	return uuidIsEqual(other, value)
}

func (*UUID) IsBefore(other, value any) bool {
	return uuidIsBefore(other, value)
}

func (*UUID) IsAfter(other, value any) bool {
	return uuidIsAfter(other, value)
}

func uuidIsValid(value any) bool {
	if value == nil {
		return false
	}

	if _, ok := value.(uuid.UUID); ok {
		return true
	}

	if rv := reflection.IndirectValue(value); rv.IsValid() && rv.Kind() == reflect.Array && rv.Len() == 16 && rv.Type().Elem().Kind() == reflect.Uint8 {
		return true
	}

	_, err := uuid.Parse(convert.ToString(value))

	return err == nil
}

func uuidIsNil(value any) bool {
	return uuidIsValid(value) && convert.ToUUID(value) == uuid.Nil()
}

func uuidIsMax(value any) bool {
	return uuidIsValid(value) && convert.ToUUID(value) == uuid.Max()
}

func uuidIsV4(value any) bool {
	return uuidVersion(value) == 4
}

func uuidIsV7(value any) bool {
	return uuidVersion(value) == 7
}

func uuidIsEqual(other, value any) bool {
	return convert.ToUUID(value).Compare(convert.ToUUID(other)) == 0
}

func uuidIsBefore(other, value any) bool {
	return convert.ToUUID(value).Compare(convert.ToUUID(other)) < 0
}

func uuidIsAfter(other, value any) bool {
	return convert.ToUUID(value).Compare(convert.ToUUID(other)) > 0
}

func uuidVersion(value any) int {
	return convert.ToInt(convert.ToUUID(value)[6] >> 4)
}
