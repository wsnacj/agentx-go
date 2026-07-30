package workflow

// Validator 对一个完整 Workflow Spec 执行 host admission。
//
// 本 package 只定义 construction seam，不提供默认实现。具体实现拥有自己的
// validation 规则和 error，并应原样返回 error。
type Validator interface {
	ValidateSpec(Spec) error
}
