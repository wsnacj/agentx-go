// Package runstore 定义 AgentX Runtime 的 Run、NodeExecution、Event 存储合同，
// 并提供进程内 MemoryStore 与 portable node execution projection 实现。
//
// 当前成熟度为 Experimental / private validation。MemoryStore 仅用于测试、示例
// 和显式接受进程生命周期语义的宿主，不是生产持久化后端。
package runstore
