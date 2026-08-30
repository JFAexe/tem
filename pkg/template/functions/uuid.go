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

func (*UUID) Before(other, value any) bool {
	return convert.ToUUID(value).Compare(convert.ToUUID(other)) < 0
}

func (*UUID) After(other, value any) bool {
	return convert.ToUUID(value).Compare(convert.ToUUID(other)) > 0
}

func (*UUID) Version(value any) int {
	return convert.ToInt(convert.ToUUID(value)[6] >> 4)
}

func (*UUID) Time(value any) time.Time {
	u := convert.ToUUID(value)

	if u == uuid.Nil() || u[6]>>4 != 7 {
		return time.Time{}
	}

	return time.UnixMilli(int64(binary.BigEndian.Uint64(u[:8]) >> 16)).UTC()
}

func (*UUID) IsValid(value any) bool {
	if _, ok := value.(uuid.UUID); ok {
		return true
	}

	if rv := reflection.IndirectValue(value); rv.IsValid() && rv.Kind() == reflect.Array && rv.Len() == 16 && rv.Type().Elem().Kind() == reflect.Uint8 {
		return true
	}

	_, err := uuid.Parse(convert.ToString(value))

	return err == nil
}
