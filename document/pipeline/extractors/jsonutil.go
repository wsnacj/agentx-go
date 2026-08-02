package extractors

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// ParseLooseJSONObject 以宽容模式解析模型返回的JSON对象字符串
// 尝试顺序：
// 1) 直接解析（去掉code fence）
// 2) 从文本中定位第一个平衡的JSON对象子串再解析
// 3) 简单修复尾逗号/无意义前后缀后重试
func ParseLooseJSONObject(s string) (map[string]any, error) {
	cleaned := stripCodeFence(s)
	// 1) 直接解析
	if m, err := parseJSONMap(cleaned); err == nil {
		return m, nil
	}
	// 2) 定位第一个平衡的对象
	if sub := firstJSONObject(cleaned); sub != "" {
		if m, err := parseJSONMap(sub); err == nil {
			return m, nil
		}
	}
	// 3) 尝试修复尾逗号等常见格式错误
	fixed := fixTrailingCommas(cleaned)
	if m, err := parseJSONMap(fixed); err == nil {
		return m, nil
	}
	return nil, errors.New("invalid json object")
}

func parseJSONMap(s string) (map[string]any, error) {
	var m map[string]any
	dec := json.NewDecoder(strings.NewReader(strings.TrimSpace(s)))
	dec.UseNumber()
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("not object")
	}
	return m, nil
}

// firstJSONObject 返回第一个平衡的大括号JSON对象子串，尽量避开字符串中的括号
func firstJSONObject(s string) string {
	inStr := false
	esc := false
	depth := 0
	start := -1
	for _, r := range s {
		ch := byte(r)
		if inStr {
			if esc {
				esc = false
				continue
			}
			switch ch {
			case '\\':
				esc = true
			case '"':
				inStr = false
			}
			continue
		}
		switch ch {
		case '"':
			inStr = true
		case '{':
			if depth == 0 {
				start = len([]rune(s[:len(string([]rune(s)))])) // not reliable; simplify below
			}
			depth++
		case '}':
			if depth > 0 {
				depth--
				if depth == 0 && start >= 0 {
					// 因为上面计算start较复杂，改用索引递增扫描实现
				}
			}
		}
	}
	// 为避免上面的rune索引复杂性，改用字节扫描
	inStr = false
	esc = false
	depth = 0
	startB := -1
	for i := 0; i < len(s); i++ {
		b := s[i]
		if inStr {
			if esc {
				esc = false
				continue
			}
			if b == '\\' {
				esc = true
			} else if b == '"' {
				inStr = false
			}
			continue
		}
		switch b {
		case '"':
			inStr = true
		case '{':
			if depth == 0 {
				startB = i
			}
			depth++
		case '}':
			if depth > 0 {
				depth--
				if depth == 0 && startB >= 0 {
					return s[startB : i+1]
				}
			}
		}
	}
	return ""
}

var trailingCommaRe = regexp.MustCompile(`,\s*([}\]])`)

func fixTrailingCommas(s string) string {
	return trailingCommaRe.ReplaceAllString(s, "$1")
}

// ParseLoosePagesMap 解析 string->[]int 的映射，宽容模式
func ParseLoosePagesMap(s string) (map[string][]int, error) {
	cleaned := stripCodeFence(s)
	// 1) 尝试直接解析为 map[string][]int
	var m1 map[string][]int
	if err := json.Unmarshal([]byte(strings.TrimSpace(cleaned)), &m1); err == nil {
		return m1, nil
	}
	// 2) 截取第一个 JSON 对象重试
	if sub := firstJSONObject(cleaned); sub != "" {
		if err := json.Unmarshal([]byte(sub), &m1); err == nil {
			return m1, nil
		}
	}
	// 3) 宽容对象 + 转换数组
	obj, err := ParseLooseJSONObject(cleaned)
	if err != nil {
		return nil, err
	}
	out := map[string][]int{}
	for k, v := range obj {
		if arr, ok := toIntSlice(v); ok {
			out[k] = arr
		}
	}
	return out, nil
}

func stripCodeFence(value string) string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "```") {
		return value
	}
	firstNewline := strings.IndexByte(value, '\n')
	if firstNewline < 0 {
		return strings.Trim(value, "`")
	}
	value = value[firstNewline+1:]
	if end := strings.LastIndex(value, "```"); end >= 0 {
		value = value[:end]
	}
	return strings.TrimSpace(value)
}

func toIntSlice(v any) ([]int, bool) {
	switch t := v.(type) {
	case []any:
		var out []int
		for _, e := range t {
			if iv, ok := toInt(e); ok {
				out = append(out, iv)
			}
		}
		return out, true
	case []int:
		return t, true
	}
	return nil, false
}

func toInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return int(i), true
		}
		if f, err := t.Float64(); err == nil {
			return int(f), true
		}
	case int:
		return t, true
	case int64:
		return int(t), true
	}
	return 0, false
}
