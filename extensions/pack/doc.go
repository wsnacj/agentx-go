// Package pack 提供 substrate-neutral 的 Domain Pack定义、校验、注册、
// Workflow选择与物化、路由选择和 Binding机制。
//
// 具体 Workflow能力策略由 Host通过 Validator显式注入；memory/eval backend、
// provider、credential和真实副作用不属于本 package。
package pack
