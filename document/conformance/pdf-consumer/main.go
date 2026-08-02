package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	pdfparser "github.com/wsnacj/agentx-go/document/pdf"
)

type output struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

func run(ctx context.Context) (output, error) {
	parser, err := pdfparser.NewParser(memoryRunner{})
	if err != nil {
		return output{}, err
	}
	file, err := os.CreateTemp("", "agentx-pdf-consumer-*.pdf")
	if err != nil {
		return output{}, err
	}
	path := file.Name()
	if _, err := file.WriteString("%PDF-1.7"); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return output{}, err
	}
	_ = file.Close()
	defer os.Remove(path)

	response, err := parser.ParsePDFContext(ctx, path, false)
	if err != nil {
		return output{}, err
	}
	return output{Status: "parsed", Text: pdfparser.NewTextFormatter(response).FormatToText()}, nil
}

func main() {
	result, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(payload))
}

type memoryRunner struct{}

func (memoryRunner) Run(ctx context.Context, _ pdfparser.RunRequest) (pdfparser.RunResult, error) {
	if err := ctx.Err(); err != nil {
		return pdfparser.RunResult{}, err
	}
	return pdfparser.RunResult{Stdout: []byte(`{"message":"ok","code":0,"version":"1","result":{"pages":[{"angle":0,"height":100,"width":100,"tables":[{"type":"plain","lines":[{"text":"canonical pdf","position":[0,0,90,10]}]}]}]}}`)}, nil
}
