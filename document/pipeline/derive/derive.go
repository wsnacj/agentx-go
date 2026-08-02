package derive

import (
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"github.com/wsnacj/agentx-go/document/pipeline/expr"
	logger "github.com/wsnacj/agentx-go/document/pipeline/internal/logging"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"github.com/wsnacj/agentx-go/document/pipeline/utils"
	"sort"
	"strings"
)

// EvaluateDerived 执行派生字段计算：分阶段求值 + 依赖图 + 循环检测
func EvaluateDerived(spec *configs.DocSpec, result *types.DocumentResult) {
	// 1) 构建派生项清单与初始依赖
	type dspec struct {
		id      string
		chapter string
		field   string
		formula string
		deps    []string // 绝对依赖 chapter.field
	}
	derived := map[string]dspec{}
	for _, ch := range spec.Chapters {
		for _, f := range ch.Fields {
			frm := strings.TrimSpace(f.DerivedFormula)
			if frm == "" {
				continue
			}
			id := ch.Key + "." + f.Key
			deps := extractIdentifiersAsGlobal(frm, ch.Key)
			derived[id] = dspec{id: id, chapter: ch.Key, field: f.Key, formula: frm, deps: deps}
			logger.Debug("[docparse] Derived spec: %s depends on %v", id, deps)
		}
	}
	// 2) 分阶段求值：每轮计算依赖已就绪的派生项，直到收敛
	pending := map[string]dspec{}
	for k, v := range derived {
		pending[k] = v
	}
	for {
		progressed := false
		for id, it := range pending {
			// 检查依赖是否就绪（依赖不是 pending 或者是已有基础字段）
			ready := true
			for _, dep := range it.deps {
				if _, isDerived := pending[dep]; isDerived {
					ready = false
					break
				}
				if _, existsInAll := derived[dep]; existsInAll {
					// 依赖是派生，但已在上一轮计算完毕（不在 pending 中），视为可用
					continue
				}
				// 尝试在当前结果中查找基础字段
				if _, ok := expr.LookupGlobal(result, dep); !ok {
					ready = false
					break
				}
			}
			if !ready {
				continue
			}
			// 计算
			v, ok := expr.EvalNumericExpr(it.formula, func(id2 string) (float64, bool) {
				return lookupGlobalWithContext(result, it.chapter, id2)
			})
			if !ok {
				// 留待之后（有时函数内嵌依赖无法静态识别）
				logger.Debug("[docparse] Deferred derived (not ready at runtime): %s = %s", id, it.formula)
				continue
			}
			// 写入结果
			cres := result.Chapters[it.chapter]
			if cres == nil {
				cres = &types.ChapterResult{Key: it.chapter, Fields: map[string]types.FieldResult{}}
				result.Chapters[it.chapter] = cres
				result.ChapterOrder = append(result.ChapterOrder, it.chapter)
			}
			if cres.Fields == nil {
				cres.Fields = map[string]types.FieldResult{}
			}
			cres.Fields[it.field] = types.FieldResult{Key: it.field, Value: v, Source: "derived"}
			logger.Debug("[docparse] Calculated derived field: %s = %f", id, v)
			delete(pending, id)
			progressed = true
		}
		if !progressed {
			break
		}
	}

	// 3) 构建未就绪诊断与循环检测
	if len(pending) > 0 {
		// 依赖图（仅派生→派生的边）
		indeg := map[string]int{}
		adj := map[string][]string{}
		pendingSet := map[string]struct{}{}
		for id := range pending {
			pendingSet[id] = struct{}{}
		}
		for id, it := range pending {
			for _, dep := range it.deps {
				if _, isPending := pendingSet[dep]; isPending {
					indeg[id]++
					adj[dep] = append(adj[dep], id)
				}
			}
		}
		// Kahn 拓扑：找出循环中的节点
		q := []string{}
		for id := range pending {
			if indeg[id] == 0 {
				q = append(q, id)
			}
		}
		visited := map[string]bool{}
		for len(q) > 0 {
			x := q[0]
			q = q[1:]
			visited[x] = true
			for _, y := range adj[x] {
				indeg[y]--
				if indeg[y] == 0 {
					q = append(q, y)
				}
			}
		}
		inCycle := map[string]bool{}
		for id := range pending {
			if !visited[id] && indeg[id] > 0 {
				inCycle[id] = true
			}
		}

		// 生成诊断
		var diags []types.DerivedDiagnostic
		for id, it := range pending {
			var missing []string
			var blocked []string
			for _, dep := range it.deps {
				if _, isPending := pendingSet[dep]; isPending {
					blocked = append(blocked, dep)
					continue
				}
				if _, isDerivedAll := derived[dep]; isDerivedAll {
					// 已计算完成的派生，不算缺失
					continue
				}
				if _, ok := expr.LookupGlobal(result, dep); !ok {
					missing = append(missing, dep)
				}
			}
			diags = append(diags, types.DerivedDiagnostic{
				ID:          id,
				Chapter:     it.chapter,
				Field:       it.field,
				Formula:     it.formula,
				MissingDeps: utils.UniqueKeepOrder(missing),
				BlockedBy:   utils.UniqueKeepOrder(blocked),
				Cycle:       inCycle[id],
			})
			logger.Warn("[docparse] Derived unresolved: %s, missing=%v, blocked=%v, cycle=%t", id, missing, blocked, inCycle[id])
		}
		sort.Slice(diags, func(i, j int) bool {
			return diags[i].ID < diags[j].ID
		})
		result.DerivedDiagnostics = diags
	}
}

// ---- 派生依赖分析 ----

// 提取表达式中的标识符列表（不包含数字与函数名），并将未带章名前缀的标识符提升为 currentChapter.id
func extractIdentifiersAsGlobal(ex, currentChapter string) []string {
	toks := tokenizeExpr(ex)
	var out []string
	for _, t := range toks {
		lt := strings.ToLower(strings.TrimSpace(t))
		if lt == "" {
			continue
		}
		// 忽略函数名
		switch lt {
		case "abs", "approx", "between", "round":
			continue
		}
		if _, ok := expr.ParseNumber(t); ok {
			continue
		}
		// 提升为 chapter.field
		if strings.Contains(t, ".") {
			out = append(out, t)
		} else {
			out = append(out, currentChapter+"."+t)
		}
	}
	return utils.UniqueKeepOrder(out)
}

// 将表达式按运算符与空白分割为 token，尽量保留中文/字母/数字/下划线/点的连续片段
func tokenizeExpr(s string) []string {
	var out []string
	var buf []rune
	flush := func() {
		if len(buf) > 0 {
			out = append(out, string(buf))
			buf = buf[:0]
		}
	}
	for _, r := range s {
		switch r {
		case '+', '-', '*', '/', '(', ')', ',', ' ', '\t':
			flush()
		default:
			buf = append(buf, r)
		}
	}
	flush()
	return out
}

// lookupGlobalWithContext 带上下文的字段查找，支持单字段名在当前章节查找
func lookupGlobalWithContext(res *types.DocumentResult, currentChapter, id string) (float64, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return 0, false
	}
	parts := strings.SplitN(id, ".", 2)
	if len(parts) == 2 {
		// chapter.field 格式
		return expr.LookupGlobal(res, id)
	} else {
		// 单字段名，在当前章节查找
		return expr.LookupGlobal(res, currentChapter+"."+id)
	}
}
