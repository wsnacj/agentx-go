package expr

import (
	"encoding/json"
	"fmt"
	logger "github.com/wsnacj/agentx-go/document/pipeline/internal/logging"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	englishBillionUnitRe = regexp.MustCompile(`(?i)\bbn\b`)
	englishMillionUnitRe = regexp.MustCompile(`(?i)\bmn\b`)
	currencyUnitRe       = regexp.MustCompile(`(?i)(rmb|cny|hkd|hk\$|hk|usd|us\$|u\.s\.\$|\$|million|billion|thousand|mn|bn|yuan|renminbi|dollars?)`)
)

// 表达式求值（数字）支持 + - * / 和括号，以及函数 abs(x), approx(a,b,tol)->0/1, between(x,a,b)->0/1
func EvalNumericExpr(expr string, lookup func(id string) (float64, bool)) (float64, bool) {
	// 先展开函数调用
	e := strings.TrimSpace(expr)
	// 处理函数 approx/abs/between，递归替换为数值
	for {
		changed := false
		// approx(a,b,t)
		if s, t, argsStr, ok := findFuncCallInsensitive(e, "approx"); ok {
			a1, a2, a3 := splitArgs(argsStr)
			v1, ok1 := EvalNumericExpr(a1, lookup)
			v2, ok2 := EvalNumericExpr(a2, lookup)
			v3, ok3 := EvalNumericExpr(a3, lookup)
			if ok1 && ok2 && ok3 {
				val := 0.0
				if abs(v1-v2) <= v3 {
					val = 1
				}
				e = e[:s] + fmt.Sprintf("%g", val) + e[t:]
				changed = true
				continue
			}
		}
		// between(x,a,b)
		if s, t, argsStr, ok := findFuncCallInsensitive(e, "between"); ok {
			a1, a2, a3 := splitArgs(argsStr)
			v1, ok1 := EvalNumericExpr(a1, lookup)
			v2, ok2 := EvalNumericExpr(a2, lookup)
			v3, ok3 := EvalNumericExpr(a3, lookup)
			if ok1 && ok2 && ok3 {
				val := 0.0
				if v1 >= v2 && v1 <= v3 {
					val = 1
				}
				e = e[:s] + fmt.Sprintf("%g", val) + e[t:]
				changed = true
				continue
			}
		}
		// abs(x)
		if s, t, argsStr, ok := findFuncCallInsensitive(e, "abs"); ok {
			v, ok := EvalNumericExpr(argsStr, lookup)
			if ok {
				e = e[:s] + fmt.Sprintf("%g", abs(v)) + e[t:]
				changed = true
				continue
			}
		}
		// round(x[,n])
		if s, t, argsStr, ok := findFuncCallInsensitive(e, "round"); ok {
			logger.Debug("[docparse] round() found at position %d in: %s", s, e)
			logger.Debug("[docparse] round() argsStr='%s'", argsStr)
			a1, a2, _ := splitArgs(argsStr)
			logger.Debug("[docparse] round() args: a1='%s', a2='%s'", a1, a2)
			v1, ok1 := EvalNumericExpr(a1, lookup)
			n2 := 0.0
			ok2 := true
			if strings.TrimSpace(a2) != "" {
				n2, ok2 = EvalNumericExpr(a2, lookup)
			}
			logger.Debug("[docparse] round() values: v1=%f (ok=%t), n2=%f (ok=%t)", v1, ok1, n2, ok2)
			if ok1 && ok2 {
				n := int(math.Round(n2))
				p := math.Pow(10, float64(n))
				val := math.Round(v1*p) / p
				logger.Debug("[docparse] round() result: %f", val)
				e = e[:s] + fmt.Sprintf("%g", val) + e[t:]
				changed = true
				continue
			} else {
				logger.Debug("[docparse] round() evaluation failed")
			}
		}
		if !changed {
			break
		}
	}
	// 递归下降解析 + - * / 和括号
	p := &Parser{src: e, pos: 0, lookup: lookup}
	v := p.ParseExpr()
	if p.err != nil {
		return 0, false
	}
	return v, true
}

