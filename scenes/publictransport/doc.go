// Package publictransport 提供可移植的公共交通只读查询合同、协调器、
// 确定性证据评估与 AgentX Pack。
//
// 本包不选择票务或地图服务商，不访问网络，不读取凭据，也不执行订票、购票或支付。
// 业务宿主必须显式注入 Collector，并负责限流、商业条款、授权、网络和产品免责声明。
package publictransport
