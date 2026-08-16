package functions

import (
	"strings"
	"time"

	"github.com/JFAexe/tem/pkg/convert"
)

var layouts = map[string]string{
	"ansic":    time.ANSIC,
	"unix":     time.UnixDate,
	"ruby":     time.RubyDate,
	"822":      time.RFC822,
	"822z":     time.RFC822Z,
	"850":      time.RFC850,
	"1123":     time.RFC1123,
	"1123z":    time.RFC1123Z,
	"3339":     time.RFC3339,
	"3339nano": time.RFC3339Nano,
	"kitchen":  time.Kitchen,
	"stamp":    time.Stamp,
	"datetime": time.DateTime,
	"date":     time.DateOnly,
	"time":     time.TimeOnly,
}

type Time struct {
	cache map[string]*time.Location
}

func (*Time) Now() time.Time {
	return time.Now()
}

func (f *Time) Parse(layout, value any) (time.Time, error) {
	return time.Parse(convert.ToString(value), convert.ToString(layout))
}

func (f *Time) In(zone, value any) (time.Time, error) {
	loc, err := f.cached(convert.ToString(zone))
	if err != nil {
		return time.Time{}, err
	}

	return convert.ToTime(value).In(loc), nil
}

func (f *Time) ParseIn(layout, zone, value any) (time.Time, error) {
	loc, err := f.cached(convert.ToString(zone))
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation(convert.ToString(value), convert.ToString(layout), loc)
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

func (*Time) Layout(value any) string {
	if layout, ok := layouts[strings.ToLower(convert.ToString(value))]; ok {
		return layout
	}

	return time.RFC3339
}

func (f *Time) cached(zone string) (*time.Location, error) {
	if f.cache == nil {
		f.cache = make(map[string]*time.Location)
	}

	if exp, ok := f.cache[zone]; ok {
		return exp, nil
	}

	loc, err := time.LoadLocation(zone)
	if err != nil {
		return nil, err
	}

	f.cache[zone] = loc

	return loc, nil
}
