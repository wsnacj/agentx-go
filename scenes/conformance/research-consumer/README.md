# Research fixed-version consumer

该 consumer 只依赖 `agentx-go/scenes` 固定 pseudo-version，不使用 `replace`、HS、Runner、
真实网络或 credential。它用 fixture 同时验证 `publicnews` 与 `companyresearch` 的
portable Host Kit、核心 identity 和成功路径，作为两个领域包的外部接入合同。
