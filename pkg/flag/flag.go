package flag

import (
	"bytes"
	"flag"
	"fmt"
	"slices"
	"strings"
)

type UsageOption = func(u *Usage)

func WithUsageExecutable(executable string) UsageOption {
	return func(u *Usage) {
		u.executable = strings.TrimSpace(executable)
	}
}

func WithUsageVersion(version string) UsageOption {
	return func(u *Usage) {
		u.version = strings.TrimSpace(version)
	}
}

func WithUsageDescription(parts ...string) UsageOption {
	return func(u *Usage) {
		u.description = parts
	}
}

func WithUsageNotes(parts ...string) UsageOption {
	return func(u *Usage) {
		u.notes = parts
	}
}

func WithUsageExec(exec bool) UsageOption {
	return func(u *Usage) {
		u.exec = exec
	}
}

type Usage struct {
	executable  string
	version     string
	description []string
	notes       []string
	exec        bool
}

func SetUsage(set *flag.FlagSet, options ...UsageOption) {
	var u Usage

	for _, option := range options {
		option(&u)
	}

	set.Usage = func() {
		b := new(bytes.Buffer)

		for _, part := range u.description {
			for line := range strings.SplitSeq(part, "\n") {
				fmt.Fprintf(b, "\n %s", line)
			}

			fmt.Fprint(b, "\n")
		}

		if u.executable != "" {
			fmt.Fprintf(b, "\n Usage: %s [flags]", u.executable)

			if u.exec {
				fmt.Fprintf(b, " -- <command> [arguments]")
			}

			fmt.Fprint(b, "\n")
		}

		fmt.Fprint(b, "\n Flags:\n")

		set.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(b, "\n  -%s", f.Name)

			name, usage := flag.UnquoteUsage(f)

			if name = strings.TrimSpace(name); len(name) > 0 {
				fmt.Fprintf(b, " %s", name)
			}

			if def := strings.TrimSpace(f.DefValue); def != "" {
				fmt.Fprintf(b, " ? default %q", def)
			}

			fmt.Fprint(b, "\n")

			for s := range strings.SplitSeq(usage, "\n") {
				fmt.Fprint(b, "\n     ", s)
			}

			fmt.Fprint(b, "\n")
		})

		if len(u.notes) > 0 {
			fmt.Fprint(b, "\n Notes:\n")
		}

		for _, part := range u.notes {
			for line := range strings.SplitSeq(part, "\n") {
				fmt.Fprintf(b, "\n  %s", line)
			}

			fmt.Fprint(b, "\n")
		}

		if u.version != "" {
			fmt.Fprintf(b, "\n Version: %s\n", u.version)
		}

		fmt.Fprint(b, "\n")

		b.WriteTo(set.Output()) //nolint:errcheck
	}
}

func ParseArgs(args []string) ([]string, []string) {
	if idx := slices.Index(args, "--"); idx != -1 {
		return args[:idx], args[idx+1:]
	}

	return args, nil
}
