# Weather Tool fixed-version consumer

该隔离module固定`github.com/wsnacj/agentx-go/tools` pseudo-version，不使用`replace`，也不导入HS、Runner、
Scene或Platform。它通过fake Host `httprequest.Preparer`注册并执行`weather_lookup`，证明外部Host可以在
保留网络准入权的同时复用canonical Open-Meteo协议和模型合同。

fixture不访问真实网络、Provider或凭据；该consumer是接入合同证据，不是发行、天气准确率或线上安全认证。
