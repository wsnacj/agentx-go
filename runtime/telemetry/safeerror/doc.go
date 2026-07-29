// Package safeerror 提供用于 telemetry 与 operator observation 的安全错误投影。
//
// 本 package 不拥有业务错误分类、retry policy、HTTP status 或根 AgentX typed
// error 合同。调用方负责选择 class/code；本 package 只保证 display-safe
// message、稳定 identity、cause chain 和结构化 projection。
package safeerror
