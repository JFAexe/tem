package main

import (
	"bytes"
	stdflag "flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/JFAexe/tem/pkg/env"
	"github.com/JFAexe/tem/pkg/flag"
	"github.com/JFAexe/tem/pkg/template"
)

type Config struct {
	InputPath  string
	OutputPath string
	Envs       flag.EnvMap
	EnvFiles   flag.StringSlice
	Includes   flag.StringSlice
}

func main() {
	var c Config

	stdflag.StringVar(&c.InputPath, "i", "", "Input file path")
	stdflag.StringVar(&c.OutputPath, "o", "", "Output file path")
	stdflag.Var(&c.Envs, "e", "Extra environment variable (format: KEY_NAME=value)")
	stdflag.Var(&c.EnvFiles, "f", "Env file path")
	stdflag.Var(&c.Includes, "t", "Template include path or glob")
	stdflag.Parse()

	if err := run(&c); err != nil {
		fmt.Fprintf(os.Stderr, "failed to render template: %s\n", err)

		os.Exit(1)
	}
}

func run(cfg *Config) error {
	var (
		input  = os.Stdin
		output = os.Stdout
	)

	if cfg.InputPath != "" {
		abs, err := filepath.Abs(cfg.InputPath)
		if err != nil {
			return fmt.Errorf("failed to get abs path for input file: %w", err)
		}

		if input, err = os.Open(abs); err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer input.Close()
	}

	if cfg.OutputPath != "" {
		abs, err := filepath.Abs(cfg.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to get abs path for output file: %w", err)
		}

		if output, err = os.Create(abs); err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		defer output.Close()
	}

	if cfg.Envs == nil {
		cfg.Envs = make(flag.EnvMap)
	}

	for _, path := range cfg.EnvFiles {
		envs, err := readEnvFile(path)
		if err != nil {
			return err
		}

		maps.Copy(cfg.Envs, envs)
	}

	var (
		buf bytes.Buffer

		tpl = template.New(cfg.Envs)
	)

	if _, err := buf.ReadFrom(input); err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	if _, err := tpl.Parse(buf.String()); err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tpl.ParsePaths(cfg.Includes); err != nil {
		return fmt.Errorf("failed to parse includes: %w", err)
	}

	if err := tpl.Execute(output, make(map[string]any)); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func readEnvFile(path string) (env.Map, error) {
	if path := strings.TrimSpace(path); path == "" {
		return make(env.Map), nil
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return make(env.Map), fmt.Errorf("failed to get abs path for env file: %w", err)
	}

	file, err := os.Open(abs)
	if err != nil {
		return make(env.Map), fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	envs, err := env.Decode(file)
	if err != nil {
		return make(env.Map), fmt.Errorf("failed to parse env file: %w", err)
	}

	return envs, nil
}
