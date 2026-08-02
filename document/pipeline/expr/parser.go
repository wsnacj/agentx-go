package expr

import (
	"fmt"
	"strconv"
	"strings"
)

// 简单语法工具
type Parser struct {
	src    string
	pos    int
	err    error
	lookup func(id string) (float64, bool)
}

func (p *Parser) ParseExpr() float64 { // expr = term { (+|-) term }
	v := p.ParseTerm()
	for p.err == nil {
		p.skipSpaces()
		if p.match('+') {
			v += p.ParseTerm()
		} else if p.match('-') {
			v -= p.ParseTerm()
		} else {
			break
		}
	}
	return v
}
func (p *Parser) ParseTerm() float64 { // term = factor { (*|/) factor }
	v := p.ParseFactor()
	for p.err == nil {
		p.skipSpaces()
		if p.match('*') {
			v *= p.ParseFactor()
		} else if p.match('/') {
			denom := p.ParseFactor()
			if denom == 0 {
				p.err = fmt.Errorf("div0")
				return 0
			}
			v /= denom
		} else {
			break
		}
	}
	return v
}
func (p *Parser) ParseFactor() float64 {
	p.skipSpaces()
	// unary sign support
	sign := 1.0
	if p.match('+') {
		// keep sign = +1
	} else if p.match('-') {
		sign = -1
	}
	p.skipSpaces()
	if p.match('(') {
		v := p.ParseExpr()
		p.skipSpaces()
		if !p.match(')') {
			p.err = fmt.Errorf(") expected")
			return 0
		}
		return sign * v
	}
	// number or identifier
	tok := p.readToken()
	if tok == "" {
		p.err = fmt.Errorf("token")
		return 0
	}
	if f, ok := ParseNumber(tok); ok {
		return sign * f
	}
	if v, ok := p.lookup(tok); ok {
		return sign * v
	}
	p.err = fmt.Errorf("unknown id %s", tok)
	return 0
}
func (p *Parser) skipSpaces() {
	for p.pos < len(p.src) {
		if p.src[p.pos] == ' ' || p.src[p.pos] == '\t' {
			p.pos++
		} else {
			break
		}
	}
}
func (p *Parser) match(c byte) bool {
	if p.pos < len(p.src) && p.src[p.pos] == c {
		p.pos++
		return true
	}
	return false
}
func (p *Parser) readToken() string {
	start := p.pos
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '(' || ch == ')' || ch == ',' || ch == ' ' || ch == '\t' {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(p.src[start:p.pos])
}

func ParseNumber(s string) (float64, bool) {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, true
	}
	return 0, false
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func splitArgs(s string) (string, string, string) {
	parts := splitTopLevel(s, ",")
	a1, a2, a3 := "", "", ""
	if len(parts) > 0 {
		a1 = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 {
		a2 = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		a3 = strings.TrimSpace(parts[2])
	}
	return a1, a2, a3
}
func splitTopLevel(s, sep string) []string {
	var out []string
	depth := 0
	last := 0
	for i := 0; i+len(sep) <= len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		}
		if depth == 0 && strings.HasPrefix(s[i:], sep) {
			out = append(out, s[last:i])
			i += len(sep) - 1
			last = i + 1
		}
	}
	out = append(out, s[last:])
	return out
}

// findFuncCallInsensitive 在原始字符串 s 中查找函数调用 name(…)
// - 函数名大小写不敏感
// - 允许 name 与 '(' 之间存在空格或制表符
// - 要求函数名前一字符不是标识符字符（字母/数字/下划线/点），避免命中 background( / my_round(
// 返回：调用起始位置 start，调用结束位置 end（紧随右括号之后），括号内参数串 args（去两端空格），以及是否找到 ok
func findFuncCallInsensitive(s, name string) (start int, end int, args string, ok bool) {
	n := len(name)
	if n == 0 || len(s) < n+2 { // 至少需要 name + ()
		return -1, -1, "", false
	}
	for i := 0; i+n <= len(s); i++ {
		// 函数名大小写不敏感匹配
		if !strings.EqualFold(s[i:i+n], name) {
			continue
		}
		// 边界：前一字符不能是标识符字符
		if i > 0 {
			if isIdentByte(s[i-1]) {
				continue
			}
		}
		// 跳过 name 后的空白，必须跟随 '('
		j := i + n
		for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
			j++
		}
		if j >= len(s) || s[j] != '(' {
			continue
		}
		// 从 '(' 后寻找匹配的右括号，处理嵌套
		open := j
		depth := 0
		k := j + 1
		for k < len(s) {
			switch s[k] {
			case '(':
				depth++
			case ')':
				if depth == 0 {
					// 命中
					return i, k + 1, strings.TrimSpace(s[open+1 : k]), true
				}
				depth--
			}
			k++
		}
		// 未找到收尾，继续扫描后续位置
	}
	return -1, -1, "", false
}

func isIdentByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_' || b == '.'
}
