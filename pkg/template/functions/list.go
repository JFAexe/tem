package functions

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"slices"
	"strings"
	"time"
	"uuid"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

type List struct{}

func ListVarargInit(n *List, args []any) (any, error) {
	return n.New(args...), nil
}

func (*List) New(values ...any) []any {
	return values
}

func (*List) Range(args ...any) []int {
	var lower, upper, step int64

	switch len(args) {
	case 0:
		return make([]int, 0)
	case 1:
		upper = convert.ToInt64(args[0])
		step = 1
	default:
		lower = convert.ToInt64(args[0])
		upper = convert.ToInt64(args[1])
		step = 1

		if len(args) > 2 {
			step = convert.ToInt64(args[2])
		}
	}

	if step == 0 {
		return make([]int, 0)
	}

	if upper < lower && step == 1 {
		step = -1
	}

	var count int64

	switch {
	case step > 0 && upper > lower:
		count = (upper - lower + step - 1) / step
	case step < 0 && upper < lower:
		count = (lower - upper - step - 1) / -step
	default:
		return make([]int, 0)
	}

	count = min(count, maxRangeLength)

	out := make([]int, count)

	for i := int64(0); i < count; i++ {
		out[i] = int(lower + i*step)
	}

	return out
}

func (*List) First(items any) any {
	rv, ok := reflection.SliceValue(items)
	if !ok || rv.Len() == 0 {
		return nil
	}

	return rv.Index(0).Interface()
}

func (*List) Last(items any) any {
	rv, ok := reflection.SliceValue(items)
	if !ok || rv.Len() == 0 {
		return nil
	}

	return rv.Index(rv.Len() - 1).Interface()
}

func (*List) Init(items any) []any {
	if out := convert.ToAnySlice(items); len(out) > 1 {
		return out[:len(out)-1]
	}

	return make([]any, 0)
}

func (*List) Tail(items any) []any {
	if out := convert.ToAnySlice(items); len(out) > 1 {
		return out[1:]
	}

	return make([]any, 0)
}

func (*List) Where(args ...any) ([]any, error) {
	items, pred, err := popPredicate(args)
	if err != nil {
		return nil, err
	}

	if !reflection.IsSlice(items) {
		return nil, fmt.Errorf("where: expected a list, got %T", items)
	}

	out := make([]any, 0)

	for item, err := range reflection.Values(items) {
		if err != nil {
			return nil, fmt.Errorf("where: %w", err)
		}

		ok, e := pred(item)
		if e != nil {
			return nil, fmt.Errorf("where: %w", e)
		}

		if ok {
			out = append(out, item)
		}
	}

	return out, nil
}

func (*List) WhereBy(key, value, items any) ([]any, error) {
	s := convert.ToAnySlice(items)

	out := make([]any, 0, len(s))

	for _, item := range s {
		if v, err := reflection.Lookup(reflect.ValueOf(item), key); err == nil && equalAny(v.Interface(), value) {
			out = append(out, item)
		}
	}

	return out, nil
}

func (*List) Remove(args ...any) ([]any, error) {
	items, pred, err := popPredicate(args)
	if err != nil {
		return nil, err
	}

	if !reflection.IsSlice(items) {
		return nil, fmt.Errorf("remove: expected a list, got %T", items)
	}

	out := make([]any, 0)

	for item, err := range reflection.Values(items) {
		if err != nil {
			return nil, fmt.Errorf("remove: %w", err)
		}

		ok, e := pred(item)
		if e != nil {
			return nil, fmt.Errorf("remove: %w", e)
		}

		if !ok {
			out = append(out, item)
		}
	}

	return out, nil
}

func (*List) RemoveBy(key, value, items any) ([]any, error) {
	s := convert.ToAnySlice(items)

	out := make([]any, 0, len(s))

	for _, item := range s {
		if v, err := reflection.Lookup(reflect.ValueOf(item), key); err != nil || !equalAny(v.Interface(), value) {
			out = append(out, item)
		}
	}

	return out, nil
}

func (*List) Append(args ...any) []any {
	if len(args) == 0 {
		return nil
	}

	item, values := popOne(args)

	return append(convert.ToAnySlice(item), values...)
}

func (*List) Prepend(args ...any) []any {
	if len(args) == 0 {
		return nil
	}

	item, values := popOne(args)

	return append(values, convert.ToAnySlice(item)...)
}

func (*List) Concat(values ...any) []any {
	return listConcat(values...)
}

func (*List) Flatten(items any) []any {
	return listFlatten(convert.ToAnySlice(items))
}

func (*List) Compact(items any) []any {
	return slices.CompactFunc(convert.ToAnySlice(items), equalAny)
}

func (*List) Reverse(items any) []any {
	out := convert.ToAnySlice(items)

	slices.Reverse(out)

	return out
}

func (*List) Sort(items any) ([]any, error) {
	out := convert.ToAnySlice(items)

	if len(out) > 1 {
		slices.SortStableFunc(out, compareAny)
	}

	return out, nil
}