func EvalBooleanExpr(expr string, res *types.DocumentResult) bool {
	s := strings.TrimSpace(expr)
	if s == "" {
		return true
	}
	// 处理逻辑 && || 简单优先：先按 || 分，再按 &&
	orParts := splitTopLevel(s, "||")
	for _, orp := range orParts {
		andParts := splitTopLevel(orp, "&&")
		all := true
		for _, ap := range andParts {
			ap = strings.TrimSpace(ap)
			// 寻找比较运算符
			if ok := EvalComparison(ap, res); !ok {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

func EvalComparison(expr string, res *types.DocumentResult) bool {
	ops := []string{"<=", ">=", "==", "!=", "<", ">"}
	for _, op := range ops {
		if idx := strings.Index(expr, op); idx >= 0 {
			left := strings.TrimSpace(expr[:idx])
			right := strings.TrimSpace(expr[idx+len(op):])
			lv, lok := EvalNumericExpr(left, func(id string) (float64, bool) { return LookupGlobal(res, id) })
			rv, rok := EvalNumericExpr(right, func(id string) (float64, bool) { return LookupGlobal(res, id) })
			if !lok || !rok {
				return false
			}
			switch op {
			case "<":
				return lv < rv
			case ">":
				return lv > rv
			case "<=":
				return lv <= rv
			case ">=":
				return lv >= rv
			case "==":
				return lv == rv
			case "!=":
				return lv != rv
			}
		}
	}
	// 若没有比较运算符，尝试将表达式本身作为数字，非零视为true
	v, ok := EvalNumericExpr(expr, func(id string) (float64, bool) { return LookupGlobal(res, id) })
	return ok && v != 0
}

// LookupGlobal 解析 id 形如 chapter.field 或 globals.key，返回数值
func LookupGlobal(res *types.DocumentResult, id string) (float64, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return 0, false
	}
	parts := strings.SplitN(id, ".", 2)
	if len(parts) == 2 {
		ch, ok := res.Chapters[parts[0]]
		if !ok || ch == nil {
			logger.Debug("[docparse] Chapter not found for lookup: %s", parts[0])
			return 0, false
		}
		fr, ok := ch.Fields[parts[1]]
		if !ok {
			logger.Debug("[docparse] Field not found for lookup: %s.%s", parts[0], parts[1])
			return 0, false
		}
		switch t := fr.Value.(type) {
		case float64:
			logger.Debug("[docparse] Found value for %s: %f", id, t)
			return t, true
		case float32:
			logger.Debug("[docparse] Found value for %s: %f", id, float64(t))
			return float64(t), true
		case int:
			logger.Debug("[docparse] Found value for %s: %f", id, float64(t))
			return float64(t), true
		case int64:
			logger.Debug("[docparse] Found value for %s: %f", id, float64(t))
			return float64(t), true
		case json.Number:
			f, err := t.Float64()
			if err == nil {
				logger.Debug("[docparse] Found value for %s: %f", id, f)
				return f, true
			}
			logger.Debug("[docparse] Failed to parse json.Number for %s: %v", id, err)
			return 0, false
		case string:
			if v2, ok := NormalizeNumber(t); ok {
				if f, ok := v2.(float64); ok {
					logger.Debug("[docparse] Found normalized value for %s: %f", id, f)
					return f, true
				}
			}
			logger.Debug("[docparse] Failed to normalize string value for %s: %v", id, t)
			return 0, false
		default:
			logger.Debug("[docparse] Unsupported value type for %s: %T", id, t)
			return 0, false
		}
	}
	// 如果不是 chapter.field 格式，说明调用有误
	logger.Debug("[docparse] Invalid field identifier format: %s", id)
	return 0, false
}

func NormalizeNumber(v any) (any, bool) {
	switch t := v.(type) {
	case float64, float32, int, int64:
		return t, true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil, false
		}
		lower := strings.ToLower(s)
		// 统一括号与负号变体
		s = strings.ReplaceAll(s, "（", "(")
		s = strings.ReplaceAll(s, "）", ")")
		for _, m := range []string{"−", "–", "—", "－"} { // 常见负号
			s = strings.ReplaceAll(s, m, "-")
		}
		// 单位判断（先亿再万）
		mult := 1.0
		if strings.Contains(lower, "billion") || englishBillionUnitRe.MatchString(lower) {
			mult = 1e9
		}
		if strings.Contains(s, "亿元") || strings.HasSuffix(s, "亿") {
			mult = 1e8
		}
		if strings.Contains(lower, "million") || englishMillionUnitRe.MatchString(lower) {
			mult = 1e6
		}
		if strings.Contains(s, "万元") || strings.HasSuffix(s, "万") {
			mult = 1e4
		}
		if strings.Contains(lower, "thousand") {
			mult = 1e3
		}
		// 去常见单位/币种
		for _, u := range []string{"亿元", "万元", "万", "亿", "元", "人民币", "港币", "美元"} {
			s = strings.ReplaceAll(s, u, "")
		}
		s = currencyUnitRe.ReplaceAllString(s, "")
		// 去千分位与空格
		s = strings.ReplaceAll(s, ",", "")
		s = strings.ReplaceAll(s, " ", "")

		// 先匹配括号负数: (123.45) -> -123.45
		if m := regexp.MustCompile(`\(([+-]?\d+(?:\.\d+)?)\)`).FindStringSubmatch(s); len(m) == 2 {
			if f, err := strconv.ParseFloat(m[1], 64); err == nil {
				return -math.Abs(f) * mult, true
			}
		}
		// 否则匹配第一个数字（可带符号）
		if m := regexp.MustCompile(`[+-]?\d+(?:\.\d+)?`).FindString(s); m != "" {
			if f, err := strconv.ParseFloat(m, 64); err == nil {
				return f * mult, true
			}
		}
		return nil, false
	default:
		return nil, false
	}
}
