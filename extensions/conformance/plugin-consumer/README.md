# Plugin Contract Fixed-version Consumer

该consumer只使用固定pseudo-version、没有`replace`，验证外部module能够解析Plugin manifest、
读取依赖/权限请求并通过typed error拒绝Host-owned字段。它不安装或执行Plugin内容。
