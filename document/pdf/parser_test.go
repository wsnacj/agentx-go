package pdfparser

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestParserUsesExplicitRunner(t *testing.T) {
	path := writePDFStub(t)
	runner := &fakeRunner{result: RunResult{Stdout: []byte(`{"message":"ok","code":0,"version":"1","result":{"pages":[]}}`)}}
	parser, err := NewParser(runner)
	if err != nil {
		t.Fatal(err)
	}
	response, err := parser.ParsePDFContext(context.Background(), path, true)
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != 0 || runner.request.PDFPath != path || !runner.request.Options.NeedCharacter {
		t.Fatalf("unexpected response/request: %#v %#v", response, runner.request)
	}
}

func TestParserPropagatesCancellation(t *testing.T) {
	path := writePDFStub(t)
	runner := &fakeRunner{wait: true, entered: make(chan struct{})}
	parser, err := NewParser(runner, WithTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-runner.entered
		cancel()
	}()
	_, err = parser.ParsePDFContext(ctx, path, false)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

type fakeRunner struct {
	request RunRequest
	result  RunResult
	err     error
	wait    bool
	entered chan struct{}
}

func (r *fakeRunner) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	r.request = request
	if r.wait {
		close(r.entered)
		<-ctx.Done()
		return RunResult{}, ctx.Err()
	}
	return r.result, r.err
}

func writePDFStub(t *testing.T) string {
	t.Helper()
	path := t.TempDir() + "/input.pdf"
	if err := os.WriteFile(path, []byte("%PDF-1.7"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
