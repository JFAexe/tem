package functions

import (
	"fmt"
	"maps"

	"github.com/JFAexe/tem/pkg/reflection"
)

var valuePredicates = map[string]any{
	"Zero":  IsZero,
	"Empty": IsEmpty,
	"Set":   IsSet,
	"In":    In,
}

var typePredicates = map[string]any{
	"Bool":   new(Type).IsBool,
	"Int":    new(Type).IsInt,
	"Uint":   new(Type).IsUint,
	"Float":  new(Type).IsFloat,
	"Number": new(Type).IsNumber,
	"String": new(Type).IsString,
	"List":   new(Type).IsList,
	"Map":    new(Type).IsMap,
	"Struct": new(Type).IsStruct,
}

var envPredicates = map[string]any{
	"Set": new(Env).IsSet,
}

var filepathPredicates = map[string]any{
	"Exists":  new(Filepath).Exists,
	"Match":   new(Filepath).Match,
	"Abs":     new(Filepath).IsAbs,
	"Dir":     new(Filepath).IsDir,
	"File":    new(Filepath).IsFile,
	"Symlink": new(Filepath).IsSymlink,
}

var pathPredicates = map[string]any{
	"Match": new(Path).Match,
	"Abs":   new(Path).IsAbs,
}

var regexPredicates = map[string]any{
	"Match": new(Regex).Match,
}

var stringPredicates = map[string]any{
	"EqualFold":   new(String).EqualFold,
	"Prefix":      new(String).HasPrefix,
	"Suffix":      new(String).HasSuffix,
	"Contains":    new(String).Contains,
	"ContainsAny": new(String).ContainsAny,
}

var timePredicates = map[string]any{
	"After":  new(Time).IsAfter,
	"Before": new(Time).IsBefore,
	"Equal":  new(Time).IsEqual,
	"Zero":   new(Time).IsZero,
}

var uuidPredicates = map[string]any{
	"Valid":  new(UUID).IsValid,
	"Nil":    new(UUID).IsNil,
	"Max":    new(UUID).IsMax,
	"V4":     new(UUID).IsV4,
	"V7":     new(UUID).IsV7,
	"Equal":  new(UUID).IsEqual,
	"Before": new(UUID).IsBefore,
	"After":  new(UUID).IsAfter,
}

var ipPredicates = map[string]any{
	"V4":                 new(IP).IsIPv4,
	"V6":                 new(IP).IsIPv6,
	"Unspecified":        new(IP).IsUnspecified,
	"Loopback":           new(IP).IsLoopback,
	"Private":            new(IP).IsPrivate,
	"Multicast":          new(IP).IsMulticast,
	"GlobalUnicast":      new(IP).IsGlobalUnicast,
	"LinkLocalUnicast":   new(IP).IsLinkLocalUnicast,
	"LinkLocalMulticast": new(IP).IsLinkLocalMulticast,
}

var mathPredicates = map[string]any{
	"Between": new(Math).Between,
}

var comparePredicates = map[string]any{
	"Equal":        wrapComparePredicate(func(c int) bool { return c == 0 }),
	"NotEqual":     wrapComparePredicate(func(c int) bool { return c != 0 }),
	"Greater":      wrapComparePredicate(func(c int) bool { return c > 0 }),
	"GreaterEqual": wrapComparePredicate(func(c int) bool { return c >= 0 }),
	"Less":         wrapComparePredicate(func(c int) bool { return c < 0 }),
	"LessEqual":    wrapComparePredicate(func(c int) bool { return c <= 0 }),
}

type Predicate struct{}

func (*Predicate) And(args ...any) (UnaryErrorPredicate, error) {
	preds, err := unaryPredicates(args)
	if err != nil {
		return nil, err
	}

	return combinePredicates(preds, true), nil
}

func (*Predicate) Or(args ...any) (UnaryErrorPredicate, error) {
	preds, err := unaryPredicates(args)
	if err != nil {
		return nil, err
	}

	return combinePredicates(preds, false), nil
}

func (*Predicate) Not(pred any) (UnaryErrorPredicate, error) {
	p, ok := unaryPredicate(pred)
	if !ok {
		return nil, fmt.Errorf("unsupported predicate type %T", pred)
	}

	return func(v any) (bool, error) {
		ok, err := p(v)
		if err != nil {
			return false, err
		}

		return !ok, nil
	}, nil
}

func (*Predicate) Bind(args ...any) (UnaryErrorPredicate, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: expected ...args, predicate", ErrValueRequired)
	}

	return bindPredicate(append(args[len(args)-1:], args[:len(args)-1]...))
}

func (*Predicate) View(args ...any) (UnaryErrorPredicate, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: expected ...path, predicate", ErrValueRequired)
	}

	last, keys := popOne(args)

	pred, ok := unaryPredicate(last)
	if !ok {
		return nil, fmt.Errorf("unsupported predicate type %T", last)
	}

	return func(item any) (bool, error) {
		current := item

		for _, key := range keys {
			v, err := reflection.Lookup(current, key)
			if err != nil {
				return false, nil
			}

			current = v.Interface()
		}

		return pred(current)
	}, nil
}

func (*Predicate) Value() any {
	return maps.Clone(valuePredicates)
}

func (*Predicate) Type() any {
	return maps.Clone(typePredicates)
}

func (*Predicate) Env() any {
	return maps.Clone(envPredicates)
}

func (*Predicate) Filepath() any {
	return maps.Clone(filepathPredicates)
}

func (*Predicate) Path() any {
	return maps.Clone(pathPredicates)
}

func (*Predicate) Regex() any {
	return maps.Clone(regexPredicates)
}

func (*Predicate) String() any {
	return maps.Clone(stringPredicates)
}

func (*Predicate) Time() any {
	return maps.Clone(timePredicates)
}

func (*Predicate) UUID() any {
	return maps.Clone(uuidPredicates)
}

func (*Predicate) IP() any {
	return maps.Clone(ipPredicates)
}

func (*Predicate) Math() any {
	return maps.Clone(mathPredicates)
}

func (*Predicate) Compare() any {
	return maps.Clone(comparePredicates)
}

func combinePredicates(preds []UnaryErrorPredicate, identity bool) UnaryErrorPredicate {
	return func(v any) (bool, error) {
		for _, pred := range preds {
			ok, err := pred(v)
			if err != nil {
				return false, err
			}

			if ok != identity {
				return ok, nil
			}
		}

		return identity, nil
	}
}

func wrapComparePredicate(op func(int) bool) BinaryErrorPredicate {
	return func(other, value any) (bool, error) {
		return op(compareAny(value, other)), nil
	}
}