func (*List) SortBy(key, items any) ([]any, error) {
	out := convert.ToAnySlice(items)
	if len(out) <= 1 {
		return out, nil
	}

	type kv struct {
		key any
		val any
	}

	pairs := make([]kv, len(out))

	for i, item := range out {
		var k any

		if v, err := reflection.Lookup(reflect.ValueOf(item), key); err == nil {
			k = v.Interface()
		}

		pairs[i] = kv{k, item}
	}

	slices.SortStableFunc(pairs, func(a, b kv) int { return compareAny(a.key, b.key) })

	for i, p := range pairs {
		out[i] = p.val
	}

	return out, nil
}

func (*List) Unique(items any) []any {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s
	}

	return uniqueBy(s, convert.ToAny)
}

func (*List) UniqueBy(key, items any) []any {
	s := convert.ToAnySlice(items)

	if len(s) <= 1 {
		return s
	}

	return uniqueBy(s, func(v any) any {
		if v, err := reflection.Lookup(reflect.ValueOf(v), key); err == nil {
			return v.Interface()
		}

		return &v
	})
}

func (*List) Chunk(size, items any) [][]any {
	var (
		n = convert.ToInt(size)
		s = convert.ToAnySlice(items)
	)

	if n <= 0 || len(s) == 0 {
		return make([][]any, 0)
	}

	out := make([][]any, 0, (len(s)+n-1)/n)

	for start := 0; start < len(s); start += n {
		var (
			end   = min(start+n, len(s))
			chunk = make([]any, end-start)
		)

		copy(chunk, s[start:end])

		out = append(out, chunk)
	}

	return out
}

func (*List) Zip(values ...any) [][]any {
	var (
		lists  = make([][]any, 0, len(values))
		minLen = -1
	)

	for _, v := range values {
		s := convert.ToAnySlice(v)

		lists = append(lists, s)

		if minLen == -1 || len(s) < minLen {
			minLen = len(s)
		}
	}

	if minLen <= 0 {
		return make([][]any, 0)
	}

	out := make([][]any, minLen)

	for i := range out {
		pair := make([]any, len(lists))

		for j, s := range lists {
			pair[j] = s[i]
		}

		out[i] = pair
	}

	return out
}

func (*List) Unzip(items any) [][]any {
	var (
		width int

		s = convert.ToAnySlice(items)
	)

	for _, item := range s {
		if n := len(convert.ToAnySlice(item)); n > width {
			width = n
		}
	}

	out := make([][]any, width)

	if width == 0 {
		return out
	}

	for i := range out {
		out[i] = make([]any, 0, len(s))
	}

	for _, item := range s {
		for j, v := range convert.ToAnySlice(item) {
			out[j] = append(out[j], v)
		}
	}

	return out
}

func listConcat(values ...any) []any {
	out := make([][]any, 0, len(values))

	for _, v := range values {
		out = append(out, convert.ToAnySlice(v))
	}

	return slices.Concat(out...)
}

func listFlatten(values []any) []any {
	out := make([]any, 0, len(values))

	for _, v := range values {
		out = append(out, convert.ToAnySlice(v)...)
	}

	return out
}

func uniqueBy(s []any, key func(any) any) []any {
	var (
		out  = make([]any, 0, len(s))
		seen = make(map[any]struct{}, len(s))
	)

	for _, item := range s {
		var (
			k = key(item)
			t = reflect.TypeOf(k)
		)

		if k == nil || (t != nil && t.Comparable()) {
			if _, ok := seen[k]; ok {
				continue
			}

			seen[k] = struct{}{}
			out = append(out, item)

			continue
		}

		var skip bool

		for _, v := range out {
			if equalAny(k, key(v)) {
				skip = true

				break
			}
		}

		if !skip {
			out = append(out, item)
		}
	}

	return out
}

func equalAny(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}

	switch a := a.(type) {
	case bool:
		if b, ok := b.(bool); ok {
			return a == b
		}
	case int:
		if b, ok := b.(int); ok {
			return a == b
		}
	case string:
		if b, ok := b.(string); ok {
			return a == b
		}
	case time.Time:
		if b, ok := b.(time.Time); ok {
			return a.Equal(b)
		}
	case uuid.UUID:
		if b, ok := b.(uuid.UUID); ok {
			return a.Compare(b) == 0
		}
	}

	return reflection.Compare(reflect.ValueOf(a), reflect.ValueOf(b))
}

func compareAny(a, b any) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	}

	if ja, ok := a.(json.Number); ok {
		if jb, ok := b.(json.Number); ok {
			return cmp.Compare(convert.ToFloat64(ja), convert.ToFloat64(jb))
		}
	}

	if ta, ok := a.(time.Time); ok {
		if tb, ok := b.(time.Time); ok {
			return ta.Compare(tb)
		}
	}

	if ua, ok := a.(uuid.UUID); ok {
		if ub, ok := b.(uuid.UUID); ok {
			return ua.Compare(ub)
		}
	}

	if reflection.IsNumber(a) && reflection.IsNumber(b) {
		return cmp.Compare(convert.ToFloat64(a), convert.ToFloat64(b))
	}

	if ra, rb := reflection.IndirectValue(a), reflection.IndirectValue(b); ra.Kind() == reflect.Bool && rb.Kind() == reflect.Bool {
		switch {
		case ra.Bool() == rb.Bool():
			return 0
		case ra.Bool():
			return 1
		default:
			return -1
		}
	}

	return cmp.Compare(strings.ToLower(convert.ToString(a)), strings.ToLower(convert.ToString(b)))
}
