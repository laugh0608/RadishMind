# Production Secret Backend Audit Store Storage Adapter 静态准入说明 v1

本文说明 `Production Secret Backend` 中 audit store storage adapter 相关 checker / fixture 应如何阅读。它是长期说明文档，不记录批次流水，也不替代 task card。

## 当前结论

当前 storage adapter 仍停在静态准入和入口复评阶段，最新状态锚点为：

- `audit_store_storage_adapter_runtime_implementation_entry_refresh_after_database_connection_lifecycle_defined`
- `storage_adapter_runtime_task_card_still_blocked_after_database_connection_lifecycle_entry_refresh`
- 下一项独立依赖：`storage_adapter_database_provider_connection_runtime_boundary_readiness`

已经完成的选择只到静态候选层：

- database capability family：`postgresql_compatible_append_only_relational_database`
- provider candidate class：`managed_postgresql_compatible_service`
- reference-only driver candidate：`github.com/jackc/pgx/v5`

这些结论不代表具体厂商、托管产品、provider、driver runtime、连接运行时或 storage adapter runtime 已可用。

## 证据链读法

storage adapter 的证据链按以下层次推进：

1. metadata contract artifact：固定 future adapter 的 input / result envelope、record identity、failure taxonomy 和 writer compatibility。
2. table schema artifact：固定 metadata-only logical table schema、field group、record identity、sequence reference、idempotency reference、retention / redaction reference 和 schema marker handoff。
3. backend product evidence / selection：只选择 static product class，不选择具体数据库厂商或托管产品。
4. database provider / driver / DSN / TLS / role policy：只定义 provider 边界、driver policy、secret-ref-only DSN、TLS mode、least privilege role 和 environment binding。
5. concrete database selection review：只把能力族收口到 PostgreSQL-compatible append-only relational database，不选择 vendor。
6. database provider selection readiness / review：只选择 provider candidate class，不创建 provider。
7. database driver selection readiness / review：只选择 reference-only driver candidate，不导入 driver、不 pin version。
8. database connection lifecycle readiness：只定义 DSN handoff、pool policy、timeout、retry / transaction / partial write recovery、duplicate / replay fail-closed、sanitized diagnostics 和 migration / schema marker handoff。
9. after database connection lifecycle entry refresh：确认 runtime task card 仍 blocked，并把下一步固定为 provider connection runtime boundary readiness。

## Checker 与 Fixture 语义

相关 checker 只校验 committed fixture、文档和聚合准入状态是否一致，重点包括：

- 状态锚点、decision、next dependency 没有漂移。
- 上游 dependency 已被正确消费，下游 runtime 仍被阻塞。
- `production-ops-secret-backend-implementation-readiness.json` 与 runtime blocker matrix 消费最新阻塞项。
- `scripts/check-repo.py` 注册顺序正确。
- forbidden artifact guard 不允许提前出现 runtime、provider、SQL、DDL、schema marker、migration runner 或 public API。
- no secret material scan 不允许 committed 文档、fixture 或 checker 输出敏感材料。

这些 checker 不是连接测试、driver smoke、database health check、migration dry-run 或 storage adapter runtime 测试。

## 对上层和运维的消费方式

上层项目、运维说明和只读 UI 只能消费脱敏状态：

- status anchor
- decision anchor
- next dependency
- sanitized failure code
- committed artifact path
- checker 名称和 fixture 名称

不得把 raw secret、DSN、endpoint、host、database name、credential payload、provider detail、provider raw URL、cloud credential、raw token 或 approval payload 写入文档、fixture、日志、诊断或 UI。

## 停止线

当前阶段仍禁止创建或声明以下能力：

- `go.mod` / `go.sum` 依赖变更
- driver import 或 version pin
- 真实 DSN parser
- connection provider、DB provider、connection factory
- pool runtime、health check runtime
- SQL、DDL、物理表名、列名、列类型
- schema marker runtime、migration runner
- storage adapter runtime task card、storage adapter runtime
- audit store runtime、production resolver runtime
- repository mode 或 public production API

## 相关文档

- [Audit Store Storage Adapter Evidence Rollup v1](production-secret-backend-audit-store-storage-adapter-evidence-rollup-v1.md)
- [Audit Store Runtime Blocker Matrix v1](production-secret-backend-audit-store-runtime-blocker-matrix-v1.md)
- [Database Connection Lifecycle Readiness v1](production-secret-backend-audit-store-storage-adapter-database-connection-lifecycle-readiness-v1.md)
- [After Database Connection Lifecycle Entry Refresh v1](production-secret-backend-audit-store-storage-adapter-runtime-implementation-entry-refresh-after-database-connection-lifecycle-v1.md)
- [Production Secret Reference 契约](../contracts/production-secret-reference.md)
