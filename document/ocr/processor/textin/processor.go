package textin

import (
	"fmt"

	"github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/diff"
	"github.com/wsnacj/agentx-go/document/ocr/model"
	"github.com/wsnacj/agentx-go/document/ocr/processor"
)

// NewProcessor 构造 TextIn 专用的 ProviderProcessor。
func NewProcessor(cfg config.ProviderConfig) (processor.ProviderProcessor, error) {
	if cfg.Kind != "textin" {
		return nil, fmt.Errorf("textin processor received provider %s", cfg.Kind)
	}
	return &provider{}, nil
}

type provider struct{}

func (p *provider) For(kind model.OperationKind) (processor.OperationProcessor, error) {
	switch kind {
	case model.OperationKindOCR:
		return ocrOperation{}, nil
	case model.OperationKindTable:
		return tableOperation{}, nil
	case model.OperationKindStamp:
		return stampOperation{}, nil
	default:
		return nil, fmt.Errorf("textin processor: unsupported operation %s", kind)
	}
}

type ocrOperation struct{}

type tableOperation struct{}

type stampOperation struct{}

func (ocrOperation) Build(raw [][]byte, files []string) (any, error) {
	return buildOCRPayload(raw, files)
}

func (ocrOperation) Diff(raw [][]byte, baseline []byte, preview int) (*model.DiffSummary, error) {
	if len(baseline) == 0 || len(raw) == 0 {
		return nil, nil
	}
	current, err := mergeTextInOCRResponses(raw)
	if err != nil {
		return nil, err
	}
	base, err := mergeTextInOCRResponses([][]byte{baseline})
	if err != nil {
		return nil, err
	}
	res, err := diff.CompareOCRJSON(base, current)
	if err != nil {
		return nil, err
	}
	summary := summarizeOCRDiff(res, preview)
	if summary == nil {
		return nil, nil
	}
	return &model.DiffSummary{OCRDiff: summary}, nil
}

func (tableOperation) Build(raw [][]byte, files []string) (any, error) {
	return buildTablePayload(raw, files)
}

func (tableOperation) Diff(raw [][]byte, baseline []byte, preview int) (*model.DiffSummary, error) {
	if len(baseline) == 0 || len(raw) == 0 {
		return nil, nil
	}
	current, err := mergeTextInTableResponses(raw)
	if err != nil {
		return nil, err
	}
	base, err := mergeTextInTableResponses([][]byte{baseline})
	if err != nil {
		return nil, err
	}
	res, err := diff.CompareTableJSON(base, current)
	if err != nil {
		return nil, err
	}
	summary := summarizeTableDiff(res, preview)
	if summary == nil {
		return nil, nil
	}
	return &model.DiffSummary{TableDiff: summary}, nil
}

func (stampOperation) Build(raw [][]byte, files []string) (any, error) {
	return buildStampPayload(raw, files)
}

func (stampOperation) Diff(raw [][]byte, baseline []byte, preview int) (*model.DiffSummary, error) {
	return nil, nil
}
