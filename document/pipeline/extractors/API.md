# Document Extractors API（Experimental）

`document/pipeline/extractors`提供无provider的字段候选提取与有限后处理机制。

- `RunRegex`、`RunRegexCandidates`处理`RegexInput`并返回位置、值和confidence；
- `RunTableCandidates`从`TableInput`生成`TableResult`候选；
- `ParseLooseJSONObject`、`ParseLoosePagesMap`解析模型返回的宽松JSON片段；
- `ScriptProcess`只支持内置allowlist中的`normalize_number`与`identity`，未知名称返回
  `ok=false`。

本包不会执行任意脚本、读取文件或调用模型。regex/table输入规模和pattern安全由上层budget
控制；客户抽取规则与候选选择policy继续由Host/spec拥有。
