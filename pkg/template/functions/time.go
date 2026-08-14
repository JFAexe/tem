package functions

import (
	"time"

	"github.com/JFAexe/tem/pkg/convert"
)

type Time struct{}

func (*Time) Now() time.Time {
	return time.Now()
}

func (*Time) Offset(offset, value any) time.Time {
	return convert.ToTime(value).Add(convert.ToDuration(offset))
}

func (*Time) Truncate(step, value any) time.Time {
	return convert.ToTime(value).Truncate(convert.ToDuration(step))
}

func (*Time) Round(step, value any) time.Time {
	return convert.ToTime(value).Round(convert.ToDuration(step))
}

func (*Time) UTC(value any) time.Time {
	return convert.ToTime(value).UTC()
}

func (*Time) Local(value any) time.Time {
	return convert.ToTime(value).Local()
}

func (*Time) After(other, value any) bool {
	return convert.ToTime(value).After(convert.ToTime(other))
}

func (*Time) Before(other, value any) bool {
	return convert.ToTime(value).Before(convert.ToTime(other))
}

func (*Time) Equal(other, value any) bool {
	return convert.ToTime(value).Equal(convert.ToTime(other))
}

func (*Time) IsZero(value any) bool {
	return convert.ToTime(value).IsZero()
}

func (*Time) Format(format, value any) string {
	return convert.ToTime(value).Format(convert.ToString(format))
}

func (*Time) ToString(value any) string {
	return convert.ToTime(value).Format(time.RFC3339)
}

func (*Time) ToTime(value any) string {
	return convert.ToTime(value).Format(time.TimeOnly)
}

func (*Time) ToDate(value any) string {
	return convert.ToTime(value).Format(time.DateOnly)
}

func (*Time) ToDateTime(value any) string {
	return convert.ToTime(value).Format(time.DateTime)
}

func (*Time) ToUnix(value any) int64 {
	return convert.ToTime(value).Unix()
}

func (*Time) Difference(other, value any) time.Duration {
	return convert.ToTime(value).Sub(convert.ToTime(other))
}

func (*Time) Since(value any) time.Duration {
	return time.Since(convert.ToTime(value))
}

func (*Time) Until(value any) time.Duration {
	return time.Until(convert.ToTime(value))
}
