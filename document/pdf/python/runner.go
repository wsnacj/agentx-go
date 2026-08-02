// Package python adapts an explicitly selected Python runtime to pdf.Runner.
package python

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	pdfparser "github.com/wsnacj/agentx-go/document/pdf"
)

// Config is entirely Host-owned. No field is discovered from the environment.
type Config struct {
	PythonPath  string
	ScriptPath  string
	Environment []string
	WorkingDir  string
}

// Runner launches one isolated command per request.
type Runner struct {
	config Config
}

// New validates an explicit Python adapter configuration.
func New(config Config) (*Runner, error) {
	config.PythonPath = strings.TrimSpace(config.PythonPath)
	config.ScriptPath = strings.TrimSpace(config.ScriptPath)
	config.WorkingDir = strings.TrimSpace(config.WorkingDir)
	if config.PythonPath == "" {
		return nil, fmt.Errorf("pdf python path is required")
	}
	if config.ScriptPath == "" {
		return nil, fmt.Errorf("pdf python script path is required")
	}
	if info, err := os.Stat(config.ScriptPath); err != nil || info.IsDir() {
		if err == nil {
			err = fmt.Errorf("path is a directory")
		}
		return nil, fmt.Errorf("pdf python script not found: %w", err)
	}
	config.Environment = append([]string(nil), config.Environment...)
	return &Runner{config: config}, nil
}

// BundledScriptPath returns the script shipped in the document module. Calling
// it does not select a Python interpreter or execute the script.
func BundledScriptPath() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to resolve bundled pdf script")
	}
	path := filepath.Join(filepath.Dir(file), "pdfparser.py")
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		if err == nil {
			err = fmt.Errorf("path is a directory")
		}
		return "", fmt.Errorf("bundled pdf script not found: %w", err)
	}
	return path, nil
}

// Run implements pdf.Runner.
func (r *Runner) Run(ctx context.Context, request pdfparser.RunRequest) (pdfparser.RunResult, error) {
	if r == nil {
		return pdfparser.RunResult{}, fmt.Errorf("pdf python runner is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	command := exec.CommandContext(ctx, r.config.PythonPath, r.arguments(request)...)
	command.Env = append([]string{}, r.config.Environment...)
	command.Dir = r.config.WorkingDir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	return pdfparser.RunResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
}

func (r *Runner) arguments(request pdfparser.RunRequest) []string {
	options := request.Options
	args := []string{
		r.config.ScriptPath,
		request.PDFPath,
		"--output-format", options.OutputFormat,
		"--page-range", options.PageRange,
		"--table-engine", options.TableEngine,
	}
	if options.NeedCharacter {
		args = append(args, "--need-character")
	}
	if options.ExtractImages {
		args = append(args, "--extract-images")
	}
	if options.HighAccuracyMode {
		args = append(args, "--high-accuracy")
	}
	return args
}
