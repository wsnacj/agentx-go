---
name: document-operations
description: 对宿主批准的本地文档、解析结果或结果 fixture 执行字段/表格提取、证据追踪、校验与回答就绪性守卫。
user-invocable: true
tool_hints:
  - docparse_spec_select
  - docparse_profile_probe
  - docparse_extract_fields
  - docparse_extract_table
  - docparse_trace_evidence
  - docparse_validate
  - docparse_guard
allowed_tools:
  - docparse_spec_select
  - docparse_profile_probe
  - docparse_extract_fields
  - docparse_extract_table
  - docparse_trace_evidence
  - docparse_validate
  - docparse_guard
context: inline
effort: medium
---

# Document Operations

Use this skill when a user asks to extract fields or tables from a local document, trace document evidence, or validate extracted values.

The reusable module does not choose private schemas, OCR credentials, document roots, retention policy, export policy, or review queues. Use host-provided document references and specs only. For remote documents, the host must first fetch or save the document as an approved workspace artifact.

Prefer the module semantic tools:

- `docparse_spec_select` for caller-provided spec candidates.
- `docparse_profile_probe` for classify-only unknown-document proposals; it should remain review-required and must not output final fields or full table rows.
- `docparse_extract_fields` for field evidence bundles from a host parser executor or caller-provided parse result fixture.
- `docparse_extract_table` for table evidence bundles from a host parser executor or caller-provided parse result fixture.
- `docparse_trace_evidence` for page/bbox/cell/source-snippet trace.
- `docparse_validate` for expected value, period, unit, and evidence checks.
- `docparse_guard` for readiness, failure class, and review boundary.

If the tool returns `docparse_parser_executor_not_configured`, stop with a bounded answer that names the missing host parser capability instead of inventing extracted facts. Field/table extraction, validation, guard, and evidence trace may run on a caller-provided parse result or host-approved result fixture, but they do not parse real documents by themselves.
