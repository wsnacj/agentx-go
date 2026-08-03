# Document Pipeline Result Types（Experimental）

`document/pipeline/types`定义Document Runtime的结果、候选、诊断、validation、fingerprint和
cache信息合同。

主要类型包括`DocumentResult`、`ChapterResult`、`FieldResult`、`FieldCandidate`、
`ValidationResult`、`DocumentDiagnostics`、`StageDiagnostic`、`ParseFingerprint`和
`ParseCacheInfo`。这些类型保留既有JSON字段，供pipeline、tools和Host adapter交换结果。

本包不拥有抽取算法、业务正确性或artifact存储。map/slice应视为调用方持有的snapshot；需要
并发共享时由调用方复制或同步。当前为Experimental，JSON变更必须经过focused差分。
