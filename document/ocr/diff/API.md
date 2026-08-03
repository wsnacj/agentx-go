# OCR Diff API（Experimental）

`document/ocr/diff`提供OCR文本与表格结果的确定性解析、规范化和差异计算。

## 主要入口

- `ParseOCRResponse`、`ParseTableResponse`解析既有JSON结果；
- `CompareOCRJSON`、`CompareOCRResponses`生成`DiffResult`；
- `CompareTableJSON`、`CompareTableResponses`生成`TableDiffResult`；
- `NormalizeOCRPage`、`NormalizeTablePage`和字符提取函数形成可比较序列；
- `PageDiffScore`、`PageDiffScoreOrigTable`及`FuzzyLocateText`提供纯计算评分。

本包不读取文件、不调用provider，也不修改输入。错误仅来自无效payload解析；评分是
Experimental验收机制，不应直接作为业务合规或财务正确性结论。
