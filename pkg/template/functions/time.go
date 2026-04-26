package functions

import "time"

type Time struct{}

func (*Time) Now() time.Time {
	return time.Now()
}

func (*Time) Offset(offset string, t time.Time) (time.Time, error) {
	dur, err := time.ParseDuration(offset)
	if err != nil {
		return time.Time{}, err
	}

	return t.Add(dur), nil
}

func (*Time) Truncate(step string, t time.Time) (time.Time, error) {
	dur, err := time.ParseDuration(step)
	if err != nil {
		return time.Time{}, err
	}

	return t.Truncate(dur), nil
}

func (*Time) Round(step string, t time.Time) (time.Time, error) {
	dur, err := time.ParseDuration(step)
	if err != nil {
		return time.Time{}, err
	}

	return t.Round(dur), nil
}

func (*Time) UTC(t time.Time) time.Time {
	return t.UTC()
}

func (*Time) Local(t time.Time) time.Time {
	return t.Local()
}

func (*Time) Format(format string, t time.Time) string {
	return t.Format(format)
}

func (*Time) String(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (*Time) Time(t time.Time) string {
	return t.Format(time.TimeOnly)
}

func (*Time) Date(t time.Time) string {
	return t.Format(time.DateOnly)
}

func (*Time) DateTime(t time.Time) string {
	return t.Format(time.DateTime)
}

func (*Time) Unix(t time.Time) int64 {
	return t.Unix()
}

func (*Time) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (*Time) Until(t time.Time) time.Duration {
	return time.Until(t)
}
