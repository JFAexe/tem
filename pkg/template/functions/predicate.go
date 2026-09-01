package functions

import (
	"net"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/reflection"
)

type Predicate struct{}

func (*Predicate) Value() map[string]any {
	return map[string]any{
		"Zero":  IsZero,
		"Empty": IsEmpty,
		"Set":   IsSet,
		"In":    In,
	}
}

func (*Predicate) Type() map[string]any {
	return map[string]any{
		"Bool":   reflection.IsBool,
		"Int":    reflection.IsInt,
		"Uint":   reflection.IsUint,
		"Float":  reflection.IsFloat,
		"Number": reflection.IsNumber,
		"String": reflection.IsString,
		"List":   reflection.IsSlice,
		"Map":    reflection.IsMap,
		"Struct": reflection.IsStruct,
	}
}

func (*Predicate) Env() map[string]any {
	return map[string]any{
		"Set": envIsSet,
	}
}

func (*Predicate) Filepath() map[string]any {
	return map[string]any{
		"Exists":  filepathExists,
		"Match":   filepathMatch,
		"Abs":     filepathIsAbs,
		"Dir":     filepathIsDir,
		"File":    filepathIsFile,
		"Symlink": filepathIsSymlink,
	}
}

func (*Predicate) Path() map[string]any {
	return map[string]any{
		"Match": pathMatch,
		"Abs":   pathIsAbs,
	}
}

func (*Predicate) Regex() map[string]any {
	return map[string]any{
		"Match": regexMatch,
	}
}

func (*Predicate) String() map[string]any {
	return map[string]any{
		"EqualFold":   stringEqualFold,
		"Prefix":      stringHasPrefix,
		"Suffix":      stringHasSuffix,
		"Contains":    stringContains,
		"ContainsAny": stringContainsAny,
	}
}

func (*Predicate) Time() map[string]any {
	return map[string]any{
		"After":  timeIsAfter,
		"Before": timeIsBefore,
		"Equal":  timeIsEqual,
		"Zero":   timeIsZero,
	}
}

func (*Predicate) UUID() map[string]any {
	return map[string]any{
		"Valid":  uuidIsValid,
		"Nil":    uuidIsNil,
		"Max":    uuidIsMax,
		"V4":     uuidIsV4,
		"V7":     uuidIsV7,
		"Equal":  uuidIsEqual,
		"Before": uuidIsBefore,
		"After":  uuidIsAfter,
	}
}

func (*Predicate) IP() map[string]any {
	wrap := func(pred func(net.IP) bool) UnaryBoolPredicate {
		return func(value any) bool {
			ip := convert.ToIP(value)

			return ip != nil && pred(ip)
		}
	}

	return map[string]any{
		"V4":                 ipIsIPv4,
		"V6":                 ipIsIPv6,
		"Unspecified":        wrap(net.IP.IsUnspecified),
		"Loopback":           wrap(net.IP.IsLoopback),
		"Private":            wrap(net.IP.IsPrivate),
		"Multicast":          wrap(net.IP.IsMulticast),
		"GlobalUnicast":      wrap(net.IP.IsGlobalUnicast),
		"LinkLocalUnicast":   wrap(net.IP.IsLinkLocalUnicast),
		"LinkLocalMulticast": wrap(net.IP.IsLinkLocalMulticast),
	}
}

func (*Predicate) Math() map[string]any {
	wrap := func(op func(int) bool) func(other, value any) (bool, error) {
		return func(other, value any) (bool, error) {
			return op(compareAny(value, other)), nil
		}
	}

	return map[string]any{
		"Between":      mathBetween,
		"Equal":        wrap(func(c int) bool { return c == 0 }),
		"NotEqual":     wrap(func(c int) bool { return c != 0 }),
		"Greater":      wrap(func(c int) bool { return c > 0 }),
		"GreaterEqual": wrap(func(c int) bool { return c >= 0 }),
		"Less":         wrap(func(c int) bool { return c < 0 }),
		"LessEqual":    wrap(func(c int) bool { return c <= 0 }),
	}
}
