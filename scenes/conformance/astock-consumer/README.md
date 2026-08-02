# A-Stock fixed-version consumer

该consumer只依赖`agentx-go`固定pseudo-version，不使用`replace`、HS、Runner、真实网络或
credential。它用fixture验证A股Manifest、embedded skill、Pack选择/物化、investigation
Host Kit和evidence evaluator，并作为`scenes/astock`的外部接入合同。
