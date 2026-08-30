package functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/JFAexe/tem/pkg/cache"
	"github.com/JFAexe/tem/pkg/convert"
)

var layouts = map[string]string{
	"ansic":       time.ANSIC,
	"unixdate":    time.UnixDate,
	"ruby":        time.RubyDate,
	"rfc822":      time.RFC822,
	"rfc822z":     time.RFC822Z,
	"rfc850":      time.RFC850,
	"rfc1123":     time.RFC1123,
	"rfc1123z":    time.RFC1123Z,
	"rfc3339":     time.RFC3339,
	"rfc3339nano": time.RFC3339Nano,
	"kitchen":     time.Kitchen,
	"stamp":       time.Stamp,
	"datetime":    time.DateTime,
	"date":        time.DateOnly,
	"time":        time.TimeOnly,
}

var locationCache = cache.NewSyncCache[string, *time.Location]()

type Time struct{}

func (*Time) Now() time.Time {
	return time.Now()
}

func (*Time) Parse(layout, value any) (time.Time, error) {
	if l, ok := layouts[normalizeString(value)]; ok {
		layout = l
	}

	return time.Parse(convert.ToString(layout), convert.ToString(value))
}

func (*Time) In(zone, value any) (time.Time, error) {
	loc, err := cachedLocation(convert.ToString(zone))
	if err != nil {
		return time.Time{}, err
	}

	return convert.ToTime(value).In(loc), nil
}

func (*Time) ParseIn(layout, zone, value any) (time.Time, error) {
	if l, ok := layouts[normalizeString(value)]; ok {
		layout = l
	}

	loc, err := cachedLocation(convert.ToString(zone))
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation(convert.ToString(layout), convert.ToString(value), loc)
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
	if l, ok := layouts[normalizeString(value)]; ok {
		format = l
	}

	return convert.ToTime(value).Format(convert.ToString(format))
}

func (*Time) ToString(value any) string {
	return convert.ToTime(value).Format(time.RFC3339)
}

func (*Time) ToUnix(value any) int64 {
	return convert.ToTime(value).Unix()
}

func (*Time) AsTime(value any) string {
	return convert.ToTime(value).Format(time.TimeOnly)
}

func (*Time) AsDate(value any) string {
	return convert.ToTime(value).Format(time.DateOnly)
}

func (*Time) AsDateTime(value any) string {
	return convert.ToTime(value).Format(time.DateTime)
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

func (*Time) Layout(value any) (string, error) {
	k := normalizeString(value)

	layout, ok := layouts[k]
	if !ok {
		return "", fmt.Errorf("invalid layout %#q, supported: %s", k, joinKeys(layouts))
	}

	return layout, nil
}

func cachedLocation(zone string) (*time.Location, error) {
	zone = strings.TrimSpace(zone)

	return locationCache.Get(zone, func() (*time.Location, error) {
		return time.LoadLocation(zone)
	})
}
