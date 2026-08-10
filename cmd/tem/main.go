package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"

	_ "time/tzdata"

	"github.com/erikgeiser/ctxio"
	"github.com/urfave/cli/v3"

	"github.com/JFAexe/tem/pkg/env"
	"github.com/JFAexe/tem/pkg/template"
)

var (
	version = "custom"
	commit  = "unknown"
	date    = "unknown date"
)

var app = &cli.Command{
	Name:    "tem",
	Usage:   "tiny go template cli renderer",
	Version: fmt.Sprintf("%s (%s) built using %s on %s", version, commit, runtime.Version(), date),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "input",
			Aliases: []string{"i"},
			Sources: cli.EnvVars("TEM_INPUT_FILE"),
			Value:   "-",
			Usage:   "input file path\vreads from stdin if not specified or set to '-'\r",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Sources: cli.EnvVars("TEM_OUTPUT_FILE"),
			Value:   "-",
			Usage:   "\vout file path\vwrites to stdout if not specified or set to '-'\r",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.StringFlag{
			Name:    "delim-left",
			Aliases: []string{"l"},
			Sources: cli.EnvVars("TEM_DELIM_LEFT"),
			Value:   "{{",
			Usage:   "left template delimiter\vresets to default if set to an empty string\r",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.StringFlag{
			Name:    "delim-right",
			Aliases: []string{"r"},
			Sources: cli.EnvVars("TEM_DELIM_RIGHT"),
			Value:   "}}",
			Usage:   "right template delimiter\vresets to default if set to an empty string\r",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.StringMapFlag{
			Name:    "env",
			Aliases: []string{"e"},
			Usage:   "list of values which are accessible as envs\vformat: KEY=val\r",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.StringSliceFlag{
			Name:    "env-file",
			Aliases: []string{"f"},
			Usage:   "list of .env file paths\vonly real paths are allowed\r",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.StringSliceFlag{
			Name:    "template-file",
			Aliases: []string{"t"},
			Usage:   "list of template definition file paths\vboth paths and globs are allowed\r",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
	},
	DisableSliceFlagSeparator: true,
	Metadata: map[string]any{
		"name": os.Args[0],
		"exec": runtime.GOOS != "windows",
		"notes": []string{
			"Does not provide access to shell, or http requests",
			"Writes raw template to output and error to stderr on failure",
			"Template definitions are parsed after root template",
			"Passed envs and read .envs take precedence over process environment",
			"Read .envs are parsed after passed envs",
			"Env values are expanded on lookup",
			"Supported substitutions: `:-`, `-`, `:=`, `=`, `:+`, `+`, `:?`, `?`",
			"Glob patterns support `**`, `{groups,...}` and `[classes]`",
		},
	},
	Action: run,
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	flagger := cli.FlagStringer

	cli.FlagStringer = func(flag cli.Flag) string {
		return strings.NewReplacer(
			"\t", "\n\n\t",
			"\v", "\n\t",
			"\r ", "\n\t\n\t",
		).Replace(flagger(flag))
	}

	cli.RootCommandHelpTemplate = strings.Join([]string{
		"\n {{ .Name }} - {{ .Usage }}",
		"Usage: {{ index .Metadata `name` }} [flags] {{ if index .Metadata `exec` }}-- <command> [arguments]{{ end }}",
		"Flags:\n{{- range .VisibleFlags }}\n{{ .String | nindent 3 }}{{- end }}",
		"Notes:\n{{- range index .Metadata `notes` }}\n {{ . | nindent 3 }}{{- end }}",
		"Version: {{ .Version }}\n\n",
	}, "\n\n ")

	if err := app.Run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	var (
		args        = cmd.Args().Slice()
		inputPath   = cmd.String("input")
		outputPath  = cmd.String("output")
		delimLeft   = cmd.String("delim-left")
		delimRight  = cmd.String("delim-right")
		envs        = cmd.StringMap("env")
		envFiles    = cmd.StringSlice("env-file")
		definitions = cmd.StringSlice("template-file")

		inputFile  = os.Stdin
		outputFile = os.Stdout
	)

	if inputPath = strings.TrimSpace(inputPath); inputPath != "" && inputPath != "-" {
		abs, err := filepath.Abs(inputPath)
		if err != nil {
			return fmt.Errorf("failed to get abs path for input file: %w", err)
		}

		if inputFile, err = os.Open(abs); err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer inputFile.Close() //nolint:errcheck
	}

	rcx, cancel := context.WithCancel(ctx)
	defer cancel()

	r, err := wrapFile(rcx, inputFile)
	if err != nil {
		return fmt.Errorf("failed to create input file reader: %w", err)
	}

	raw, err := io.ReadAll(r)
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

		if outputFile, err = os.Create(abs); err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close() //nolint:errcheck
	}

	if envs, err = env.ParseMap(envs); err != nil {
		return fmt.Errorf("failed to parse raw env values: %w", err)
	}

	for _, path := range envFiles {
		dotenv, err := readDotEnvFile(ctx, path)
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
		if _, e := outputFile.Write(raw); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to write raw template to output: %w", e))
		}

		return fmt.Errorf("failed to execute template: %w", err)
	}

	if _, err = buffer.WriteTo(outputFile); err != nil {
		return fmt.Errorf("failed to write rendered template to output: %w", err)
	}

	if idx := slices.Index(args, "--"); idx != -1 {
		if err := execute(args[idx+1:]); err != nil {
			return fmt.Errorf("failed to execute process: %w", err)
		}
	}

	return nil
}

func execute(args []string) error {
	if len(args) == 0 || runtime.GOOS == "windows" {
		return nil
	}

	path, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("failed to find %#q: %w", path, err)
	}

	if err = syscall.Exec(path, args, os.Environ()); err != nil {
		return fmt.Errorf("failed to exec %#q: %w", path, err)
	}

	return nil
}

func readDotEnvFile(ctx context.Context, path string) (env.Map, error) {
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

	rcx, cancel := context.WithCancel(ctx)
	defer cancel()

	r, err := wrapFile(rcx, file)
	if err != nil {
		return nil, fmt.Errorf("failed to create env file reader: %w", err)
	}

	var envs env.Map

	if err = env.NewDecoder(r, env.WithDecoderExpand(false)).Decode(&envs); err != nil {
		return nil, fmt.Errorf("failed to parse env file: %w", err)
	}

	return envs, nil
}

func wrapFile(ctx context.Context, file *os.File) (io.ReadCloser, error) {
	if file == nil {
		return nil, fmt.Errorf("nil file")
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.Mode().IsRegular() {
		return file, nil
	}

	r, err := ctxio.WrapFile(file)
	if err != nil {
		return nil, err
	}

	go func() { <-ctx.Done(); r.Close() }() //nolint:errcheck

	return r, nil
}
