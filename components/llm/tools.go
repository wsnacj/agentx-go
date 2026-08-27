package llm

// Function 描述模型可调用的函数签名。
type Function struct {
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
	Strict       *bool          `json:"strict,omitempty"`
}

// Tool 表示一个可暴露给模型的工具，目前仅支持 function 类型。
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// ToolChoice 控制模型的工具选择策略。
type ToolChoice struct {
	Type     string              `json:"type"`
	Function *ToolChoiceFunction `json:"function,omitempty"`
}

// ToolChoiceFunction 在指定函数时提供名称。
type ToolChoiceFunction struct {
	Name string `json:"name"`
}

// FunctionCall 表示模型返回的函数调用指令。
type FunctionCall struct {
	ID                string `json:"id,omitempty"`
	Type              string `json:"type,omitempty"`
	Name              string `json:"name"`
	Arguments         string `json:"arguments"`
	ContinuationToken string `json:"continuation_token,omitempty"`
}

// FunctionCallDelta 描述流式函数调用的增量更新。
type FunctionCallDelta struct {
	ID                string
	Type              string
	Name              string
	Arguments         string
	ContinuationToken string
	Index             int
}

// HasName 判断 delta 中是否包含函数名。
func (d FunctionCallDelta) HasName() bool {
	return d.Name != ""
}

// HasArguments 判断 delta 是否携带参数增量。
func (d FunctionCallDelta) HasArguments() bool {
	return d.Arguments != ""
}

// FunctionResult 封装工具执行后的结果。
type FunctionResult struct {
	Call   FunctionCall
	Output string
	Err    error
}
