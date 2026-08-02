package processor

import (
	"fmt"

	"github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/model"
)

// OperationProcessor 抽象单个管线的解析与 diff 行为，方便按 Provider 定制实现。
type OperationProcessor interface {
	Build(raw [][]byte, files []string) (any, error)
	Diff(raw [][]byte, baseline []byte, preview int) (*model.DiffSummary, error)
}

// ProviderProcessor 负责为某个 Provider 提供各 Operation 对应的处理器。
type ProviderProcessor interface {
	For(kind model.OperationKind) (OperationProcessor, error)
}

// Factory 根据 ProviderConfig 创建 ProviderProcessor。
type Factory func(config.ProviderConfig) (ProviderProcessor, error)

// Registry 维护不同 Provider 的处理器工厂。
type Registry map[string]Factory

// Lookup 返回指定 Provider kind 对应的工厂。
func (r Registry) Lookup(kind string) (Factory, error) {
	if r == nil {
		return nil, fmt.Errorf("processor registry empty")
	}
	fac, ok := r[kind]
	if !ok {
		return nil, fmt.Errorf("processor for provider %s not registered", kind)
	}
	return fac, nil
}
