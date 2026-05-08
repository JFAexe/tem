package main

import (
	"bytes"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	_ "time/tzdata"

	"github.com/JFAexe/tem/pkg/env"
	xflag "github.com/JFAexe/tem/pkg/flag"
	"github.com/JFAexe/tem/pkg/template"
)

var (
	version = "custom"
	commit  = "unknown"
	date    = "unknown date"
)

func init() {
	xflag.SetUsage(
		flag.CommandLine,
		xflag.WithUsageExecutable(os.Args[0]),
		xflag.WithUsageExec(runtime.GOOS != "windows"),
		xflag.WithUsageVersion(fmt.Sprintf("%s (%s) built using %s on %s", version, commit, runtime.Version(), date)),
		xflag.WithUsageDescription("tem - tiny go template cli renderer"),
		xflag.WithUsageNotes(
			"Writes raw template to output and error to stderr on failure",
			"Template definitions are parsed after root template",
			"Multiple list values passed as separate flags (e.g. '-e KEY1=\"value1\" -e KEY2=\"value2\"')",
			"Passed envs and read .envs take precedence over process environment",
			"Env values are expanded on lookup, supported substitutions: `:-`, `-`, `:=`, `=`, `:+`, `+`, `:?`, `?`",
			"Glob patterns support `**`, `{groups,...}` and `[classes]`",
		),
	)
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}

func run(args []string) error {
	args, xargs := xflag.ParseArgs(args)

	if err := render(args); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	if runtime.GOOS == "windows" {
		return nil
	}

	if err := execute(xargs); err != nil {
		return fmt.Errorf("failed to execute process: %w", err)
	}

	return nil
}

func render(args []string) error {
	var (
		input       = os.Stdin
		output      = os.Stdout
		inputPath   = "-"
		outputPath  = "-"
		delimLeft   = "{{"
		delimRight  = "}}"
		envs        = make(xflag.EnvMap)
		envFiles    = make(xflag.StringSlice, 0)
		definitions = make(xflag.StringSlice, 0)
	)

	flag.StringVar(&inputPath, "i", inputPath, "Input file `path`\n\nReads from stdin if not specified or set to '-'")
	flag.StringVar(&outputPath, "o", outputPath, "Output file `path`\n\nWrites to stdout if not specified or set to '-'")
	flag.StringVar(&delimLeft, "l", delimLeft, "Left template `delimiter`\n\nResets to default if set to empty string")
	flag.StringVar(&delimRight, "r", delimRight, "Right template `delimiter`\n\nResets to default if set to empty string")
	flag.Var(&envs, "e", "List of values which are accessible as `envs`\n\nFormat: KEY_NAME=value")
	flag.Var(&envFiles, "f", "List of .env file `paths`")
	flag.Var(&definitions, "t", "List of template definition files `paths or globs`")

	if err := flag.CommandLine.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if inputPath = strings.TrimSpace(inputPath); inputPath != "" && inputPath != "-" {
		abs, err := filepath.Abs(inputPath)
		if err != nil {
			return fmt.Errorf("failed to get abs path for input file: %w", err)
		}

		if input, err = os.Open(abs); err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer input.Close() //nolint:errcheck
	}

	raw, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read root template: %w", err)
	}

	if outputPath = strings.TrimSpace(outputPath); outputPath != "" && outputPath != "-" {
		abs, err := filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("failed to get abs path for output file: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
			return fmt.Errorf("failed to create full path for output file: %w", err)
		}

		if output, err = os.Create(abs); err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close() //nolint:errcheck
	}

	for _, path := range envFiles {
		dotenv, err := readDotEnvFile(path)
		if err != nil {
			return err
		}

		maps.Copy(envs, dotenv)
	}

	if err := env.BatchSet(envs); err != nil {
		return fmt.Errorf("failed to update process environment: %w", err)
	}

	tpl := template.New(
		fmt.Sprint("root_", strings.ToLower(rand.Text())),
		template.WithDelims(delimLeft, delimRight),
	)

	if _, err := tpl.Parse(string(raw)); err != nil {
		return fmt.Errorf("failed to parse root template: %w", err)
	}

	if err := tpl.ParsePaths(definitions); err != nil {
		return fmt.Errorf("failed to parse template definitions: %w", err)
	}

	var buffer bytes.Buffer

	if err := tpl.Execute(&buffer, make(map[string]any)); err != nil {
		if _, e := output.Write(raw); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to write raw template to output: %w", e))
		}

		return fmt.Errorf("failed to execute template: %w", err)
	}

	if _, err = buffer.WriteTo(output); err != nil {
		return fmt.Errorf("failed to write rendered template to output: %w", err)
	}

	return nil
}

func execute(args []string) error {
	if len(args) == 0 {
		return nil
	}

	path, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("failed find %#q: %w", path, err)
	}

	if err = syscall.Exec(path, args, os.Environ()); err != nil {
		return fmt.Errorf("failed to exec %#q: %w", path, err)
	}

	return nil
}

func readDotEnvFile(path string) (env.Map, error) {
	if path := strings.TrimSpace(path); path == "" {
		return nil, nil
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get abs path for env file: %w", err)
	}

	file, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	var envs env.Map

	if err = env.NewDecoder(file, env.WithDecoderExpand(false)).Decode(&envs); err != nil {
		return nil, fmt.Errorf("failed to parse env file: %w", err)
	}

	return envs, nil
}
