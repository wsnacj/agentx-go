package extractors

import (
	"github.com/wsnacj/agentx-go/document/pipeline/expr"
	"strings"
)

// 简单脚本型处理器：用于归一化/后处理，先提供最常用 normalize_number

func ScriptProcess(name string, v any) (any, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "normalize_number":
		return expr.NormalizeNumber(v)
	case "identity":
		return v, true
	default:
		return v, false
	}
}
