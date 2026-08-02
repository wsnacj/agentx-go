package pipeline

import (
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/expr"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
)

// ---- 派生与校验 ----

func runValidations(spec *configs.DocSpec, res *types.DocumentResult) []types.ValidationResult {
	var out []types.ValidationResult
	for _, vr := range spec.Validations {
		passed := expr.EvalBooleanExpr(vr.Expr, res)
		out = append(out, types.ValidationResult{Name: vr.Name, Passed: passed, Severity: vr.Severity, Message: vr.Message})
	}
	return out
}
